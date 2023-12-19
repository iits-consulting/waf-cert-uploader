package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
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

	http.HandleFunc("/upload-cert-to-waf", controller.HandleUploadCertToWaf)
	addr := ":" + strconv.Itoa(parameters.port)
	err = http.ListenAndServeTLS(addr, parameters.certFile, parameters.keyFile, nil)
	if err != nil {
		log.Println("https server failed: ", err)
	}
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
