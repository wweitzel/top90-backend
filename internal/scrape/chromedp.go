package scrape

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type chromeDpScraper struct {
	logger *slog.Logger
}

func NewChromDpScraper(logger *slog.Logger) chromeDpScraper {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return chromeDpScraper{logger: logger}
}

func (s chromeDpScraper) getVideoSourceUrl(ctx context.Context, url string) string {
	newTabCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	newTabCtx, cancel = context.WithTimeout(newTabCtx, 15*time.Second)
	defer cancel()

	s.logger.Debug("New tab", "url", url)

	var videoNodes []*cdp.Node
	var sourceNodes []*cdp.Node

	err := chromedp.Run(newTabCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.WaitVisible(`video`, chromedp.ByQuery),
		chromedp.ActionFunc(func(context.Context) error {
			s.logger.Debug("Video is visible in DOM")
			return nil
		}),
		chromedp.Nodes(`video`, &videoNodes, chromedp.AtLeast(0)),
		chromedp.Nodes(`source`, &sourceNodes, chromedp.AtLeast(0)),
	)

	if err != nil {
		s.logger.Debug("ChromeDP timed out", "url", url, "err", err)
		return ""
	}

	sourceUrl := getSource(videoNodes, url)
	if len(sourceUrl) > 0 {
		return sourceUrl
	}

	sourceUrl = getSource(sourceNodes, url)
	return sourceUrl
}

// Uses chromedp to scan network for xhr ending in m3u8 and returns that url
func (s chromeDpScraper) getVideoSourceNetwork(ctx context.Context, url string) (string, error) {
	newTabCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	newTabCtx, cancel = context.WithTimeout(newTabCtx, 3*time.Second)
	defer cancel()

	s.logger.Debug("New tab for juststream processing", "url", url)

	var sourceUrl string
	chromedp.ListenTarget(newTabCtx, func(ev interface{}) {
		if ev, ok := ev.(*network.EventResponseReceived); ok {
			if ev.Type != "XHR" {
				return
			}

			if strings.HasSuffix(ev.Response.URL, "video.m3u8") {
				s.logger.Debug("XHR event had .m3u8 ending", "url", ev.Response.URL)
				sourceUrl = ev.Response.URL
				return
			}
		}
	})

	err := chromedp.Run(newTabCtx,
		network.Enable(),
		chromedp.Navigate(url),
	)
	if err != nil {
		return "", err
	}

	return sourceUrl, nil
}

func getSource(nodes []*cdp.Node, url string) string {
	for _, node := range nodes {
		sourceUrl := node.AttributeValue("src")

		id := node.AttributeValue("id")
		if (strings.HasPrefix(url, "https://streamin.one") || strings.HasPrefix(url, "https://streamin.me")) && id == "video" {
			return sourceUrl
		}

		// Sometimes streamff sources are a relative path
		if strings.HasPrefix(url, "https://streamff") && strings.HasPrefix(sourceUrl, "/uploads") {
			sourceUrl = "https://streamff.com" + sourceUrl
		}

		if len(sourceUrl) > 0 &&
			!strings.HasSuffix(sourceUrl, ".js") &&
			!strings.HasPrefix(sourceUrl, "https://ad.plus") {
			return sourceUrl
		}
	}
	return ""
}
