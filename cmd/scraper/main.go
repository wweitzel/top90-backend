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

	logger.Info("Starting initialization")

	init := cmd.NewInit(logger)

	logger.Info("Initializing S3 client...")
	s3Client := init.S3Client(s3.Config{
		AccessKey:       config.AwsAccessKey,
		SecretAccessKey: config.AwsSecretAccessKey,
		Endpoint:        config.AwsS3Endpoint,
		Logger:          logger,
	}, config.AwsBucketName)

	logger.Info("Done")

	logger.Info("Initializing Reddit client...")
	redditClient := init.RedditClient(10 * time.Second)

	logger.Info("Done")

	logger.Info("Initializing database...")
	db := init.DB(
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbHost,
		config.DbPort)

	logger.Info("Done")

	logger.Info("Initializing dao...")
	dao := init.Dao(db)

	logger.Info("Done")

	logger.Info("Initializing chromdp...")
	chromeCtx, cancel := init.ChromeDP()
	defer cancel()
	logger.Info("Done")

	logger.Info("Initializing apifootball client...")
	var apifbClient *apifootball.Client
	if config.ApiFootballPlayerLinkEnabled {
		apifbClient = init.ApiFootballClient(
			config.ApiFootballRapidApiHost,
			config.ApiFootballRapidApiKey,
			config.ApiFootballRapidApiKeyBackup,
			10*time.Second,
			config.ApiFootballCurrentSeason)
	}

	logger.Info("Done")

	logger.Info("Initializing scraper...")
	scraper := scrape.NewScraper(
		chromeCtx,
		dao,
		redditClient,
		s3Client,
		config.AwsBucketName,
		apifbClient,
		logger)

	logger.Info("Done...")

	err := scraper.ScrapeNewPosts()
	if err != nil {
		init.Exit("Failed scraping new posts", err)
	}
}
