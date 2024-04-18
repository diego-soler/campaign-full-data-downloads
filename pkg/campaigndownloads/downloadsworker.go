package campaigndownloads

import (
	"fmt"
	"sync"
)

func DownloadWorker(jobChannel chan CampaignDownloadJob, stdOutput chan string, stdError chan string, wg *sync.WaitGroup) {
	fmt.Println("Worker started. Wainting for jobs...")
	defer wg.Done()
	for job := range jobChannel {
		stdOutput <- fmt.Sprintf("Downloading campaign %d...\n", job.Campaign.IdCampaign)
		err := DownloadFile(&job.Campaign, job.UserToken, job.PlannerToken)
		if err != nil {
			stdOutput <- fmt.Sprintf("Campaign %d failed download.\n", job.Campaign.IdCampaign)
			stdError <- fmt.Sprintf("Error downloading campaign %d: %s\n", job.Campaign.IdCampaign, err.Error())
			continue
		}
		stdOutput <- fmt.Sprintf("Campaign %d downloaded successfully.\n", job.Campaign.IdCampaign)
	}
}
