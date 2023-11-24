package main

import (
	"fmt"
	waf "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/certificates"
	apiv1 "k8s.io/api/core/v1"
	"log"
	"regexp"
	"time"
)

type CertificateSecret struct {
	tlsCert   string
	tlsKey    string
	certWafId string
}

func CreateOrUpdateCertificate(secret apiv1.Secret) string {
	certSecret := getSecret(secret)

	if len(certSecret.certWafId) == 0 {
		return uploadNewCertificate(secret, certSecret)
	}

	return updateExistingCertificate(secret, certSecret)
}

func uploadNewCertificate(secret apiv1.Secret, certSecret CertificateSecret) string {
	log.Println("uploading a new certificate to web application firewall...")
	log.Println("certificate domain name: " + secret.Annotations["cert-manager.io/certificate-name"])

	dt := time.Now()
	createOpts := waf.CreateOpts{
		Name:    fmt.Sprintf("tls-cert-%s", dt.Format("01-02-2006-15-04-05")),
		Content: certSecret.tlsCert,
		Key:     certSecret.tlsKey,
	}

	createResult := waf.Create(wafClient, createOpts)
	certificate, err := createResult.Extract()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("created a new certificate in waf with id " + certificate.Id)
	return certificate.Id
}

func updateExistingCertificate(secret apiv1.Secret, certSecret CertificateSecret) string {
	log.Println("trying to delete expired certificate with id " + certSecret.certWafId)
	deleteResult := waf.Delete(wafClient, certSecret.certWafId)

	extracted, err := deleteResult.Extract()

	if err != nil {
		log.Println(extracted.ErrorMsg)
		log.Fatal(err)
	}
	log.Println("certificate deleted successfully")

	return uploadNewCertificate(secret, certSecret)
}

func getSecret(secret apiv1.Secret) CertificateSecret {
	tlsCertificate := string(secret.Data["tls.crt"])
	tlsKey := string(secret.Data["tls.key"])
	certWafId := string(secret.Data["cert-waf-id"])

	regex := regexp.MustCompile(`\r?\n`)
	tlsCertificate = regex.ReplaceAllString(tlsCertificate, "")
	tlsKey = regex.ReplaceAllString(tlsKey, "")

	return CertificateSecret{
		tlsCert:   tlsCertificate,
		tlsKey:    tlsKey,
		certWafId: certWafId,
	}
}
