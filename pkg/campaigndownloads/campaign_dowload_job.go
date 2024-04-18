package campaigndownloads

import "campaign-downloads/pkg/bmdatabase"

type CampaignDownloadJob struct {
	Campaign     bmdatabase.BmCampaign
	UserToken    string
	PlannerToken string
}
