package service

import (
	"context"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

var WafClient *golangsdk.ServiceClient

var newWafV1 = func(
	provider *golangsdk.ProviderClient,
	opts golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return openstack.NewWAFV1(provider, opts)
}

var getSecret = func(
	clientSet kubernetes.Clientset,
	namespace string,
	context context.Context,
	secretName string,
	getOptions metav1.GetOptions) (*apiv1.Secret, error) {
	return clientSet.CoreV1().Secrets(namespace).Get(context, secretName, getOptions)
}

var getProviderClient = func(authOpts golangsdk.AuthOptions) (*golangsdk.ProviderClient, error) {
	return openstack.AuthenticatedClient(authOpts)
}

func SetupOtcClient(clientSet *kubernetes.Clientset) error {
	secret, err := getOtcCredentials(clientSet)
	if err != nil {
		return err
	}
	authOpts := getAuthOptions(secret)
	provider, err := createProviderClient(authOpts)
	if err != nil {
		return err
	}
	err = createWafServiceClient(provider)
	if err != nil {
		return err
	}
	return nil
}

func createWafServiceClient(provider *golangsdk.ProviderClient) error {
	opts := golangsdk.EndpointOpts{Region: "eu-de"}
	var err error
	WafClient, err = newWafV1(provider, opts)

	if err != nil {
		log.Println("error creating waf service client", err)
		return err
	}
	log.Println("new waf client created successfully!")
	return nil
}

func createProviderClient(authOpts golangsdk.AuthOptions) (*golangsdk.ProviderClient, error) {
	provider, err := getProviderClient(authOpts)
	if err != nil {
		log.Println("error creating otc client", err)
		return nil, err
	}

	log.Println("new otc client created successfully!")
	return provider, nil
}

func getAuthOptions(secret *apiv1.Secret) golangsdk.AuthOptions {
	return golangsdk.AuthOptions{
		IdentityEndpoint: "https://iam.eu-de.otc.t-systems.com:443/v3",
		Username:         string(secret.Data["username"]),
		Password:         string(secret.Data["password"]),
		DomainID:         string(secret.Data["domainID"]),
		TenantID:         string(secret.Data["tenantID"]),
		AllowReauth:      true,
	}
}

func getOtcCredentials(clientSet *kubernetes.Clientset) (*apiv1.Secret, error) {
	secret, err := getSecret(*clientSet, "default", context.Background(), "otc-credentials", metav1.GetOptions{})

	if err != nil {
		log.Println("error getting kubernetes secrets", err)
		return nil, err
	}
	return secret, nil
}
