package main

import (
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	waf "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/certificates"
	wafDomain "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/domains"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestCreateOrUpdateCertificate_Create(t *testing.T) {
	setupWafTestClient()
	secret := apiv1.Secret{
		ObjectMeta: v1.ObjectMeta{
			ResourceVersion: "version1",
			Labels: map[string]string{
				"waf-domain-id": "45656165da65456",
			},
		},
		Data: map[string][]byte{
			"tls.crt": []byte("any cert"),
			"tls.key": []byte("any private key"),
		},
	}
	var functionCalls []string

	listAndExtract = func(c *golangsdk.ServiceClient, opts waf.ListOptsBuilder) ([]waf.Certificate, error) {
		functionCalls = append(functionCalls, "listAndExtract")
		return []waf.Certificate{}, nil
	}

	createAndExtract = func(c *golangsdk.ServiceClient, opts waf.CreateOpts) (*waf.Certificate, error) {
		functionCalls = append(functionCalls, "createAndExtract")
		return &waf.Certificate{Id: "1"}, nil
	}

	updateDomainAndExtract = func(
		c *golangsdk.ServiceClient,
		domainID string,
		opts wafDomain.UpdateOptsBuilder) (*wafDomain.Domain, error) {
		functionCalls = append(functionCalls, "updateDomainAndExtract")
		emptyWafDomain := wafDomain.Domain{}
		return &emptyWafDomain, nil
	}

	result, _ := CreateOrUpdateCertificate(secret)
	assert.Equal(t, "1", *result)
	assert.EqualValues(t, []string{"listAndExtract", "createAndExtract", "updateDomainAndExtract"}, functionCalls)
}

func TestCreateOrUpdateCertificate_Create_AlreadyExists(t *testing.T) {
	secret := apiv1.Secret{
		Data: map[string][]byte{
			"tls.crt": []byte("any cert"),
			"tls.key": []byte("any private key"),
		},
	}
	setupWafTestClient()
	var functionCalls []string

	listAndExtract = func(c *golangsdk.ServiceClient, opts waf.ListOptsBuilder) ([]waf.Certificate, error) {
		functionCalls = append(functionCalls, "listAndExtract")
		return []waf.Certificate{{
			Name: "4f4eb3c8aaf131baaf5d781449260177b6a4099240d8c999acb7b3b60cb318ed",
			Id:   "previous-id",
		}}, nil
	}

	result, _ := CreateOrUpdateCertificate(secret)

	assert.Equal(t, "previous-id", *result)
	assert.EqualValues(t, []string{"listAndExtract"}, functionCalls)
}

func TestCreateOrUpdateCertificate_Update(t *testing.T) {
	setupWafTestClient()
	secret := apiv1.Secret{
		ObjectMeta: v1.ObjectMeta{
			ResourceVersion: "version1",
			Labels: map[string]string{
				"cert-waf-id":   "previous-id",
				"waf-domain-id": "45656165da65456",
			},
		},
		Data: map[string][]byte{
			"tls.crt": []byte("any cert"),
			"tls.key": []byte("any private key"),
		},
	}
	var functionCalls []string
	listAndExtract = func(c *golangsdk.ServiceClient, opts waf.ListOptsBuilder) ([]waf.Certificate, error) {
		functionCalls = append(functionCalls, "listAndExtract")
		return []waf.Certificate{}, nil
	}
	createAndExtract = func(c *golangsdk.ServiceClient, opts waf.CreateOpts) (*waf.Certificate, error) {
		functionCalls = append(functionCalls, "createAndExtract")
		return &waf.Certificate{Id: "new-id"}, nil
	}
	updateDomainAndExtract = func(
		c *golangsdk.ServiceClient,
		domainID string,
		opts wafDomain.UpdateOptsBuilder) (*wafDomain.Domain, error) {
		functionCalls = append(functionCalls, "updateDomainAndExtract")
		emptyWafDomain := wafDomain.Domain{}
		return &emptyWafDomain, nil
	}
	deleteAndExtract = func(c *golangsdk.ServiceClient, id string) (*golangsdk.ErrRespond, error) {
		functionCalls = append(functionCalls, "deleteAndExtract")
		errRespond := golangsdk.ErrRespond{}
		return &errRespond, nil
	}

	result, _ := CreateOrUpdateCertificate(secret)
	assert.Equal(t, "new-id", *result)
	assert.EqualValues(t, []string{"listAndExtract", "createAndExtract",
		"updateDomainAndExtract", "deleteAndExtract"}, functionCalls)
}

func TestCreateOrUpdateCertificate_Fails(t *testing.T) {
	setupWafTestClient()
	secret := apiv1.Secret{
		Data: map[string][]byte{
			"tls.crt": []byte("any cert"),
			"tls.key": []byte("any private key"),
		},
	}
	var functionCalls []string

	listAndExtract = func(c *golangsdk.ServiceClient, opts waf.ListOptsBuilder) ([]waf.Certificate, error) {
		functionCalls = append(functionCalls, "listAndExtract")
		return []waf.Certificate{}, golangsdk.BaseError{Info: "error occurred"}
	}

	result, err := CreateOrUpdateCertificate(secret)

	assert.Equal(t, "error occurred", err.Error())
	assert.Nil(t, result)
	assert.EqualValues(t, []string{"listAndExtract"}, functionCalls)
}

func setupWafTestClient() {
	provider := &golangsdk.ProviderClient{}
	wafClient = &golangsdk.ServiceClient{
		ProviderClient: provider,
	}
}
