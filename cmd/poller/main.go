package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/poller"
	"github.com/wweitzel/top90/internal/reddit"
	"github.com/wweitzel/top90/internal/s3"
	"github.com/wweitzel/top90/internal/scrape"
)

func main() {
	log.Println("Starting Run...")

	// Load config from .env into environment variables
	config := top90.LoadConfig()

	// Connect to database
	DB, err := db.NewPostgresDB(config.DbUser, config.DbPassword, config.DbName, config.DbHost, config.DbPort)
	if err != nil {
		log.Fatalf("Could not setup database: %v", err)
	}
	defer DB.Close()

	// Connect to s3
	s3Client := s3.NewClient(config.AwsAccessKey, config.AwsSecretAccessKey)
	err = s3Client.VerifyConnection(config.AwsBucketName)
	if err != nil {
		log.Fatalln("Failed to connect to s3 bucket", err)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3830.0 Safari/537.36"),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Setup a headless chrome browser context
	ctx, cancel = chromedp.NewContext(ctx)
	if err := chromedp.Run(ctx); err != nil {
		log.Fatalf("Coult not setup chromedp: %v", err)
	}
	defer cancel()

	redditClient := reddit.NewRedditClient(&http.Client{Timeout: time.Second * 10})
	scraper := scrape.Scraper{BrowserContext: ctx}
	dao := db.NewPostgresDAO(DB)

	// Initialize poller with options
	goalPoller := poller.GoalPoller{
		RedditClient: &redditClient,
		Scraper:      &scraper,
		Dao:          dao,
		S3Client:     &s3Client,
		BucketName:   config.AwsBucketName,
		// TODO: Take these from command line input
		Options: poller.Options{
			DryRun:     false,
			RunMode:    poller.Newest,
			SearchTerm: "tottenham",
		},
	}

	// https://streamin.me/v/317cdeaf
	// https://dubz.co/c/dd6e07
	// https://dubz.co/c/5c71b0
	// https://streamin.me/v/016f60d8
	// goalPoller.IngestPost("https://dubz.co/c/0e7b61")

	goalPoller.Run()

	log.Println("Finished.")
}
