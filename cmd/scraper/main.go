package main

import (
	"time"

	"github.com/wweitzel/top90/internal/clients/apifootball"
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

	chromeCtx, cancel := init.ChromeDP()
	defer cancel()

	var apifbClient *apifootball.Client
	if config.ApiFootballPlayerLinkEnabled {
		apifbClient = init.ApiFootballClient(
			config.ApiFootballRapidApiHost,
			config.ApiFootballRapidApiKey,
			config.ApiFootballRapidApiKeyBackup,
			10*time.Second,
			config.ApiFootballCurrentSeason)
	}

	scraper := scrape.NewScraper(
		chromeCtx,
		dao,
		redditClient,
		s3Client,
		config.AwsBucketName,
		apifbClient,
		logger)

	err := scraper.ScrapeNewPosts()
	if err != nil {
		init.Exit("Failed scraping new posts", err)
	}
}
