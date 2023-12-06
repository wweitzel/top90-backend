package main

import (
	"context"
	"log/slog"
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

var logger = jsonlogger.New(&jsonlogger.Options{
	Level:    slog.LevelDebug,
	Colorize: true,
})

func main() {
	config := config.Load()

	DB, err := db.NewPostgresDB(config.DbUser, config.DbPassword, config.DbName, config.DbHost, config.DbPort)
	if err != nil {
		exit("Failed setting database", err)
	}

	s3Client, err := s3.NewClient(s3.Config{
		AccessKey:       config.AwsAccessKey,
		SecretAccessKey: config.AwsSecretAccessKey,
		Endpoint:        config.AwsS3Endpoint,
	})
	if err != nil {
		exit("Failed creating s3 client", err)
	}

	err = s3Client.VerifyConnection(config.AwsBucketName)
	if err != nil {
		exit("Failed connecting to s3 bucket", err)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3830.0 Safari/537.36"),
	)
	ctx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ = chromedp.NewContext(ctx)
	if err := chromedp.Run(ctx); err != nil {
		exit("Failed setting up chromedp", err)
	}

	redditClient, err := reddit.NewClient(reddit.Config{Timeout: time.Second * 10})
	if err != nil {
		exit("Failed creating reddit client", err)
	}

	dao := dao.NewPostgresDAO(DB)
	scraper := scrape.NewScraper(ctx, dao, *redditClient, *s3Client, config.AwsBucketName, logger)

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

func exit(msg string, err error) {
	logger.Error(msg, "error", err)
	os.Exit(1)
}
