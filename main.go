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

	kubeConfig, err := getKubeConfig()
	if err != nil {
		return
	}
	kubeClientSet, err := getClientSet(kubeConfig)
	if err != nil {
		return
	}

	err = service.SetupOtcClient(kubeClientSet)
	if err != nil {
		log.Println("otc client setup failed", err)
		return
	}

	http.HandleFunc("/upload-cert-to-waf", controller.HandleUploadCertToWaf)
	addr := ":" + strconv.Itoa(parameters.port)
	err = http.ListenAndServeTLS(addr, parameters.certFile, parameters.keyFile, nil)
	if err != nil {
		log.Println("https server failed: ", err)
	}
}

func getClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("error creating new client set from rest config", err)
		return nil, err
	}
	return clientSet, nil
}

func getKubeConfig() (*rest.Config, error) {
	useKubeConfig := os.Getenv("USE_KUBECONFIG")
	kubeConfigFilePath := os.Getenv("KUBECONFIG")
	if len(useKubeConfig) == 0 {
		inClusterConfig, err := rest.InClusterConfig()
		if err != nil {
			log.Println("error getting the in-cluster-config", err)
			return nil, err
		}
		return inClusterConfig, nil
	} else {
		return getLocalKubeConfig(kubeConfigFilePath)
	}
}

func getLocalKubeConfig(kubeConfigFilePath string) (*rest.Config, error) {
	var kubeConfig string

	if kubeConfigFilePath == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeConfig = filepath.Join(home, ".kube", "config")
		}
	} else {
		kubeConfig = kubeConfigFilePath
	}

	localKubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		log.Println("error getting local kube-config", err)
		return nil, err
	}
	return localKubeConfig, nil
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
