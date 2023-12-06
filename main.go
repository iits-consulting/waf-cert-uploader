package main

import (
	"flag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"waf-webhook/controller"
	"waf-webhook/service"
)

type ServerParameters struct {
	port     int    // webhook server port
	certFile string // path to the x509 certificate for https
	keyFile  string // path to the x509 private key matching `CertFile`
}

var parameters ServerParameters

func main() {
	flagWebhookParameters()

	kubeConfig := getKubeConfig()
	kubeClientSet := getClientSet(kubeConfig)

	service.SetupOtcClient(kubeClientSet)

	http.HandleFunc("/upload-cert-to-waf", controller.HandleUploadCertToWaf)
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
