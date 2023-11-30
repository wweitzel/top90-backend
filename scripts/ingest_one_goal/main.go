package main

import (
	"log"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/ingest"
	"github.com/wweitzel/top90/internal/reddit"
)

func main() {
	log.Println("Starting Run...")

	config := top90.LoadConfig()
	goalIngest := ingest.NewGoalIngest(config)

	post := reddit.RedditPost{
		Data: struct {
			URL                  string
			Title                string
			Created_utc          float64
			Id                   string
			Link_flair_css_class string
		}{
			URL: "https://dubz.co/v/gkkyv5",
		}}

	var posts []reddit.RedditPost
	posts = append(posts, post)
	goalIngest.IngestPosts(posts)

	log.Println("Finished.")

}
