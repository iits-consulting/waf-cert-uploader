package adapter

import (
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	waf "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/certificates"
	wafDomain "github.com/opentelekomcloud/gophertelekomcloud/openstack/waf/v1/domains"
	"log"
)

var CreateAndExtract = func(c *golangsdk.ServiceClient, opts waf.CreateOpts) (*waf.Certificate, error) {
	return waf.Create(c, opts).Extract()
}

var DeleteAndExtract = func(c *golangsdk.ServiceClient, id string) (*golangsdk.ErrRespond, error) {
	return waf.Delete(c, id).Extract()
}

var ListAndExtract = func(c *golangsdk.ServiceClient, opts waf.ListOptsBuilder) ([]waf.Certificate, error) {
	pages, err := waf.List(c, opts).AllPages()
	if err != nil {
		log.Println(err)
		return []waf.Certificate{}, err
	}
	return waf.ExtractCertificates(pages)
}

var UpdateDomainAndExtract = func(
	c *golangsdk.ServiceClient,
	domainID string,
	opts wafDomain.UpdateOptsBuilder) (*wafDomain.Domain, error) {
	return wafDomain.Update(c, domainID, opts).Extract()
}
