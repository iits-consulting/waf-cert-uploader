package service

import (
	"errors"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetupOtcClient(t *testing.T) {
	var authOptsSlot golangsdk.AuthOptionsProvider
	var providerAddress *golangsdk.ProviderClient
	var providerSlot *golangsdk.ProviderClient
	var endpointOptsSlot golangsdk.EndpointOpts
	var serviceClientSlot *golangsdk.ServiceClient

	getAuthOptionsFromMountedSecret = func() error {
		authOptions = OtcAuthOptionsSecret{
			username:       "Robin",
			password:       "abc123",
			accessKey:      "",
			secretKey:      "",
			otcAccountName: "asdf5455fd4",
			projectName:    "qwer541235g3",
			region:         "eu-de",
		}
		return nil
	}
	getProviderClient = func(authOpts golangsdk.AuthOptionsProvider) (*golangsdk.ProviderClient, error) {
		authOptsSlot = authOpts
		provider := golangsdk.ProviderClient{}
		providerAddress = &provider
		return &provider, nil
	}
	newWafV1 = func(provider *golangsdk.ProviderClient, opts golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
		providerSlot = provider
		endpointOptsSlot = opts
		serviceClient := golangsdk.ServiceClient{}
		serviceClientSlot = &golangsdk.ServiceClient{}
		return &serviceClient, nil
	}

	err := SetupOtcClient()

	assert.Nil(t, err)
	assert.Equal(t, "Robin", authOptsSlot.(golangsdk.AuthOptions).Username)
	assert.Equal(t, "abc123", authOptsSlot.(golangsdk.AuthOptions).Password)
	assert.Equal(t, "asdf5455fd4", authOptsSlot.(golangsdk.AuthOptions).DomainName)
	assert.Equal(t, "qwer541235g3", authOptsSlot.(golangsdk.AuthOptions).TenantName)
	assert.Equal(t, "https://iam.eu-de.otc.t-systems.com:443/v3", authOptsSlot.(golangsdk.AuthOptions).IdentityEndpoint)
	assert.Equal(t, true, authOptsSlot.(golangsdk.AuthOptions).AllowReauth)
	assert.Equal(t, providerAddress, providerSlot)
	assert.Equal(t, "eu-de", endpointOptsSlot.Region)
	assert.Equal(t, WafClient, serviceClientSlot)
}

func TestSetupOtcClientForAkSk(t *testing.T) {
	var authOptsSlot golangsdk.AuthOptionsProvider
	var providerAddress *golangsdk.ProviderClient
	var providerSlot *golangsdk.ProviderClient
	var endpointOptsSlot golangsdk.EndpointOpts
	var serviceClientSlot *golangsdk.ServiceClient

	getAuthOptionsFromMountedSecret = func() error {
		authOptions = OtcAuthOptionsSecret{
			username:       "",
			password:       "",
			accessKey:      "access-key",
			secretKey:      "secret-key",
			otcAccountName: "",
			projectName:    "qwer541235g3",
			region:         "eu-de",
		}
		return nil
	}
	getProviderClient = func(authOpts golangsdk.AuthOptionsProvider) (*golangsdk.ProviderClient, error) {
		authOptsSlot = authOpts
		provider := golangsdk.ProviderClient{}
		providerAddress = &provider
		return &provider, nil
	}
	newWafV1 = func(provider *golangsdk.ProviderClient, opts golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
		providerSlot = provider
		endpointOptsSlot = opts
		serviceClient := golangsdk.ServiceClient{}
		serviceClientSlot = &golangsdk.ServiceClient{}
		return &serviceClient, nil
	}

	err := SetupOtcClient()

	assert.Nil(t, err)

	assert.Equal(t, "", authOptsSlot.(golangsdk.AKSKAuthOptions).Domain)
	assert.Equal(t, "", authOptsSlot.(golangsdk.AKSKAuthOptions).DomainID)
	assert.Equal(t, "qwer541235g3", authOptsSlot.(golangsdk.AKSKAuthOptions).ProjectName)
	assert.Equal(t, "access-key", authOptsSlot.(golangsdk.AKSKAuthOptions).AccessKey)
	assert.Equal(t, "secret-key", authOptsSlot.(golangsdk.AKSKAuthOptions).SecretKey)
	assert.Equal(t, "https://iam.eu-de.otc.t-systems.com:443/v3", authOptsSlot.(golangsdk.AKSKAuthOptions).IdentityEndpoint)
	assert.Equal(t, providerAddress, providerSlot)
	assert.Equal(t, "eu-de", endpointOptsSlot.Region)
	assert.Equal(t, WafClient, serviceClientSlot)
}

func TestSetupOtcClient_providerFails(t *testing.T) {

	getAuthOptionsFromMountedSecret = func() error {
		authOptions = OtcAuthOptionsSecret{}
		return nil
	}
	getProviderClient = func(authOpts golangsdk.AuthOptionsProvider) (*golangsdk.ProviderClient, error) {
		return nil, errors.New("auth fail")
	}

	err := SetupOtcClient()

	assert.Equal(t, "auth fail", err.Error())
}

func TestSetupOtcClient_wafClientFails(t *testing.T) {

	getAuthOptionsFromMountedSecret = func() error {
		authOptions = OtcAuthOptionsSecret{}
		return nil
	}
	getProviderClient = func(authOpts golangsdk.AuthOptionsProvider) (*golangsdk.ProviderClient, error) {
		provider := golangsdk.ProviderClient{}
		return &provider, nil
	}
	newWafV1 = func(provider *golangsdk.ProviderClient, opts golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
		return nil, errors.New("service client not created")
	}
	err := SetupOtcClient()

	assert.Equal(t, "service client not created", err.Error())
}
