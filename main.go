package main

import (
	"flag"
	"fmt"
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
	"waf-webhook/internal"
)

// var config *rest.Config
var clientSet *kubernetes.Clientset

type ServerParameters struct {
	port     int    // webhook server port
	certFile string // path to the x509 certificate for https
	keyFile  string // path to the x509 private key matching `CertFile`
}

var parameters ServerParameters

var (
	universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
)

func main() {
	logger := internal.NewLogger("INFO")
	createOtcClient(logger)

	flag.IntVar(&parameters.port, "port", 8443, "Webhook server port.")
	flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	kubeConfig := getKubeConfig()
	kubeClientSet := getClientSet(kubeConfig)

	test(kubeClientSet)
	http.HandleFunc("/", HandleRoot)
	http.HandleFunc("/mutate", HandleMutate)
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

func createOtcClient(logger internal.ILogger) *internal.OtcWrapper {
	config := internal.ConfigStruct{
		AuthenticationData: internal.AuthenticationData{
			Username:             "",
			Password:             "",
			AccessKey:            "",
			SecretKey:            "",
			IsAkSkAuthentication: false,
			ProjectId:            "",
			DomainName:           "",
			Region:               "",
		},
		Namespaces:                nil,
		Port:                      0,
		WaitDuration:              0,
		ResourceIdNameMappingFlag: false,
	}
	client, err := internal.NewOtcClientFromConfig(config, logger)
	if err != nil {
		logger.Panic("Error creating OTC client", "error", err)
	}

	logger.Info("New OTC Client created!")
	return client
}
