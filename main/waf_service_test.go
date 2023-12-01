package main

//import (
//	"errors"
//	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
//	waf "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/certificates"
//	"github.com/opentelekomcloud/gophertelekomcloud/pagination"
//	"github.com/stretchr/testify/assert"
//	apiv1 "k8s.io/api/core/v1"
//	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"testing"
//)
//
//func TestCreateOrUpdateCertificate_Create(t *testing.T) {
//	setupWafTestClient()
//	var functionCalls []string
//	wafCreate = func(c *golangsdk.ServiceClient, opts waf.CreateOpts) (r waf.CreateResult) {
//		functionCalls = append(functionCalls, "wafCreate")
//		return waf.CreateResult{}
//	}
//	createExtract = func(createResult waf.CreateResult) (*waf.Certificate, error) {
//		functionCalls = append(functionCalls, "createExtract")
//		return &waf.Certificate{Id: "1"}, nil
//	}
//
//	secret := apiv1.Secret{
//		Data: map[string][]byte{
//			"tls.crt": []byte("any cert"),
//			"tls.key": []byte("any private key"),
//		},
//	}
//
//	result := CreateOrUpdateCertificate(secret)
//	assert.Equal(t, "1", result)
//	assert.EqualValues(t, functionCalls, []string{"wafCreate", "createExtract"})
//}
//
//func TestCreateOrUpdateCertificate_Create_Fails(t *testing.T) {
//	secret := apiv1.Secret{
//		ObjectMeta: v1.ObjectMeta{
//			ResourceVersion: "version1",
//		},
//		Data: map[string][]byte{
//			"tls.crt": []byte("any cert"),
//			"tls.key": []byte("any private key"),
//		},
//	}
//	setupWafTestClient()
//	var functionCalls []string
//	wafCreate = func(c *golangsdk.ServiceClient, opts waf.CreateOpts) (r waf.CreateResult) {
//		functionCalls = append(functionCalls, "wafCreate")
//		return waf.CreateResult{}
//	}
//	createExtract = func(createResult waf.CreateResult) (*waf.Certificate, error) {
//		functionCalls = append(functionCalls, "createExtract")
//		return nil, errors.New("any error")
//	}
//	wafList = func(c *golangsdk.ServiceClient, opts waf.ListOptsBuilder) (p pagination.Pager) {
//		functionCalls = append(functionCalls, "wafList")
//		return pagination.Pager{}
//	}
//	allPages = func(p pagination.Pager) (pagination.Page, error) {
//		functionCalls = append(functionCalls, "allPages")
//		return waf.CertificatePage{}, nil
//	}
//	wafExtractCertificates = func(p pagination.Page) ([]waf.Certificate, error) {
//		functionCalls = append(functionCalls, "wafExtractCertificates")
//		return []waf.Certificate{{Id: "12345", Name: "tls-cert-" + secret.ResourceVersion}}, nil
//	}
//
//	result := CreateOrUpdateCertificate(secret)
//
//	assert.Equal(t, "12345", result)
//	assert.EqualValues(t,
//		functionCalls,
//		[]string{"wafCreate", "createExtract", "wafList", "allPages", "wafExtractCertificates"})
//}
//
//func TestCreateOrUpdateCertificate_Update_Successful(t *testing.T) {
//	setupWafTestClient()
//	var functionCalls []string
//	wafCreate = func(c *golangsdk.ServiceClient, opts waf.CreateOpts) (r waf.CreateResult) {
//		functionCalls = append(functionCalls, "wafCreate")
//		return waf.CreateResult{}
//	}
//	createExtract = func(createResult waf.CreateResult) (*waf.Certificate, error) {
//		functionCalls = append(functionCalls, "createExtract")
//		return &waf.Certificate{Id: "1"}, nil
//	}
//	wafDelete = func(c *golangsdk.ServiceClient, id string) (r waf.DeleteResult) {
//		functionCalls = append(functionCalls, "wafDelete")
//		return waf.DeleteResult{}
//	}
//	deleteExtract = func(deleteResult waf.DeleteResult) (*golangsdk.ErrRespond, error) {
//		functionCalls = append(functionCalls, "deleteExtract")
//		return nil, nil
//	}
//
//	secret := apiv1.Secret{
//		Data: map[string][]byte{
//			"tls.crt": []byte("any cert"),
//			"tls.key": []byte("any private key"),
//		},
//		ObjectMeta: v1.ObjectMeta{
//			Labels: map[string]string{
//				"cert-waf-id": "12345",
//			},
//		},
//	}
//
//	result := CreateOrUpdateCertificate(secret)
//	assert.Equal(t, "1", result)
//	assert.EqualValues(t, functionCalls, []string{"wafDelete", "deleteExtract", "wafCreate", "createExtract"})
//}
//
//func setupWafTestClient() {
//	provider := &golangsdk.ProviderClient{}
//	wafClient = &golangsdk.ServiceClient{
//		ProviderClient: provider,
//	}
//}
