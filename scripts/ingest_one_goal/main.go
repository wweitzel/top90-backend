package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/wweitzel/top90/internal/clients/reddit"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/scrape"
)

func main() {
	log.Println("Starting Run...")

	config := config.Load()

	DB, err := db.NewPostgresDB(config.DbUser, config.DbPassword, config.DbName, config.DbHost, config.DbPort)
	if err != nil {
		log.Fatalf("Could not setup database: %v", err)
	}

	s3Client := s3.NewClient(config.AwsAccessKey, config.AwsSecretAccessKey)
	err = s3Client.VerifyConnection(config.AwsBucketName)
	if err != nil {
		log.Fatalln("Failed to connect to s3 bucket", err)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3830.0 Safari/537.36"),
	)
	ctx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ = chromedp.NewContext(ctx)
	if err := chromedp.Run(ctx); err != nil {
		log.Fatalf("Coult not setup chromedp: %v", err)
	}

	redditClient := reddit.NewRedditClient(&http.Client{Timeout: time.Second * 10})
	dao := db.NewPostgresDAO(DB)

	scraper := scrape.NewScraper(ctx, dao, redditClient, s3Client, config.AwsBucketName)

	post := reddit.RedditPost{
		Data: struct {
			URL                  string
			Title                string
			Created_utc          float64
			Id                   string
			Link_flair_css_class string
		}{
			URL: "https://www.ole.com.ar/futbol-primera/colon-vs-gimansia-hoy-vivo-desempate-descenso-categoria-hora-ver-resumen_0_BIf8pOsDCf.html",
		}}

	scraper.Scrape(post)

	log.Println("Finished.")

}
