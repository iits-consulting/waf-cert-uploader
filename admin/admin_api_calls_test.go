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
	extracted, err := wafDomain.Get(wafClientTest, "bedfc178b7c0474c93250ea951ce2f93").Extract()

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
		"6f47e9defbf643e8bfd64498e527deb4",
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
