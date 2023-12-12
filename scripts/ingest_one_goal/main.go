package main

import (
	"time"

	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/clients/reddit"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/cmd"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/jsonlogger"
	"github.com/wweitzel/top90/internal/scrape"
)

func main() {
	config := config.Load()
	logger := jsonlogger.New(&jsonlogger.Options{
		Level:    config.LogLevel,
		Colorize: config.LogColor,
	})

	init := cmd.NewInit(logger)

	s3Client := init.S3Client(s3.Config{
		AccessKey:       config.AwsAccessKey,
		SecretAccessKey: config.AwsSecretAccessKey,
		Endpoint:        config.AwsS3Endpoint,
		Logger:          logger,
	}, config.AwsBucketName)

	redditClient := init.RedditClient(10 * time.Second)

	db := init.DB(
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbHost,
		config.DbPort)

	dao := init.Dao(db)

	chromeCtx := init.ChromeDP()

	var apifbClient *apifootball.Client
	if config.ApiFootballPlayerLinkEnabled {
		apifbClient = init.ApiFootballClient(
			config.ApiFootballRapidApiHost,
			config.ApiFootballRapidApiKey,
			config.ApiFootballRapidApiKeyBackup,
			10*time.Second,
			config.ApiFootballCurrentSeason,
		)
	}

	scraper := scrape.NewScraper(
		chromeCtx,
		dao,
		redditClient,
		s3Client,
		config.AwsBucketName,
		apifbClient,
		logger)

	post := reddit.Post{
		Data: struct {
			URL                  string
			Title                string
			Created_utc          float64
			Id                   string
			Link_flair_css_class string
		}{
			URL:         "https://dubz.co/v/ca05c6",
			Created_utc: float64(time.Now().Unix()),
			Title:       `Empoli 1 - [2] Arsenal - Gabriel Jesus 32'`,
		}}

	scraper.Scrape(post)
}
