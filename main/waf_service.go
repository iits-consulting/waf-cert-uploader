package main

import (
	"crypto/sha256"
	"encoding/hex"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	waf "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/certificates"
	wafDomain "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/domains"
	"github.com/thoas/go-funk"
	apiv1 "k8s.io/api/core/v1"
	"log"
	"regexp"
)

type CertificateSecret struct {
	certName    string
	tlsCert     string
	tlsKey      string
	domainName  string
	wafDomainId string
	certWafId   string
}

var createAndExtract = func(c *golangsdk.ServiceClient, opts waf.CreateOpts) (*waf.Certificate, error) {
	return waf.Create(c, opts).Extract()
}

var deleteAndExtract = func(c *golangsdk.ServiceClient, id string) (*golangsdk.ErrRespond, error) {
	return waf.Delete(c, id).Extract()
}

var listAndExtract = func(c *golangsdk.ServiceClient, opts waf.ListOptsBuilder) ([]waf.Certificate, error) {
	pages, err := waf.List(c, opts).AllPages()
	if err != nil {
		log.Println(err)
		return []waf.Certificate{}, err
	}
	return waf.ExtractCertificates(pages)
}

var updateDomainAndExtract = func(
	c *golangsdk.ServiceClient,
	domainID string,
	opts wafDomain.UpdateOptsBuilder) (*wafDomain.Domain, error) {
	return wafDomain.Update(c, domainID, opts).Extract()
}

func CreateOrUpdateCertificate(secret apiv1.Secret) (*string, error) {
	certSecret := getCertificateSecret(secret)
	certIdInWaf, err := findCertInWaf(certSecret)

	if err != nil {
		return nil, err
	}

	if certIdInWaf != nil {
		log.Println("the certificate already exists in the waf")
		return certIdInWaf, nil
	} else {
		log.Println("the certificate does not exist in the waf yet...")
		certId, err := uploadNewCertificate(certSecret)

		if err != nil {
			return nil, err
		}

		err = attachCertificateToWafDomain(certSecret.wafDomainId, *certId)

		if err != nil {
			return nil, err
		}

		if len(certSecret.certWafId) > 0 {
			deletePreviousCertificate(certSecret.certWafId)
		}
		return certId, nil
	}
}

func attachCertificateToWafDomain(domainId string, certId string) error {
	_, err := updateDomainAndExtract(wafClient, domainId, wafDomain.UpdateOpts{
		CertificateId: certId,
	})
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("certificate %s has been attached to waf domain %s successfully", certId, domainId)
	return nil
}

func deletePreviousCertificate(id string) {
	_, err := deleteAndExtract(wafClient, id)
	if err != nil {
		log.Println("previous certificate couldn't be deleted", err)
	}
	log.Printf("previous certificate with id %s was deleted successfully", id)
}

func findCertInWaf(secret CertificateSecret) (*string, error) {
	log.Println("trying to find certificate in the waf...")
	certs, err := listAndExtract(wafClient, waf.ListOpts{})
	if err != nil {
		log.Println("couldn't get existing certificates from the waf ", err)
		return nil, err
	}

	found := funk.Find(certs, func(cert waf.Certificate) bool {
		if cert.Name == secret.certName {
			log.Println("the certificate was found!")
			return true
		}
		return false
	})

	if found != nil {
		certId := found.(waf.Certificate).Id
		return &certId, nil
	} else {
		return nil, nil
	}
}

func uploadNewCertificate(certSecret CertificateSecret) (*string, error) {
	log.Println("uploading a new certificate to web application firewall...")
	log.Println("certificate domain name: " + certSecret.domainName)

	createOpts := waf.CreateOpts{
		Name:    certSecret.certName,
		Content: certSecret.tlsCert,
		Key:     certSecret.tlsKey,
	}

	certificate, err := createAndExtract(wafClient, createOpts)
	if err != nil {
		log.Println("certificate couldn't be uploaded ", err)
		return nil, err
	}

	log.Println("created a new certificate in waf with id " + certificate.Id)
	return &certificate.Id, nil
}

func getCertificateSecret(secret apiv1.Secret) CertificateSecret {
	tlsCertificate := string(secret.Data["tls.crt"])
	tlsKey := string(secret.Data["tls.key"])
	certWafId := secret.Labels["cert-waf-id"]
	wafDomainId := secret.Labels["waf-domain-id"]

	regex := regexp.MustCompile(`\r?\n`)
	tlsCertificate = regex.ReplaceAllString(tlsCertificate, "")
	tlsKey = regex.ReplaceAllString(tlsKey, "")

	certHash := sha256.New()
	certHash.Write([]byte(tlsCertificate))
	certHashString := hex.EncodeToString(certHash.Sum(nil))

	return CertificateSecret{
		certName:    certHashString,
		tlsCert:     tlsCertificate,
		tlsKey:      tlsKey,
		domainName:  secret.Annotations["cert-manager.io/certificate-name"],
		certWafId:   certWafId,
		wafDomainId: wafDomainId,
	}
}
