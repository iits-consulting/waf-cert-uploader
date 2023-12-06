package service

import (
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	waf "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/certificates"
	wafDomain "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/domains"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"waf-cert-uploader/adapter"
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

	adapter.ListAndExtract = func(c *golangsdk.ServiceClient, opts waf.ListOptsBuilder) ([]waf.Certificate, error) {
		functionCalls = append(functionCalls, "listAndExtract")
		return []waf.Certificate{}, nil
	}

	adapter.CreateAndExtract = func(c *golangsdk.ServiceClient, opts waf.CreateOpts) (*waf.Certificate, error) {
		functionCalls = append(functionCalls, "CreateAndExtract")
		return &waf.Certificate{Id: "1"}, nil
	}

	adapter.UpdateDomainAndExtract = func(
		c *golangsdk.ServiceClient,
		domainID string,
		opts wafDomain.UpdateOptsBuilder) (*wafDomain.Domain, error) {
		functionCalls = append(functionCalls, "UpdateDomainAndExtract")
		emptyWafDomain := wafDomain.Domain{}
		return &emptyWafDomain, nil
	}

	result, _ := CreateOrUpdateCertificate(secret)
	assert.Equal(t, "1", *result)
	assert.EqualValues(t, []string{"listAndExtract", "CreateAndExtract", "UpdateDomainAndExtract"}, functionCalls)
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

	adapter.ListAndExtract = func(c *golangsdk.ServiceClient, opts waf.ListOptsBuilder) ([]waf.Certificate, error) {
		functionCalls = append(functionCalls, "ListAndExtract")
		return []waf.Certificate{{
			Name: "4f4eb3c8aaf131baaf5d781449260177b6a4099240d8c999acb7b3b60cb318ed",
			Id:   "previous-id",
		}}, nil
	}

	result, _ := CreateOrUpdateCertificate(secret)

	assert.Equal(t, "previous-id", *result)
	assert.EqualValues(t, []string{"ListAndExtract"}, functionCalls)
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
	adapter.ListAndExtract = func(c *golangsdk.ServiceClient, opts waf.ListOptsBuilder) ([]waf.Certificate, error) {
		functionCalls = append(functionCalls, "ListAndExtract")
		return []waf.Certificate{}, nil
	}
	adapter.CreateAndExtract = func(c *golangsdk.ServiceClient, opts waf.CreateOpts) (*waf.Certificate, error) {
		functionCalls = append(functionCalls, "CreateAndExtract")
		return &waf.Certificate{Id: "new-id"}, nil
	}
	adapter.UpdateDomainAndExtract = func(
		c *golangsdk.ServiceClient,
		domainID string,
		opts wafDomain.UpdateOptsBuilder) (*wafDomain.Domain, error) {
		functionCalls = append(functionCalls, "UpdateDomainAndExtract")
		emptyWafDomain := wafDomain.Domain{}
		return &emptyWafDomain, nil
	}
	adapter.DeleteAndExtract = func(c *golangsdk.ServiceClient, id string) (*golangsdk.ErrRespond, error) {
		functionCalls = append(functionCalls, "DeleteAndExtract")
		errRespond := golangsdk.ErrRespond{}
		return &errRespond, nil
	}

	result, _ := CreateOrUpdateCertificate(secret)
	assert.Equal(t, "new-id", *result)
	assert.EqualValues(t, []string{"ListAndExtract", "CreateAndExtract",
		"UpdateDomainAndExtract", "DeleteAndExtract"}, functionCalls)
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

	adapter.ListAndExtract = func(c *golangsdk.ServiceClient, opts waf.ListOptsBuilder) ([]waf.Certificate, error) {
		functionCalls = append(functionCalls, "ListAndExtract")
		return []waf.Certificate{}, golangsdk.BaseError{Info: "error occurred"}
	}

	result, err := CreateOrUpdateCertificate(secret)

	assert.Equal(t, "error occurred", err.Error())
	assert.Nil(t, result)
	assert.EqualValues(t, []string{"ListAndExtract"}, functionCalls)
}

func setupWafTestClient() {
	provider := &golangsdk.ProviderClient{}
	WafClient = &golangsdk.ServiceClient{
		ProviderClient: provider,
	}
}
