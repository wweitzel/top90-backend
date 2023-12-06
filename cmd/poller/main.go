package main

import (
	"context"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/wweitzel/top90/internal/clients/reddit"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/db/postgres/dao"
	"github.com/wweitzel/top90/internal/jsonlogger"
	"github.com/wweitzel/top90/internal/scrape"
)

func main() {
	config := config.Load()

	logger := jsonlogger.New(&jsonlogger.Options{
		Level:    config.LogLevel,
		Colorize: config.LogColor,
	})

	DB, err := db.NewPostgresDB(
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbHost,
		config.DbPort,
	)
	if err != nil {
		logger.Error("Failed connecting to database", "error", err)
		os.Exit(1)
	}

	s3Client, err := s3.NewClient(s3.Config{
		AccessKey:       config.AwsAccessKey,
		SecretAccessKey: config.AwsSecretAccessKey,
		Endpoint:        config.AwsS3Endpoint,
	})
	if err != nil {
		logger.Error("Failed creating s3 client", "error", err)
		os.Exit(1)
	}

	err = s3Client.VerifyConnection(config.AwsBucketName)
	if err != nil {
		logger.Error("Failed connecting to s3 bucket", "error", err)
		os.Exit(1)
	}

	const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/77.0.3830.0 Safari/537.36"

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(userAgent),
	)
	ctx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ = chromedp.NewContext(ctx)
	err = chromedp.Run(ctx)
	if err != nil {
		logger.Error("Failed initializing chromedp", "error", err)
		os.Exit(1)
	}

	redditClient, err := reddit.NewClient(reddit.Config{
		Timeout: time.Second * 10,
		Logger:  logger,
	})
	if err != nil {
		logger.Error("Failed creating reddit client", "error", err)
		os.Exit(1)
	}

	dao := dao.NewPostgresDAO(DB)

	scraper := scrape.NewScraper(ctx, dao, *redditClient, *s3Client, config.AwsBucketName, logger)
	err = scraper.ScrapeNewPosts()
	if err != nil {
		logger.Error("Failed scraping new posts", "error", err)
		os.Exit(1)
	}
}
