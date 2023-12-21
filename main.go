package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"waf-cert-uploader/controller"
	"waf-cert-uploader/service"
)

type ServerParameters struct {
	port     int    // webhook server port
	certFile string // path to the x509 certificate for https
	keyFile  string // path to the x509 private key matching `CertFile`
}

var parameters ServerParameters

func main() {
	flagWebhookParameters()

	err := service.SetupOtcClient()
	if err != nil {
		log.Println("otc client setup failed", err)
		return
	}

	httpPort := os.Getenv("HTTP_PORT")
	httpsPort := os.Getenv("HTTPS_PORT")

	http.HandleFunc("/health", controller.HandleHealth)
	http.HandleFunc("/upload-cert-to-waf", controller.HandleUploadCertToWaf)
	go func() {
		err = http.ListenAndServe(":"+httpPort, nil)
		if err != nil {
			log.Println("http server failed: ", err)
		}
	}()
	err = http.ListenAndServeTLS(":"+httpsPort, parameters.certFile, parameters.keyFile, nil)
	if err != nil {
		log.Println("https server failed: ", err)
	}
}

func flagWebhookParameters() {
	certMountPath := os.Getenv("CERT_MOUNT_PATH")
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
}
