package campaigndownloads

import (
	"campaign-downloads/pkg/bmdatabase"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/kennygrant/sanitize"
	"github.com/spf13/viper"
)

func DownloadFile(campaign *bmdatabase.BmCampaign, adminToken string, plannerToken string) error {

	sb := strings.Builder{}
	sb.WriteString(adminToken)
	// sb.WriteString("-")
	// sb.WriteString(plannerToken)
	token := sb.String()

	env := viper.GetString("GO_ENV")

	var downloadUrl string
	var prodUrl string
	var version string

	if env == "development" {
		downloadUrl = fmt.Sprintf("https://admin-dev.beeyondcloud.com/campaign.export.php?format=D1V&id=%d", campaign.IdCampaign)
		encodedURL := url.QueryEscape(downloadUrl)
		prodUrl = fmt.Sprintf("https://admin-dev.beeyondcloud.com/controller.php?p=login&token=%s&r=%s", token, encodedURL)
		version = "v1"
	} else {

		version = PlatformVersion(campaign)

		if version == "v2" {
			// var downloadUrl = fmt.Sprintf("https://adminv2.beeyondcloud.com/campaign.export.php?format=D1V&id=%d", idCampaign)
			downloadUrl = fmt.Sprintf("https://adminv2.beeyondcloud.com/campaign.export.php?format=D1V&id=%d", campaign.IdCampaign)
			encodedURL := url.QueryEscape(downloadUrl)
			// WARNING: the query param redirect=1 was removed because http.Get() follows the redirection but it does not
			// resend the same cookies, so the request fails.
			prodUrl = fmt.Sprintf("https://adminv2.beeyondcloud.com/controller.php?p=login&token=%s&r=%s", token, encodedURL)
		} else {
			// var downloadUrl = fmt.Sprintf("https://admin.beeyondcloud.com/campaign.export.php?format=D1V&id=%d", idCampaign)
			downloadUrl = fmt.Sprintf("https://admin.beeyondcloud.com/campaign.export.php?format=D1&id=%d", campaign.IdCampaign)
			encodedURL := url.QueryEscape(downloadUrl)
			prodUrl = fmt.Sprintf("https://admin.beeyondcloud.com/controller.php?p=login&token=%s&r=%s", token, encodedURL)
		}
	}

	// Send the first GET request
	resp, err := http.Get(prodUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create a variable with the value of campaign.AdvertiserName but replacing the blank spaces by underscores
	var advertiserName string
	advertiserName = strings.ReplaceAll(campaign.AdvertiserName, " ", "_")
	// Remove any special character from the advertiser name
	advertiserName = sanitize.BaseName(advertiserName)

	// Create the output file name
	fileName := fmt.Sprintf("downloaded-files/%s_campaign_%d-%s.xlsx", advertiserName, campaign.IdCampaign, version)

	// Create a new HTTP client. This is necessary because we need to send a new request with the cookies from the previous response
	client := &http.Client{}

	// Create a new request
	req, err := http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		return err
	}

	// Get the cookies from the previous response
	cookies := resp.Cookies()

	// Set the cookies from the previous response in the new request
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// Send the request using the client
	resp2, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()

	// Check if the request was successful
	if resp2.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading campaign %d ERROR: request failed with status code %d", campaign.IdCampaign, resp.StatusCode)
	}

	// Check if the content type is correct
	contentType := resp2.Header.Get("Content-Type")
	if contentType != "application/vnd.ms-excel; charset=utf-8" {
		return fmt.Errorf("downloading campaign %d ERROR: unexpected content type: %s", campaign.IdCampaign, contentType)
	}

	// Create the output file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the response body to the output file
	_, err = io.Copy(file, resp2.Body)
	if err != nil {
		return fmt.Errorf("downloading campaign %d ERROR: %s", campaign.IdCampaign, err.Error())
	}

	return nil
}
