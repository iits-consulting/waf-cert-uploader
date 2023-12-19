package service

import (
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	"log"
	"os"
)

var WafClient *golangsdk.ServiceClient

type OtcAuthOptionsSecret struct {
	username   string
	password   string
	accessKey  string
	secretKey  string
	domainName string
	tenantName string
}

var newWafV1 = func(
	provider *golangsdk.ProviderClient,
	opts golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return openstack.NewWAFV1(provider, opts)
}

var getProviderClient = func(authOpts golangsdk.AuthOptionsProvider) (*golangsdk.ProviderClient, error) {
	return openstack.AuthenticatedClient(authOpts)
}

func SetupOtcClient() error {
	provider, err := createProviderClient()
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

func createProviderClient() (*golangsdk.ProviderClient, error) {
	authOptsProvider, err := getAuthOptions()
	if err != nil {
		log.Println("couldn't get auth options", err)
		return nil, err
	}
	provider, err := getProviderClient(*authOptsProvider)
	if err != nil {
		log.Println("error creating otc client", err)
		return nil, err
	}

	log.Println("new otc client created successfully!")
	return provider, nil
}

func getAuthOptions() (*golangsdk.AuthOptionsProvider, error) {
	authOptions, err := getAuthOptionsFromMountedSecret()
	if err != nil {
		return nil, err
	}

	var authOptsProvider golangsdk.AuthOptionsProvider
	if len(authOptions.accessKey) > 0 && len(authOptions.secretKey) > 0 {
		authOptsProvider = golangsdk.AKSKAuthOptions{
			IdentityEndpoint: "https://iam.eu-de.otc.t-systems.com:443/v3",
			Region:           "eu-de",
			ProjectName:      authOptions.tenantName,
			Domain:           authOptions.domainName,
			AccessKey:        authOptions.accessKey,
			SecretKey:        authOptions.secretKey,
		}
		return &authOptsProvider, nil
	}

	authOptsProvider = golangsdk.AuthOptions{
		IdentityEndpoint: "https://iam.eu-de.otc.t-systems.com:443/v3",
		Username:         authOptions.username,
		Password:         authOptions.password,
		DomainName:       authOptions.domainName,
		TenantName:       authOptions.tenantName,
		AllowReauth:      true,
	}
	return &authOptsProvider, nil
}

var getAuthOptionsFromMountedSecret = func() (*OtcAuthOptionsSecret, error) {
	username, err := os.ReadFile("/etc/otc-auth-options/username")
	password, err := os.ReadFile("/etc/otc-auth-options/password")
	accessKey, err := os.ReadFile("/etc/otc-auth-options/accessKey")
	secretKey, err := os.ReadFile("/etc/otc-auth-options/secretKey")
	domainName, err := os.ReadFile("/etc/otc-auth-options/domainName")
	tenantName, err := os.ReadFile("/etc/otc-auth-options/tenantName")

	if err != nil {
		return nil, err
	}

	return &OtcAuthOptionsSecret{
		username:   string(username),
		password:   string(password),
		accessKey:  string(accessKey),
		secretKey:  string(secretKey),
		domainName: string(domainName),
		tenantName: string(tenantName),
	}, nil
}
