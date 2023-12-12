package admin

import (
	"fmt"
	"github.com/joho/godotenv"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	waf "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/certificates"
	wafDomain "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/domains"
	"github.com/thoas/go-funk"
	"os"
	"testing"
)

var wafClientTest *golangsdk.ServiceClient

func TestMain(m *testing.M) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(".env file was not found")
	}
	if os.Getenv("CI") == "false" {
		var authProvider golangsdk.AuthOptionsProvider
		authProvider = golangsdk.AuthOptions{
			IdentityEndpoint: os.Getenv("IAM_IDENTITY_ENDPOINT"),
			Username:         os.Getenv("IAM_USERNAME"),
			Password:         os.Getenv("IAM_PASSWORD"),
			DomainName:       os.Getenv("OTC_DOMAIN_NAME"),
			TenantName:       os.Getenv("OTC_TENANT_NAME"),
		}

		//authProvider = golangsdk.AKSKAuthOptions{
		//	IdentityEndpoint: os.Getenv("IAM_IDENTITY_ENDPOINT"),
		//	Domain:           os.Getenv("OTC_DOMAIN_NAME"),
		//	ProjectName:      os.Getenv("OTC_TENANT_NAME"),
		//	AccessKey:        os.Getenv("ACCESS_KEY"),
		//	SecretKey:        os.Getenv("SECRET_KEY"),
		//}

		provider, err := openstack.AuthenticatedClient(authProvider)
		if err != nil {
			panic(err.Error())
		}

		opts := golangsdk.EndpointOpts{Region: "eu-de"}

		wafClientTest, err = openstack.NewWAFV1(provider, opts)

		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("waf client created successfully with token: " + wafClientTest.Token())

		code := m.Run()
		os.Exit(code)
	}
}

func Test_CreateCertificate(t *testing.T) {
	createOpts := waf.CreateOpts{
		Name:    fmt.Sprintf("test-cert"),
		Content: os.Getenv("TLS_CERT"),
		Key:     os.Getenv("TLS_KEY"),
	}

	certificate, err := waf.Create(wafClientTest, createOpts).Extract()

	if err != nil {
		fmt.Errorf("certificate creation failed")
		panic(err)
	}

	fmt.Println("created a new certificate with id " + certificate.Id)
}

func Test_ListCertificates(t *testing.T) {
	pageWaf, err := waf.List(wafClientTest, waf.ListOpts{}).AllPages()
	if err != nil {
		panic(err)
	}

	extracted, err := waf.ExtractCertificates(pageWaf)

	funk.ForEach(extracted, func(cert waf.Certificate) {
		fmt.Println("cert name: " + cert.Name + " | cert id: " + cert.Id)
	})
}

func Test_GetWafDomain(t *testing.T) {
	extracted, err := wafDomain.Get(wafClientTest, "").Extract()

	if err != nil {
		panic(err)
	}

	fmt.Println("host: " + extracted.HostName)
	fmt.Println("cert id: " + extracted.CertificateId)
}

func Test_AttachCertToWafDomain(t *testing.T) {
	opts := wafDomain.UpdateOpts{
		CertificateId: "9097d08958af4d1caac0d3efb0d27a4c",
		Server: []wafDomain.ServerOpts{{
			ClientProtocol: "HTTPS",
			ServerProtocol: "HTTPS",
			Address:        "waf-cert-uploader.iits.tech",
			Port:           443,
		}},
	}
	updateResult := wafDomain.Update(wafClientTest, "bedfc178b7c0474c93250ea951ce2f93", opts)

	extracted, err := updateResult.Extract()

	if err != nil {
		panic(err)
	}
	fmt.Println(extracted.HostName)
	fmt.Println(extracted.CertificateId)
}

func Test_DeleteCertificates(t *testing.T) {
	idsToDelete := []string{
		"b521c923925c476292e54331ba9e81e3",
		"2b3b9bf569bb4356b86c3d699f6f0e2b",
		"eee036d8bc3a4aba84f8f431b49b20bb",
		"9097d08958af4d1caac0d3efb0d27a4c",
		"cef22f6aaa21433fbbfe131063f6fab7",
		"b423c4e522134b53aea700f02f30f59f",
		"62f392ef3a0341e8978891b4594fe1f0",
		"3bc0383314424b9cb1c732a2ec236133",
	}

	funk.ForEach(idsToDelete, func(id string) {
		deleteResult := waf.Delete(wafClientTest, id)

		_, err := deleteResult.Extract()

		if err != nil {
			panic(err.Error())
		}

		fmt.Println("certificate deleted successfully")
	})

}
