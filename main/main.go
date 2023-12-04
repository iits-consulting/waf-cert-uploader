package main

import (
	"context"
	"flag"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	apiv1 "k8s.io/api/core/v1"
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
	flagWebhookParameters()

	kubeConfig := getKubeConfig()
	kubeClientSet := getClientSet(kubeConfig)

	setupOtcClient(kubeClientSet)

	http.HandleFunc("/upload-cert-to-waf", HandleUploadCertToWaf)
	addr := ":" + strconv.Itoa(parameters.port)
	log.Fatal(http.ListenAndServeTLS(addr, parameters.certFile, parameters.keyFile, nil))
}

func getClientSet(config *rest.Config) *kubernetes.Clientset {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("error creating new client set from rest config", err)
	}
	return clientSet
}

func getKubeConfig() *rest.Config {
	useKubeConfig := os.Getenv("USE_KUBECONFIG")
	kubeConfigFilePath := os.Getenv("KUBECONFIG")
	if len(useKubeConfig) == 0 {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			log.Fatal("error getting the in-cluster-config", err)
		}
		return inClusterConfig
	} else {
		return getLocalKubeConfig(kubeConfigFilePath)
	}
}

func getLocalKubeConfig(kubeConfigFilePath string) *rest.Config {
	var kubeconfig string

	if kubeConfigFilePath == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	} else {
		kubeconfig = kubeConfigFilePath
	}

	localKubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal("error getting local kube-config", err)
	}
	return localKubeConfig
}

func setupOtcClient(clientSet *kubernetes.Clientset) {
	secret := getOtcCredentials(clientSet)
	authOpts := getAuthOptions(secret)
	provider := createProviderClient(authOpts)
	createWafServiceClient(provider)
}

func createWafServiceClient(provider *golangsdk.ProviderClient) {
	opts := golangsdk.EndpointOpts{Region: "eu-de"}
	var err error
	wafClient, err = openstack.NewWAFV1(provider, opts)

	if err != nil {
		log.Fatal("error creating waf service client", err)
	}
	log.Println("new waf client created successfully!")
}

func createProviderClient(authOpts golangsdk.AuthOptions) *golangsdk.ProviderClient {
	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		log.Fatal("error creating otc client", err)
	}

	log.Println("new otc client created successfully!")
	return provider
}

func getAuthOptions(secret *apiv1.Secret) golangsdk.AuthOptions {
	return golangsdk.AuthOptions{
		IdentityEndpoint: "https://iam.eu-de.otc.t-systems.com:443/v3",
		Username:         string(secret.Data["username"]),
		Password:         string(secret.Data["password"]),
		DomainID:         string(secret.Data["domainID"]),
		TenantID:         string(secret.Data["tenantID"]),
		AllowReauth:      true,
	}
}

func getOtcCredentials(clientSet *kubernetes.Clientset) *apiv1.Secret {
	secret, err := clientSet.CoreV1().
		Secrets("default").
		Get(context.Background(), "otc-credentials", metav1.GetOptions{})

	if err != nil {
		log.Fatal("error getting kubernetes secrets", err)
	}
	return secret
}

func flagWebhookParameters() {
	flag.IntVar(&parameters.port, "port", 8443, "Webhook server port.")
	flag.StringVar(
		&parameters.certFile,
		"tlsCertFile",
		"/etc/webhook/certs/tls.crt",
		"File containing the x509 Certificate for HTTPS.")
	flag.StringVar(
		&parameters.keyFile,
		"tlsKeyFile",
		"/etc/webhook/certs/tls.key",
		"File containing the x509 private key to --tlsCertFile.")
	flag.Parse()
}
