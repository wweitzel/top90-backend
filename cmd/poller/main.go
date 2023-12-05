package main

import (
	"context"
	"log"
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
		log.Fatal("Could not setup database: ", err)
	}

	s3Client, err := s3.NewClient(
		config.AwsAccessKey,
		config.AwsSecretAccessKey,
	)
	if err != nil {
		log.Fatal("Failed to create s3 client: ", err)
	}

	err = s3Client.VerifyConnection(config.AwsBucketName)
	if err != nil {
		log.Fatal("Failed to connect to s3 bucket: ", err)
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
	if err := chromedp.Run(ctx); err != nil {
		log.Fatal("Could not setup chromedp: ", err)
	}

	redditClient := reddit.NewClient(reddit.Config{
		Timeout: time.Second * 10,
	})
	dao := dao.NewPostgresDAO(DB)

	scraper := scrape.NewScraper(ctx, dao, redditClient, *s3Client, config.AwsBucketName, logger)
	err = scraper.ScrapeNewPosts()
	if err != nil {
		log.Fatal(err)
	}
}
