package main

import (
	"context"
	"flag"
	"fmt"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type ServerParameters struct {
	port     int    // webhook server port
	certFile string // path to the x509 certificate for https
	keyFile  string // path to the x509 private key matching `CertFile`
}

var parameters ServerParameters
var wafClient *golangsdk.ServiceClient

var (
	universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
)

func main() {
	flag.IntVar(&parameters.port, "port", 8443, "Webhook server port.")
	flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	kubeConfig := getKubeConfig()
	kubeClientSet := getClientSet(kubeConfig)

	createWafServiceClient(kubeClientSet)

	http.HandleFunc("/", HandleRoot)
	http.HandleFunc("/upload-cert-to-waf", HandleUploadCertToWaf)
	log.Fatal(http.ListenAndServeTLS(":"+strconv.Itoa(parameters.port), parameters.certFile, parameters.keyFile, nil))
}

func getClientSet(config *rest.Config) *kubernetes.Clientset {
	var clientSet *kubernetes.Clientset
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	clientSet = cs
	return clientSet
}

func getKubeConfig() *rest.Config {
	var config *rest.Config
	useKubeConfig := os.Getenv("USE_KUBECONFIG")
	kubeConfigFilePath := os.Getenv("KUBECONFIG")
	if len(useKubeConfig) == 0 {
		// default to service account in cluster token
		c, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		config = c
	} else {
		//load from a kube config
		var kubeconfig string

		if kubeConfigFilePath == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
		} else {
			kubeconfig = kubeConfigFilePath
		}

		fmt.Println("kubeconfig: " + kubeconfig)

		c, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		config = c
	}
	return config
}

func createWafServiceClient(clientSet *kubernetes.Clientset) *golangsdk.ServiceClient {
	secret, err := clientSet.CoreV1().Secrets("default").Get(context.TODO(), "otc-credentials", metav1.GetOptions{})

	if err != nil {
		panic(err.Error())
	}

	authOpts := golangsdk.AuthOptions{
		IdentityEndpoint: "https://iam.eu-de.otc.t-systems.com:443/v3",
		Username:         string(secret.Data["username"]),
		Password:         string(secret.Data["password"]),
		DomainID:         string(secret.Data["domainID"]),
		TenantID:         string(secret.Data["tenantID"]),
		AllowReauth:      true,
	}

	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		log.Fatal("error creating otc client", err)
	}

	log.Println("new otc client created successfully!")

	opts := golangsdk.EndpointOpts{Region: "eu-de"}
	wafClient, err = openstack.NewWAFV1(provider, opts)

	if err != nil {
		log.Fatal("error creating waf service client", err)
	}
	log.Println("new waf client created successfully!")
	return wafClient
}
