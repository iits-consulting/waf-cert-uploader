package service

import (
	"context"
	"errors"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"testing"
)

func TestSetupOtcClient(t *testing.T) {
	clientSet := kubernetes.Clientset{}
	var authOptsSlot golangsdk.AuthOptionsProvider
	var providerAddress *golangsdk.ProviderClient
	var providerSlot *golangsdk.ProviderClient
	var endpointOptsSlot golangsdk.EndpointOpts
	var clientSetSlot *kubernetes.Clientset
	var namespaceSlot string
	var secretNameSlot string
	var serviceClientSlot *golangsdk.ServiceClient

	getSecret = func(clientSet kubernetes.Clientset, namespace string, context context.Context,
		secretName string, getOptions metav1.GetOptions) (*apiv1.Secret, error) {
		clientSetSlot = &clientSet
		namespaceSlot = namespace
		secretNameSlot = secretName

		return &apiv1.Secret{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Data: map[string][]byte{
				"username":   []byte("Robin"),
				"password":   []byte("abc123"),
				"domainName": []byte("asdf5455fd4"),
				"tenantID":   []byte("qwer541235g3"),
			},
		}, nil
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

	err := SetupOtcClient(&clientSet)

	assert.Nil(t, err)
	assert.Equal(t, "Robin", authOptsSlot.(golangsdk.AuthOptions).Username)
	assert.Equal(t, "abc123", authOptsSlot.(golangsdk.AuthOptions).Password)
	assert.Equal(t, "asdf5455fd4", authOptsSlot.(golangsdk.AuthOptions).DomainName)
	assert.Equal(t, "qwer541235g3", authOptsSlot.(golangsdk.AuthOptions).TenantID)
	assert.Equal(t, "https://iam.eu-de.otc.t-systems.com:443/v3", authOptsSlot.(golangsdk.AuthOptions).IdentityEndpoint)
	assert.Equal(t, true, authOptsSlot.(golangsdk.AuthOptions).AllowReauth)
	assert.Equal(t, providerAddress, providerSlot)
	assert.Equal(t, "eu-de", endpointOptsSlot.Region)
	assert.Equal(t, &clientSet, clientSetSlot)
	assert.Equal(t, "default", namespaceSlot)
	assert.Equal(t, "otc-auth-options", secretNameSlot)
	assert.Equal(t, WafClient, serviceClientSlot)
}

func TestSetupOtcClient_providerFails(t *testing.T) {
	clientSet := kubernetes.Clientset{}

	getSecret = func(clientSet kubernetes.Clientset, namespace string, context context.Context,
		secretName string, getOptions metav1.GetOptions) (*apiv1.Secret, error) {

		return &apiv1.Secret{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Data: map[string][]byte{
				"username":   []byte("Robin"),
				"password":   []byte("abc123"),
				"domainName": []byte("asdf5455fd4"),
				"tenantID":   []byte("qwer541235g3"),
			},
		}, nil
	}
	getProviderClient = func(authOpts golangsdk.AuthOptionsProvider) (*golangsdk.ProviderClient, error) {
		return nil, errors.New("auth fail")
	}

	err := SetupOtcClient(&clientSet)

	assert.Equal(t, "auth fail", err.Error())
}

func TestSetupOtcClient_secretFails(t *testing.T) {
	clientSet := kubernetes.Clientset{}

	getSecret = func(clientSet kubernetes.Clientset, namespace string, context context.Context,
		secretName string, getOptions metav1.GetOptions) (*apiv1.Secret, error) {

		return nil, errors.New("secret not found")
	}

	err := SetupOtcClient(&clientSet)

	assert.Equal(t, "secret not found", err.Error())
}

func TestSetupOtcClient_wafClientFails(t *testing.T) {
	clientSet := kubernetes.Clientset{}

	getSecret = func(clientSet kubernetes.Clientset, namespace string, context context.Context,
		secretName string, getOptions metav1.GetOptions) (*apiv1.Secret, error) {

		return &apiv1.Secret{}, nil
	}
	getProviderClient = func(authOpts golangsdk.AuthOptionsProvider) (*golangsdk.ProviderClient, error) {
		provider := golangsdk.ProviderClient{}
		return &provider, nil
	}
	newWafV1 = func(provider *golangsdk.ProviderClient, opts golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
		return nil, errors.New("service client not created")
	}
	err := SetupOtcClient(&clientSet)

	assert.Equal(t, "service client not created", err.Error())
}
