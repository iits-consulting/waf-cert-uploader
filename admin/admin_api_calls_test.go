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

var authOpts golangsdk.AuthOptions
var wafClientTest *golangsdk.ServiceClient

func TestMain(m *testing.M) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(".env file was not found")
	}
	if os.Getenv("CI") == "false" {
		authOpts = golangsdk.AuthOptions{
			IdentityEndpoint: os.Getenv("IAM_IDENTITY_ENDPOINT"),
			Username:         os.Getenv("IAM_USERNAME"),
			Password:         os.Getenv("IAM_PASSWORD"),
			DomainID:         os.Getenv("OTC_DOMAIN_ID"),
			TenantID:         os.Getenv("OTC_TENANT_ID"),
			AllowReauth:      true,
		}

		provider, err := openstack.AuthenticatedClient(authOpts)
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
	extracted, err := wafDomain.Get(wafClientTest, "442fb239f66d44c198665c2f4285d129").Extract()

	if err != nil {
		panic(err)
	}

	fmt.Println("host: " + extracted.HostName)
	fmt.Println("cert id: " + extracted.CertificateId)
}

func Test_AttachCertToWafDomain(t *testing.T) {
	opts := wafDomain.UpdateOpts{
		CertificateId: "1b7eb2febbdf4182ae624caa15058dc6",
		Server: []wafDomain.ServerOpts{{
			ClientProtocol: "HTTPS",
			ServerProtocol: "HTTPS",
			Address:        "80.158.32.51",
			Port:           443,
		}},
	}
	updateResult := wafDomain.Update(wafClientTest, "442fb239f66d44c198665c2f4285d129", opts)

	extracted, err := updateResult.Extract()

	if err != nil {
		panic(err)
	}
	fmt.Println(extracted.HostName)
	fmt.Println(extracted.CertificateId)
}

func Test_DeleteCertificates(t *testing.T) {
	idsToDelete := []string{
		"8d9c54648d8145d79aa1f8e6838ca3c5",
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
