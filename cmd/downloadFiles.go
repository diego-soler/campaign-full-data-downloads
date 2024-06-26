/*
Copyright © 2024 Diego Soler <solerdiego@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"campaign-downloads/pkg/bmdatabase"
	"campaign-downloads/pkg/campaigndownloads"
	"os"
	"runtime"
	"sync"

	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// downloadFilesCmd represents the downloadFiles command
var downloadFilesCmd = &cobra.Command{
	Use:   "downloadFiles",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		env := viper.GetString("GO_ENV")

		var userEmail string
		if env == "development" {
			userEmail = "admin@beeyondmedia.com"
		} else {
			userEmail = "rosana.admin.prod@beeyondmedia.com"
		}

		v1Flag, _ := cmd.Flags().GetBool("v1")
		v2Flag, _ := cmd.Flags().GetBool("v2")

		if !v1Flag && !v2Flag {
			v1Flag = true
			v2Flag = true
		}

		// Use the 'version' variable in your code to download files accordingly

		var campaignStatusId int64 = 4
		campaigns, err := bmdatabase.GetCampaigns(campaignStatusId)
		if err != nil {
			panic(err)
		}

		jobChannel := make(chan campaigndownloads.CampaignDownloadJob)
		var wg sync.WaitGroup
		stdOutput := make(chan string)
		stdError := make(chan string)
		numWorkers := runtime.NumCPU() / 2

		// Add the number of worker goroutines to the wait group
		wg.Add(numWorkers)

		// Launch worker goroutines
		for i := 0; i < numWorkers; i++ {
			go campaigndownloads.DownloadWorker(jobChannel, stdOutput, stdError, &wg)
		}
		// Send the jobs to the worker goroutines
		go func() {
			fmt.Println("Starting to send jobs to workers")
			for i, campaign := range campaigns {
				version := campaigndownloads.PlatformVersion(&campaign)

				if (version == "v1" && !v1Flag) || (version == "v2" && !v2Flag) {
					continue
				}

				user, err := bmdatabase.GetUser(userEmail)
				if err != nil {
					panic(err)
				}
				plannerToken, err := bmdatabase.GetCampaignPlannerToken(campaign.IdCampaign)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting planner token for campaign %d\n", campaign.IdCampaign)
					continue
				}
				var job campaigndownloads.CampaignDownloadJob = campaigndownloads.CampaignDownloadJob{
					Campaign:     campaign,
					UserToken:    user.Token,
					PlannerToken: plannerToken,
				}
				fmt.Printf("Sending job to worker %d\n", i)
				jobChannel <- job
				fmt.Printf("Job sent %d by the channel\n", i)
			}
			close(jobChannel)

		}()

		go func() {
			for output := range stdOutput {
				fmt.Print(output)
			}
		}()
		go func() {
			for output := range stdError {
				fmt.Print(output)
			}
		}()

		// Wait for all worker goroutines to finish
		wg.Wait()
		// Close the output channels
		close(stdOutput)
		close(stdError)

		fmt.Println("All workers finished")
	},
}

func init() {
	rootCmd.AddCommand(downloadFilesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	downloadFilesCmd.PersistentFlags().Bool("v2", false, "Donwload only files from version 2")
	downloadFilesCmd.PersistentFlags().Bool("v1", false, "Donwload only files from version 1")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadFilesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
