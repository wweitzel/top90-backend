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

	// Setup a headless chrome browser context
	ctx, cancel := chromedp.NewContext(context.Background())
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

	goalPoller.Run()

	log.Println("Finished.")
}
