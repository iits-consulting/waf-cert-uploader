package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"waf-cert-uploader/controller"
	"waf-cert-uploader/service"
)

type ServerParameters struct {
	certFile string // path to the x509 certificate for https
	keyFile  string // path to the x509 private key matching `CertFile`
}

var parameters ServerParameters

func main() {
	err := flagWebhookParameters()
	if err != nil {
		log.Println(err)
		return
	}

	err = service.SetupOtcClient()
	if err != nil {
		log.Println("otc client setup failed", err)
		return
	}

	registerHttpControllers()
	setupHttpServers()
}

func registerHttpControllers() {
	http.HandleFunc("/health", controller.HandleHealth)
	http.HandleFunc("/upload-cert-to-waf", controller.HandleUploadCertToWaf)
}

func setupHttpServers() {
	httpsPort, err := lookupPort()
	if err != nil {
		return
	}
	err = http.ListenAndServeTLS(":"+*httpsPort, parameters.certFile, parameters.keyFile, nil)
	if err != nil {
		log.Println("https server failed: ", err)
	}
}

func lookupPort() (*string, error) {
	httpsPort, httpsFound := os.LookupEnv("PORT")
	if !httpsFound {
		return nil, errors.New("environment variable for https port is not set")
	}
	return &httpsPort, nil
}

func flagWebhookParameters() error {
	certMountPath, foundMountPath := os.LookupEnv("CERT_MOUNT_PATH")
	if !foundMountPath {
		return errors.New("mount path environment variable was not set")
	}
	flag.StringVar(
		&parameters.certFile,
		"tlsCertFile",
		certMountPath+"tls.crt",
		"File containing the x509 Certificate for HTTPS.")
	flag.StringVar(
		&parameters.keyFile,
		"tlsKeyFile",
		certMountPath+"tls.key",
		"File containing the x509 private key to --tlsCertFile.")
	flag.Parse()
	return nil
}
