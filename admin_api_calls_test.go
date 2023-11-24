package waf_webhook

import (
	"fmt"
	"github.com/joho/godotenv"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	waf "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/certificates"
	"os"
	"strings"
	"testing"
	"time"
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
	dt := time.Now()
	createOpts := waf.CreateOpts{
		Name:    fmt.Sprintf("tls-cert-%s", dt.Format("01-02-2006-15-04-05")),
		Content: os.Getenv("TLS_CERT"),
		Key:     os.Getenv("TLS_KEY"),
	}

	certificate, err := waf.Create(wafClientTest, createOpts).Extract()

	if err != nil {
		fmt.Errorf("certificate creation failed")
		panic(err.Error())
	}

	fmt.Println("created a new certificate with id " + certificate.Id)
}

func Test_ListCertificates(t *testing.T) {
	pageWaf, err := waf.List(wafClientTest, waf.ListOpts{}).AllPages()
	if err != nil {
		panic(err.Error())
	}

	extracted, err := waf.ExtractCertificates(pageWaf)

	var certList []string
	for _, certificate := range extracted {
		certList = append(certList, "cert name: "+certificate.Name+" | cert id: "+certificate.Id)
	}

	if err != nil {
		panic(err.Error())
	}

	fmt.Println("certificate list:\n" + strings.Join(certList, "\n"))
}

func Test_DeleteCertificate(t *testing.T) {
	deleteResult := waf.Delete(wafClientTest, "1e9e8e1125204e4c916c1d2649b2760b")

	extracted, err := deleteResult.Extract()

	if err != nil {
		fmt.Println(extracted.ErrorMsg)
		panic(err.Error())
	}
	fmt.Println("certificate deleted successfully")
}
