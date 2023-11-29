package main

import (
	"fmt"
	waf "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/certificates"
	apiv1 "k8s.io/api/core/v1"
	"log"
	"regexp"
)

type CertificateSecret struct {
	tlsCert   string
	tlsKey    string
	certWafId string
}

func CreateOrUpdateCertificate(secret apiv1.Secret) string {
	certSecret := getSecret(secret)
	var certId string
	if len(certSecret.certWafId) == 0 {
		log.Println("no certificate has been uploaded to the waf yet...")
		certId = uploadNewCertificate(secret, certSecret)
	} else {
		log.Println("a certificate was already created, replacing current certificate by a new one...")
		certId = updateExistingCertificate(secret, certSecret)
	}
	if certId == "0" {
		return findCertIdRelatedToCurrentResourceVersion(secret)
	}
	return certId
}

func findCertIdRelatedToCurrentResourceVersion(secret apiv1.Secret) string {
	log.Println("find certificate id related to current resource version " + secret.ResourceVersion + "...")
	certList, err := waf.List(wafClient, waf.ListOpts{}).AllPages()
	if err != nil {
		log.Fatal(err)
	}
	extracted, err := waf.ExtractCertificates(certList)
	if err != nil {
		log.Fatal(err)
	}
	for _, certificate := range extracted {
		if certificate.Name == fmt.Sprintf("tls-cert-%s", secret.ResourceVersion) {
			log.Println("found a certificate! Id:" + certificate.Id)
			return certificate.Id
		}
	}
	log.Println("no previous certificate was found in the waf")
	return "0"
}

func uploadNewCertificate(secret apiv1.Secret, certSecret CertificateSecret) string {
	log.Println("uploading a new certificate to web application firewall...")
	log.Println("certificate domain name: " + secret.Annotations["cert-manager.io/certificate-name"])

	createOpts := waf.CreateOpts{
		Name:    fmt.Sprintf("tls-cert-%s", secret.ResourceVersion),
		Content: certSecret.tlsCert,
		Key:     certSecret.tlsKey,
	}

	createResult := waf.Create(wafClient, createOpts)
	certificate, err := createResult.Extract()
	if err != nil {
		log.Println(err)
		return "0"
	}

	log.Println("created a new certificate in waf with id " + certificate.Id)
	return certificate.Id
}

func updateExistingCertificate(secret apiv1.Secret, certSecret CertificateSecret) string {
	log.Println("trying to delete expired certificate with id " + certSecret.certWafId)
	deleteResult := waf.Delete(wafClient, certSecret.certWafId)

	_, err := deleteResult.Extract()

	if err != nil {
		// we can't panic here because the endpoint is non-blocking and needs to be idempotent.
		// scenario: two admission reviews a1, a2. a1 has deleted the certificate in waf but secret not mutated yet
		log.Println("certificate couldn't be deleted, possible reason: deletion by a previous admission review")
	} else {
		log.Println("certificate deleted successfully")
		return uploadNewCertificate(secret, certSecret)
	}
	return "0"
}

func getSecret(secret apiv1.Secret) CertificateSecret {
	tlsCertificate := string(secret.Data["tls.crt"])
	tlsKey := string(secret.Data["tls.key"])
	certWafId := secret.Labels["cert-waf-id"]

	regex := regexp.MustCompile(`\r?\n`)
	tlsCertificate = regex.ReplaceAllString(tlsCertificate, "")
	tlsKey = regex.ReplaceAllString(tlsKey, "")

	return CertificateSecret{
		tlsCert:   tlsCertificate,
		tlsKey:    tlsKey,
		certWafId: certWafId,
	}
}
