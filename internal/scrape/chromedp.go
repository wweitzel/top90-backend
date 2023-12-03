package scrape

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type chromeDpScraper struct{}

func (chromeDpScraper) getVideoSourceUrl(ctx context.Context, url string) string {
	var sourceUrl string

	newTabCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	newTabCtx, cancel = context.WithTimeout(newTabCtx, 15*time.Second)
	defer cancel()

	log.Printf("New tab: %s", url)

	var videoNodes []*cdp.Node
	var sourceNodes []*cdp.Node

	err := chromedp.Run(newTabCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.WaitVisible(`video`, chromedp.ByQuery),
		chromedp.ActionFunc(func(context.Context) error {
			log.Printf(">>>>>>>>>>>>>>>>>>>> video IS VISIBLE")
			return nil
		}),
		chromedp.Nodes(`video`, &videoNodes, chromedp.AtLeast(0)),
		chromedp.Nodes(`source`, &sourceNodes, chromedp.AtLeast(0)),
	)
	if err != nil {
		log.Printf("%v %s", err, url)
	} else {
		log.Println("chromedp did NOT timeout!")
	}

	if len(videoNodes) > 0 {
		for _, videoNode := range videoNodes {
			sourceUrl = getSource(videoNode, url)
			if len(sourceUrl) > 0 && !strings.HasSuffix(sourceUrl, ".js") {
				return sourceUrl
			}
		}
	}

	if len(sourceNodes) > 0 {
		for _, sourceNode := range sourceNodes {
			sourceUrl = getSource(sourceNode, url)
			if len(sourceUrl) > 0 && !strings.HasSuffix(sourceUrl, ".js") {
				return sourceUrl
			}
		}
	}

	return ""
}

// Uses chromedp to scan network for xhr ending in m3u8 and returns that url
func (chromeDpScraper) getVideoSourceNetwork(ctx context.Context, url string) string {
	newTabCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	newTabCtx, cancel = context.WithTimeout(newTabCtx, 3*time.Second)
	defer cancel()

	log.Printf("New tab for juststream processing: %s", url)

	var sourceUrl string

	chromedp.ListenTarget(newTabCtx, func(ev interface{}) {
		if ev, ok := ev.(*network.EventResponseReceived); ok {
			if ev.Type != "XHR" {
				return
			}

			if strings.HasSuffix(ev.Response.URL, "video.m3u8") {
				log.Println("XHR event had .m3u8 ending:", ev.Response.URL)
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
		log.Printf("%v %s", err, url)
	}

	return sourceUrl
}

func getSource(node *cdp.Node, url string) string {
	sourceUrl := node.AttributeValue("src")
	// Sometimes streamff sources are a relative path
	if strings.HasPrefix(url, "https://streamff") && strings.HasPrefix(sourceUrl, "/uploads") {
		sourceUrl = "https://streamff.com" + sourceUrl
	}
	if len(sourceUrl) > 0 {
		return sourceUrl
	}

	return ""
}
