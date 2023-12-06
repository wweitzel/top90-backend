package main

import (
	"time"

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

	s3Client := init.S3Client(
		s3.Config{
			AccessKey:       config.AwsAccessKey,
			SecretAccessKey: config.AwsSecretAccessKey,
			Endpoint:        config.AwsS3Endpoint,
			Logger:          logger,
		},
		config.AwsBucketName)

	redditClient := init.RedditClient(10 * time.Second)

	db := init.DB(
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbHost,
		config.DbPort)

	dao := init.Dao(db)

	chromeCtx := init.ChromeDP()

	scraper := scrape.NewScraper(
		chromeCtx,
		dao,
		redditClient,
		s3Client,
		config.AwsBucketName,
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
			Created_utc: 1701809991,
			Title:       `Luton 1 - [2] Arsenal - Gabriel Jesus 45'`,
		}}

	scraper.Scrape(post)
}
