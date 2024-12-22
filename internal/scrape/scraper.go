package scrape

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"time"

	fuzzy "github.com/paul-mannino/go-fuzzywuzzy"
	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/clients/reddit"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/db/dao"
	db "github.com/wweitzel/top90/internal/db/models"
	"github.com/wweitzel/top90/internal/email"
)

type Scraper struct {
	ctx          context.Context
	dao          dao.Top90DAO
	redditClient reddit.Client
	apifbClient  *apifootball.Client
	s3Client     s3.S3Client
	s3Buckent    string
	logger       *slog.Logger
}

func NewScraper(
	ctx context.Context,
	dao dao.Top90DAO,
	redditClient reddit.Client,
	s3Client s3.S3Client,
	s3Bucket string,
	apifbClient *apifootball.Client,
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
		apifbClient:  apifbClient,
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
			s.logger.Debug("Error scraping post", "post", post, "error", err)
		}
	}

	return nil
}

func (s *Scraper) Scrape(p reddit.Post) error {
	if len(p.Data.Title) > 110 {
		s.logger.Debug("Post title is longer than 110 characters")
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

	twoMinutesAgo := time.Now().Add(-2 * time.Minute)
	recentGoals, err := s.dao.GetGoalsSince(twoMinutesAgo)
	if err != nil {
		return fmt.Errorf("failed to get recent goals: %w", err)
	}

	for _, recentGoal := range recentGoals {
		ratio := fuzzy.Ratio(p.Data.Title, recentGoal.RedditPostTitle)
		if ratio > 80 {
			s.logger.Debug("Similar goal already exists", "title", p.Data.Title, "existing_title", recentGoal.RedditPostTitle, "match_ratio", ratio)
			email.Send("[TOP90] [ALERT]", fmt.Sprintf("Similar goal already exists: %s\n\n%s\n\n%d%%", p.Data.Title, recentGoal.RedditPostTitle, ratio))
			return nil
		}
	}

	fixture, err := s.findFixture(p)
	if err != nil {
		return err
	}
	if fixture == nil {
		s.logger.Debug("No fixture found in db", "title", p.Data.Title)
		return nil
	}

	//  1 - World Cup
	//  2 - Champions League
	//  3 - Europa League
	// 39 - Premier League
	// 45 - FA Cup
	// 48 - League Cup
	supportedLeagueIds := []int{1, 2, 3, 39, 45, 48}
	if !slices.Contains(supportedLeagueIds, fixture.LeagueId) {
		s.logger.Debug("Fixture not in supported leagues", "title", p.Data.Title, "leagueId", fixture.LeagueId)
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

	goal := db.Goal{
		RedditFullname:      redditFullName,
		RedditPostCreatedAt: createdTime(p),
		RedditPostTitle:     p.Data.Title,
		RedditLinkUrl:       p.Data.URL,
		FixtureId:           db.NullInt(fixture.Id),
	}

	if s.apifbClient != nil {
		player, event, err := s.linkPlayerWithApiFootball(p.Data.Title, fixture.Id)
		if err != nil {
			s.logger.Warn("Failed linking player with apifootball event")
		}
		goal.PlayerId = db.NullInt(player.Id)
		goal.Type = db.NullString(event.Type)
		goal.TypeDetail = db.NullString(event.Detail)
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
