package scrape

import (
	"strings"

	"github.com/gocolly/colly"
)

type collyscraper struct{}

func (collyscraper) getVideoSourceUrl(url string) string {
	var sourceUrl string

	c := colly.NewCollector()

	switch {
	case strings.HasPrefix(url, "https://streamable"):
		c.OnHTML("video", func(e *colly.HTMLElement) {
			sourceUrl = e.Attr(("src"))
			sourceUrl = "https:" + sourceUrl
		})
	case strings.HasPrefix(url, "https://www.clippit"):
		c.OnHTML("div[data-hd-file]", func(e *colly.HTMLElement) {
			sourceUrl = e.Attr(("data-hd-file"))
		})
	default:
		c.OnHTML("video source", func(e *colly.HTMLElement) {
			sourceUrl = e.Attr(("src"))
		})
	}

	c.Visit(url)

	return sourceUrl
}
