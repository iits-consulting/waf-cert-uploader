package internal

import (
	"fmt"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"golang.org/x/exp/slices"
)

type AuthenticationData struct {
	AccessKey   string
	SecretKey   string
	ProjectName string
	DomainName  string
	Region      OtcRegion
}

func (ad AuthenticationData) ToOtcGopherAuthOptionsProvider() golangsdk.AuthOptionsProvider {
	return golangsdk.AKSKAuthOptions{
		IdentityEndpoint: ad.Region.IamEndpoint(),
		AccessKey:        ad.AccessKey,
		SecretKey:        ad.SecretKey,
		Domain:           ad.DomainName,
		ProjectName:      ad.ProjectName,
	}
}

type OtcRegion string

const (
	otcRegionEuDe OtcRegion = "eu-de"
	otcRegionEuNl OtcRegion = "eu-nl"
)

func NewOtcRegionFromString(region string) (OtcRegion, error) {
	otcRegion := OtcRegion(region)
	if slices.Contains([]OtcRegion{otcRegionEuNl, otcRegionEuDe}, otcRegion) {
		return otcRegion, nil
	}

	return "", fmt.Errorf("invalid argument %s does not represent a valid region", region)
}

func (r OtcRegion) IamEndpoint() string {
	return fmt.Sprintf("https://iam.%s.otc.t-systems.com:443/v3", r)
}
