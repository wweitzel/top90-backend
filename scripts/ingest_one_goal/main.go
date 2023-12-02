package main

import (
	"log"

	"github.com/wweitzel/top90/internal/clients/reddit"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/scrape"
)

func main() {
	log.Println("Starting Run...")

	config := config.Load()
	goalIngest := scrape.NewGoalIngest(config)

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

	var posts []reddit.RedditPost
	posts = append(posts, post)
	goalIngest.IngestPosts(posts)

	log.Println("Finished.")

}
