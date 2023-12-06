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

func SetupOtcClient(clientSet *kubernetes.Clientset) {
	secret := getOtcCredentials(clientSet)
	authOpts := getAuthOptions(secret)
	provider := createProviderClient(authOpts)
	createWafServiceClient(provider)
}

func createWafServiceClient(provider *golangsdk.ProviderClient) {
	opts := golangsdk.EndpointOpts{Region: "eu-de"}
	var err error
	WafClient, err = openstack.NewWAFV1(provider, opts)

	if err != nil {
		log.Fatal("error creating waf service client", err)
	}
	log.Println("new waf client created successfully!")
}

func createProviderClient(authOpts golangsdk.AuthOptions) *golangsdk.ProviderClient {
	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		log.Fatal("error creating otc client", err)
	}

	log.Println("new otc client created successfully!")
	return provider
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

func getOtcCredentials(clientSet *kubernetes.Clientset) *apiv1.Secret {
	secret, err := clientSet.CoreV1().
		Secrets("default").
		Get(context.Background(), "otc-credentials", metav1.GetOptions{})

	if err != nil {
		log.Fatal("error getting kubernetes secrets", err)
	}
	return secret
}
