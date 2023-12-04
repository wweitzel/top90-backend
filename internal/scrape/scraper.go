package scrape

import (
	"context"
	"log"
	"strings"
	"time"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/clients/reddit"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/db"
)

type Scraper struct {
	ctx           context.Context
	dao           db.Top90DAO
	redditClient  reddit.Client
	s3Client      s3.S3Client
	s3BuckentName string
}

func NewScraper(
	ctx context.Context,
	dao db.Top90DAO,
	redditClient reddit.Client,
	s3Client s3.S3Client,
	s3BucketName string,
) Scraper {

	return Scraper{
		ctx:           ctx,
		dao:           dao,
		redditClient:  redditClient,
		s3Client:      s3Client,
		s3BuckentName: s3BucketName,
	}
}

func (s *Scraper) ScrapeNewPosts() error {
	posts, err := s.redditClient.GetNewPosts()
	if err != nil {
		return err
	}

	for _, post := range posts {
		log.Println("\nprocessing...", post.Data.Id)
		s.Scrape(post)
	}

	return nil
}

func (s *Scraper) Scrape(p reddit.Post) {
	if len(p.Data.Title) > 110 {
		log.Println("error: post title does not look like the title of a goal post.")
		return
	}

	redditFullName := p.Kind + "_" + p.Data.Id

	goalExists, err := s.dao.GoalExists(redditFullName)
	if err != nil {
		log.Println("warning:", "could not check if goal exists", err)
		return
	}
	if goalExists {
		log.Println("warning:", "goal already exists", p.Data.Title)
		return
	}

	fixture, err := s.findFixture(p)
	if err != nil {
		log.Println("warning:", "no fixture found in db for", p.Data.Title)
		return
	}

	sourceUrl := s.findVideoSourceUrl(p)
	log.Println("final source url: ", "[", sourceUrl, "]")
	if sourceUrl == "" {
		return
	}

	createdAt := createdTime(p)

	goal := top90.Goal{
		RedditFullname:      redditFullName,
		RedditPostCreatedAt: createdAt,
		RedditPostTitle:     p.Data.Title,
		RedditLinkUrl:       p.Data.URL,
		FixtureId:           fixture.Id,
	}

	loader := Loader{
		dao:          s.dao,
		s3Client:     s.s3Client,
		s3BucketName: s.s3BuckentName,
	}

	loader.Load(sourceUrl, goal)
}

func (s *Scraper) findVideoSourceUrl(p reddit.Post) string {
	sourceUrl := collyscraper{}.getVideoSourceUrl(p.Data.URL)

	if len(sourceUrl) == 0 {
		sourceUrl = chromeDpScraper{}.getVideoSourceUrl(s.ctx, p.Data.URL)
	}

	if strings.HasPrefix(sourceUrl, "blob") && strings.Contains(p.Data.URL, "juststream") {
		sourceUrl = chromeDpScraper{}.getVideoSourceNetwork(s.ctx, p.Data.URL)
	} else if strings.HasPrefix(sourceUrl, "blob") {
		// TODO: Need a way to download from any blob, not just juststream
		//   For now, just set to empty string since we cant handle other blobs
		sourceUrl = ""
	}

	return sourceUrl
}

func createdTime(p reddit.Post) time.Time {
	unixTimestamp := p.Data.Created_utc
	return time.Unix(int64(unixTimestamp), 0).UTC()
}
