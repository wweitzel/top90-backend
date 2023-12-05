package scrape

import (
	"io"
	"log/slog"
	"strings"

	"github.com/gocolly/colly"
)

type collyScraper struct {
	logger *slog.Logger
}

func NewCollyScraper(logger *slog.Logger) collyScraper {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return collyScraper{logger: logger}
}

func (collyScraper) getVideoSourceUrl(url string) string {
	c := colly.NewCollector()

	var sourceUrl string
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
