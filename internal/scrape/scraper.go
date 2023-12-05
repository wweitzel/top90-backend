package scrape

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"time"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/clients/reddit"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/db"
)

type Scraper struct {
	ctx          context.Context
	dao          db.Top90DAO
	redditClient reddit.Client
	s3Client     s3.S3Client
	s3Buckent    string
	logger       *slog.Logger
}

func NewScraper(
	ctx context.Context,
	dao db.Top90DAO,
	redditClient reddit.Client,
	s3Client s3.S3Client,
	s3Bucket string,
	logger *slog.Logger,
) Scraper {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return Scraper{
		ctx:          ctx,
		dao:          dao,
		redditClient: redditClient,
		s3Client:     s3Client,
		s3Buckent:    s3Bucket,
		logger:       logger,
	}
}

func (s *Scraper) ScrapeNewPosts() error {
	posts, err := s.redditClient.GetNewPosts()
	if err != nil {
		return err
	}

	for _, post := range posts {
		s.logger.Debug("Processing... ", "title", post.Data.Title)
		err := s.Scrape(post)
		if err != nil {
			s.logger.Debug("Error scraping post", "post", post)
		}
	}

	return nil
}

func (s *Scraper) Scrape(p reddit.Post) error {
	if len(p.Data.Title) > 110 {
		s.logger.Debug("Post title is lnoger than 110 characters")
		return nil
	}

	redditFullName := p.Kind + "_" + p.Data.Id

	goalExists, err := s.dao.GoalExists(redditFullName)
	if err != nil {
		return err
	}
	if goalExists {
		s.logger.Debug("Goal already exists", "title", p.Data.Title)
		return nil
	}

	fixture, err := s.findFixture(p)
	if err != nil {
		return err
	}
	if fixture == nil {
		s.logger.Debug("No fixture found in db", "title", p.Data.Title)
		return nil
	}

	sourceUrl, err := s.findVideoSourceUrl(p)
	if err != nil {
		return err
	}

	s.logger.Debug("Final source url: [" + sourceUrl + "]")
	if sourceUrl == "" {
		return nil
	}

	createdAt := createdTime(p)

	goal := top90.Goal{
		RedditFullname:      redditFullName,
		RedditPostCreatedAt: createdAt,
		RedditPostTitle:     p.Data.Title,
		RedditLinkUrl:       p.Data.URL,
		FixtureId:           fixture.Id,
	}

	loader := NewLoader(
		s.dao,
		s.s3Client,
		s.s3Buckent,
		s.logger,
	)

	err = loader.Load(sourceUrl, goal)
	if err != nil {
		return err
	}

	s.logger.Debug("Successfully loaded goal into db", "title", p.Data.Title)
	return nil
}

func (s *Scraper) findVideoSourceUrl(p reddit.Post) (string, error) {
	collyScraper := NewCollyScraper(s.logger)
	sourceUrl := collyScraper.getVideoSourceUrl(p.Data.URL)

	chromeDpScraper := NewChromDpScraper(s.logger)
	if len(sourceUrl) == 0 {
		sourceUrl = chromeDpScraper.getVideoSourceUrl(s.ctx, p.Data.URL)
	}

	if strings.HasPrefix(sourceUrl, "blob") && strings.Contains(p.Data.URL, "juststream") {
		var err error
		sourceUrl, err = chromeDpScraper.getVideoSourceNetwork(s.ctx, p.Data.URL)
		if err != nil {
			return "", err
		}
	} else if strings.HasPrefix(sourceUrl, "blob") {
		// TODO: Need a way to download from any blob, not just juststream
		//   For now, just set to empty string since we cant handle other blobs
		sourceUrl = ""
	}
	return sourceUrl, nil
}

func createdTime(p reddit.Post) time.Time {
	unixTimestamp := p.Data.Created_utc
	return time.Unix(int64(unixTimestamp), 0).UTC()
}
