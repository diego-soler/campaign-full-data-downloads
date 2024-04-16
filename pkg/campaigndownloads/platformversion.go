package campaigndownloads

import (
	"campaign-downloads/pkg/bmdatabase"

	"github.com/Masterminds/semver"
)

func PlatformVersion(campaign *bmdatabase.BmCampaign) string {
	minVersion, err := semver.NewVersion("1.2.0")
	if err != nil {
		panic(err)
	}
	if campaign.Version.Equal(minVersion) || campaign.Version.GreaterThan(minVersion) {
		return "v2"
	} else {
		return "v1"
	}
}
