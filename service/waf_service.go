package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	waf "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/certificates"
	wafDomain "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/domains"
	"github.com/thoas/go-funk"
	apiv1 "k8s.io/api/core/v1"
	"log"
	"regexp"
	"waf-cert-uploader/adapter"
)

type CertificateSecret struct {
	certName    string
	tlsCert     string
	tlsKey      string
	domainName  string
	wafDomainId string
	certWafId   string
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
	newServerOpts, err := getNewServerOpts(domainId)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = adapter.UpdateDomainAndExtract(WafClient, domainId, wafDomain.UpdateOpts{
		CertificateId: certId,
		Server:        *newServerOpts,
	})
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("certificate %s has been attached to waf domain %s successfully", certId, domainId)
	return nil
}

func getNewServerOpts(domainId string) (*[]wafDomain.ServerOpts, error) {
	existingDomain, err := adapter.GetWafDomainAndExtract(WafClient, domainId)
	if err != nil {
		return nil, err
	}
	newOpts := []wafDomain.ServerOpts{
		{
			ClientProtocol: "HTTPS",
			ServerProtocol: "HTTPS",
			Address:        existingDomain.Server[0].Address,
			Port:           443,
		},
		{
			ClientProtocol: "HTTP",
			ServerProtocol: "HTTP",
			Address:        existingDomain.Server[0].Address,
			Port:           80,
		},
	}
	return &newOpts, nil
}

func deletePreviousCertificate(id string) {
	_, err := adapter.DeleteAndExtract(WafClient, id)
	if err != nil {
		log.Println("previous certificate couldn't be deleted", err)
	} else {
		log.Printf("previous certificate with id %s was deleted successfully", id)
	}
}

func findCertInWaf(secret CertificateSecret) (*string, error) {
	log.Println("trying to find certificate in the waf...")
	certs, err := adapter.ListAndExtract(WafClient, waf.ListOpts{})
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

	certificate, err := adapter.CreateAndExtract(WafClient, createOpts)
	if err != nil {
		log.Println("certificate couldn't be uploaded ", err)
		return nil, err
	}

	log.Println("created a new certificate in waf with id " + certificate.Id)
	return &certificate.Id, nil
}

func getCertificateSecret(secret apiv1.Secret) CertificateSecret {
	tlsCertificateBase64 := secret.Data["tls.crt"]
	tlsKeyBase64 := secret.Data["tls.key"]

	certWafId := secret.Annotations["cert-waf-id"]
	wafDomainId := secret.Annotations["waf-domain-id"]

	//newline must be removed, else waf api fails
	tlsCertificate := string(tlsCertificateBase64)
	tlsKey := string(tlsKeyBase64)

	tlsCertificateDecoded, _ := base64.StdEncoding.DecodeString(tlsCertificate)
	tlsKeyDecoded, _ := base64.StdEncoding.DecodeString(tlsKey)

	certHashString := getCertificateHash(tlsCertificateBase64)

	return CertificateSecret{
		certName:    certHashString,
		tlsCert:     string(tlsCertificateDecoded),
		tlsKey:      string(tlsKeyDecoded),
		domainName:  secret.Annotations["cert-manager.io/certificate-name"],
		certWafId:   certWafId,
		wafDomainId: wafDomainId,
	}
}

func getCertificateHash(tlsCertificateBase64 []byte) string {
	certHash := sha256.New()
	certHash.Write(tlsCertificateBase64)
	certHashString := hex.EncodeToString(certHash.Sum(nil))
	return certHashString
}

//func replaceLineEndingsWithNewline(stringToConvert string)

func getStringWithoutNewline(stringWithNewline string) string {
	regex := regexp.MustCompile(`\r?\n`)
	return regex.ReplaceAllString(stringWithNewline, "")
}
