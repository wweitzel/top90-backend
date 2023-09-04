package scrape

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"github.com/wweitzel/top90/internal/reddit"
)

type Scraper struct {
	BrowserContext context.Context
}

// Attampts to use Colly retrieve the url of a direct download link to the video.
// If that fails, will try to use chromedp.
func (scraper *Scraper) GetVideoSourceUrl(url string) string {
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

	// If colly could not get the url, try using chromedp
	if len(sourceUrl) == 0 {
		sourceUrl = getVideoSourceChromeDp(scraper.BrowserContext, url)
	}

	// If video source url is a blob, have to go back and get the real source
	// using the network tab
	if strings.HasPrefix(sourceUrl, "blob") && strings.Contains(url, "juststream") {
		sourceUrl = getVideoSourceChromeDpNetwork(scraper.BrowserContext, url)
	} else if strings.HasPrefix(sourceUrl, "blob") {
		// TODO: Need a way to download from any blob, not just juststream
		//   For now, just set to empty string since we cant handle other blobs
		sourceUrl = ""
	}

	return sourceUrl
}

// Loops theough the replies and returns immediately when a video source
// could successfully be extracted from one
func (scraper *Scraper) GetSourceUrlFromMirrors(replyIds []string) string {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	redditClient := reddit.NewRedditClient(httpClient)

	var sourceUrl string
	for _, replyId := range replyIds {
		reply := redditClient.GetComment(replyId)
		body := reply.Data.Body
		url := getUrlFromBody(body)

		if url == "" {
			continue
		}

		sourceUrl = scraper.GetVideoSourceUrl(url)

		if sourceUrl != "" {
			log.Println("Found source from mirror: ", url)
			break
		}
	}

	return sourceUrl
}

// Uses chrome dp to load page with javascript and get the source url
func getVideoSourceChromeDp(ctx context.Context, url string) string {
	var sourceUrl string

	newTabCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	newTabCtx, cancel = context.WithTimeout(newTabCtx, 30*time.Second)
	defer cancel()

	log.Printf("New tab: %s", url)

	var videoNodes []*cdp.Node

	err := chromedp.Run(newTabCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.WaitVisible(`video`, chromedp.ByQuery),
		chromedp.ActionFunc(func(context.Context) error {
			log.Printf(">>>>>>>>>>>>>>>>>>>> video IS VISIBLE")
			return nil
		}),
		chromedp.Nodes(`source`, &videoNodes, chromedp.ByQuery),
	)
	if err != nil {
		log.Printf("%v %s", err, url)
	} else {
		log.Println("chromedp did NOT timeout!")
	}

	for _, videoNode := range videoNodes {
		sourceUrl = videoNode.AttributeValue("src")
		// Sometimes streamff sources are a relative path
		if strings.HasPrefix(url, "https://streamff") && strings.HasPrefix(sourceUrl, "/uploads") {
			sourceUrl = "https://streamff.com" + sourceUrl
		}
		if len(sourceUrl) > 0 {
			return sourceUrl
		}
	}

	return ""
}

// Uses chromedp to scan network for xhr ending in m3u8 and returns that url
func getVideoSourceChromeDpNetwork(ctx context.Context, url string) string {
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

func getUrlFromBody(body string) string {
	// Body format is [Juststream Mirror](https://juststream.live/DefyForgingsVain)
	// So slice from the '(' to the ')'
	startIndex := strings.IndexByte(body, '(') + 1
	endIndex := strings.IndexByte(body, ')')

	if startIndex < 0 || endIndex < 0 {
		return ""
	}

	return body[startIndex:endIndex]
}
