package bmdatabase

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

var db *sql.DB

func getDbConnection() (*sql.DB, error) {
	if db != nil {
		return db, nil
	}
	cfg := mysql.Config{
		User:                 viper.GetString("MYSQL_USER"),
		Passwd:               viper.GetString("MYSQL_PASSWORD"),
		Net:                  viper.GetString("MYSQL_NET"),
		Addr:                 viper.GetString("MYSQL_ADDR"),
		DBName:               viper.GetString("MYSQL_DBNAME"),
		AllowNativePasswords: true,
	}
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	pingErr := db.Ping()
	if pingErr != nil {
		return nil, pingErr
	}
	return db, nil
}

func GetUser(email string) (BmUser, error) {
	if db == nil {
		_, err := getDbConnection()
		if err != nil {
			return BmUser{}, err
		}
	}

	var user BmUser
	row := db.QueryRow("SELECT idusername, mail, token FROM users WHERE mail = ?", email)
	if err := row.Scan(&user.Idusername, &user.Mail, &user.Token); err != nil {
		if err == sql.ErrNoRows {
			return user, fmt.Errorf("user with email %s not foud", email)
		}
		return user, fmt.Errorf("error in GetUser():: email %s: %w", email, err)
	}
	return user, nil
}

func GetCampaigns(campaignStatusId int64) ([]BmCampaign, error) {
	if db == nil {
		_, err := getDbConnection()
		if err != nil {
			return nil, err
		}
	}

	rows, err := db.Query("SELECT idcampaign, idcampaignstatus, version, IF(advertiser IS NULL, 'no_advertiser', advertiser) as advertiser FROM campaign as c LEFT JOIN advertiser AS a ON c.idadvertiser = a.idadvertiser WHERE idcampaignstatus = ?", campaignStatusId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []BmCampaign
	for rows.Next() {
		var idcampaign int64
		var idcampaignstatus int64
		var versionStr string
		var advertiser string
		err := rows.Scan(&idcampaign, &idcampaignstatus, &versionStr, &advertiser)
		if err != nil {
			return nil, fmt.Errorf("error in campaign id: %d: %w", idcampaign, err)
		}
		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return nil, fmt.Errorf("error in campaign id %d: %w", idcampaign, err)
		}
		campaigns = append(
			campaigns,
			BmCampaign{
				IdCampaign:       idcampaign,
				IdCampaignStatus: idcampaignstatus,
				Version:          version,
				AdvertiserName:   advertiser,
			},
		)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return campaigns, nil
}

func GetCampaignPlannerToken(campaignId int64) (string, error) {
	if db == nil {
		_, err := getDbConnection()
		if err != nil {
			return "", fmt.Errorf("getCampaignPlannerToken()::campaign id: %d: %w", campaignId, err)
		}
	}

	var token string
	row := db.QueryRow("SELECT u.token FROM beeyond.campaign AS c JOIN accounts AS a ON c.idaccount = a.idaccount JOIN users AS u ON a.idusername = u.idusername WHERE c.idcampaign = ?", campaignId)
	if err := row.Scan(&token); err != nil {
		if err == sql.ErrNoRows {
			return token, fmt.Errorf("planner token for campaign with id %d not found", campaignId)
		}
		return token, fmt.Errorf("error in GetCampaignPlannerToken()::campaign id: %d: %w", campaignId, err)
	}
	return token, nil
}
