package waf_webhook

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
		Name:    fmt.Sprintf("91867d470a43a40a1413bce9c72e733d41760ac135b5737d69915a9148bce0a8"),
		Content: "-----BEGIN CERTIFICATE-----\nMIIDIjCCAougAwIBAgIJALV96mEtVF4EMA0GCSqGSIb3DQEBBQUAMGoxCzAJBgNV\nBAYTAnh4MQswCQYDVQQIEwJ4eDELMAkGA1UEBxMCeHgxCzAJBgNVBAoTAnh4MQsw\nCQYDVQQLEwJ4eDELMAkGA1UEAxMCeHgxGjAYBgkqhkiG9w0BCQEWC3h4eEAxNjMu\nY29tMB4XDTE3MTExMzAyMjYxM1oXDTIwMTExMjAyMjYxM1owajELMAkGA1UEBhMC\neHgxCzAJBgNVBAgTAnh4MQswCQYDVQQHEwJ4eDELMAkGA1UEChMCeHgxCzAJBgNV\nBAsTAnh4MQswCQYDVQQDEwJ4eDEaMBgGCSqGSIb3DQEJARYLeHh4QDE2My5jb20w\ngZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAMU832iM+d3FILgTWmpZBUoYcIWV\ncAAYE7FsZ9LNerOyjJpyi256oypdBvGs9JAUBN5WaFk81UQx29wAyNixX+bKa0DB\nWpUDqr84V1f9vdQc75v9WoujcnlKszzpV6qePPC7igJJpu4QOI362BrWzJCYQbg4\nUzo1KYBhLFxl0TovAgMBAAGjgc8wgcwwHQYDVR0OBBYEFMbTvDyvE2KsRy9zPq/J\nWOjovG+WMIGcBgNVHSMEgZQwgZGAFMbTvDyvE2KsRy9zPq/JWOjovG+WoW6kbDBq\nMQswCQYDVQQGEwJ4eDELMAkGA1UECBMCeHgxCzAJBgNVBAcTAnh4MQswCQYDVQQK\nEwJ4eDELMAkGA1UECxMCeHgxCzAJBgNVBAMTAnh4MRowGAYJKoZIhvcNAQkBFgt4\neHhAMTYzLmNvbYIJALV96mEtVF4EMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQEF\nBQADgYEAASkC/1iwiALa2RU3YCxqZFEEsZZvQxikrDkDbFeoa6Tk49Fnb1f7FCW6\nPTtY3HPWl5ygsMsSy0Fi3xp3jmuIwzJhcQ3tcK5gC99HWp6Kw37RL8WoB8GWFU0Q\n4tHLOjBIxkZROPRhH+zMIrqUexv6fsb3NWKhnlfh1Mj5wQE4Ldo=\n-----END CERTIFICATE-----",
		Key:     "-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQDFPN9ojPndxSC4E1pqWQVKGHCFlXAAGBOxbGfSzXqzsoyacotu\neqMqXQbxrPSQFATeVmhZPNVEMdvcAMjYsV/mymtAwVqVA6q/OFdX/b3UHO+b/VqL\no3J5SrM86Veqnjzwu4oCSabuEDiN+tga1syQmEG4OFM6NSmAYSxcZdE6LwIDAQAB\nAoGBAJvLzJCyIsCJcKHWL6onbSUtDtyFwPViD1QrVAtQYabF14g8CGUZG/9fgheu\nTXPtTDcvu7cZdUArvgYW3I9F9IBb2lmF3a44xfiAKdDhzr4DK/vQhvHPuuTeZA41\nr2zp8Cu+Bp40pSxmoAOK3B0/peZAka01Ju7c7ZChDWrxleHZAkEA/6dcaWHotfGS\neW5YLbSms3f0m0GH38nRl7oxyCW6yMIDkFHURVMBKW1OhrcuGo8u0nTMi5IH9gRg\n5bH8XcujlQJBAMWBQgzCHyoSeryD3TFieXIFzgDBw6Ve5hyMjUtjvgdVKoxRPvpO\nkclc39QHP6Dm2wrXXHEej+9RILxBZCVQNbMCQQC42i+Ut0nHvPuXN/UkXzomDHde\nh1ySsOAO4H+8Y6OSI87l3HUrByCQ7stX1z3L0HofjHqV9Koy9emGTFLZEzSdAkB7\nEi6cUKKmztkYe3rr+RcATEmwAw3tEJOHmrW5ErApVZKr2TzLMQZ7WZpIPzQRCYnY\n2ZZLDuZWFFG3vW+wKKktAkAaQ5GNzbwkRLpXF1FZFuNF7erxypzstbUmU/31b7tS\ni5LmxTGKL/xRYtZEHjya4Ikkkgt40q1MrUsgIYbFYMf2\n-----END RSA PRIVATE KEY-----",
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
		"c11564ffee8a4728bffcd58c16699231",
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
