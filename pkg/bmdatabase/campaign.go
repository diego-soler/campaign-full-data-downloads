package bmdatabase

import "github.com/Masterminds/semver"

type BmCampaign struct {
	IdCampaign       int64           `json:"idcampaign"`
	IdCampaignStatus int64           `json:"idcampaignstatus"`
	Version          *semver.Version `json:"version"`
	AdvertiserName   string          `json:"advertisername"`
}
