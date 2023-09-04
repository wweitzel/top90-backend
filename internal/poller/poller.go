package poller

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/reddit"
	"github.com/wweitzel/top90/internal/s3"
	"github.com/wweitzel/top90/internal/scrape"
)

type GoalPoller struct {
	Dao          db.Top90DAO
	S3Client     *s3.S3Client
	RedditClient *reddit.RedditClient
	Scraper      *scrape.Scraper
	Options      Options
	BucketName   string
}

type RunMode int

const (
	Newest RunMode = iota
	SearchBackfill
)

type Options struct {
	DryRun     bool
	RunMode    RunMode
	SearchTerm string
}

func (poller *GoalPoller) Run() {
	switch poller.Options.RunMode {
	case Newest:
		poller.RunNewest(poller.Options)
	case SearchBackfill:
		poller.RunSearchBackfill(poller.Options, poller.Options.SearchTerm)
	default:
		log.Fatalf("Unkown run mode %v", poller.Options.RunMode)
	}
}

func (poller *GoalPoller) RunNewest(options Options) {
	// Get reddit posts that have been submitted since the last run
	newestGoals, err := poller.Dao.GetGoals(db.Pagination{Limit: 1}, db.GetGoalsFilter{})
	// newestGoal, err := poller.Dao.GetNewestGoal()
	if err != nil {
		log.Fatal(err)
	}

	startEpoch := time.Now().AddDate(0, 0, -1).Unix()

	if len(newestGoals) > 0 {
		newestGoal := newestGoals[0]
		startEpoch = newestGoal.RedditPostCreatedAt.Unix()
	}

	posts := poller.RedditClient.GetNewPosts(startEpoch)
	poller.ingestPosts(posts, options)
}

func (poller *GoalPoller) RunSearchBackfill(options Options, searchTerm string) {
	posts := poller.RedditClient.GetPosts(searchTerm)
	poller.ingestPosts(posts, options)
}

func (poller *GoalPoller) RunPremierLeagueBackfill(options Options) {
	teamNames := []string{
		"city",
		"united",
		"chelsea",
		"arsenal",
		"west",
		"ham",
		"wolves",
		"leicester",
		"brighton",
		"brentford",
		"southampton",
		"crystal",
		"palace",
		"newcastle",
		"aston",
		"villa",
		"leeds",
		"everton",
		"burnley",
		"watford",
		"norwich",
	}

	for _, teamName := range teamNames {
		posts := poller.RedditClient.GetPosts(teamName)
		poller.ingestPosts(posts, options)
	}
}

func (poller *GoalPoller) ingestPosts(posts []reddit.RedditPost, options Options) {
	var wg sync.WaitGroup

	wg.Add(len(posts))

	for _, post := range posts {
		// Sleep here prevents getting rate limited during backfills
		time.Sleep(200 * time.Millisecond)
		go poller.ingest(&wg, post)
	}

	wg.Wait()
}

func (poller *GoalPoller) ingest(wg *sync.WaitGroup, post reddit.RedditPost) {
	defer wg.Done()

	log.Println("\nprocessing...", post.Data.Id)

	sourceUrl := poller.getSourceUrl(post)
	log.Println("final source url: ", "[", sourceUrl, "]")

	if sourceUrl == "" {
		return
	}

	// Download the video
	var videoFile *os.File
	if strings.HasSuffix(sourceUrl, ".m3u8") {
		videoFile = downloadBlob("https://juststream.live/", sourceUrl)
	} else {
		videoFile = downloadVideo(sourceUrl)
	}
	defer videoFile.Close()
	defer os.Remove(videoFile.Name())

	// Extract the thumbnail
	randomId := uuid.NewString()
	randomId = strings.Replace(randomId, "-", "", -1)
	thumbnailFilename := fmt.Sprintf("tmp/%s.jpg", randomId)
	defer os.Remove(thumbnailFilename)

	ffmpegPath := os.Getenv("TOP90_FFMPEG_PATH")
	cmd := exec.Command(ffmpegPath, "-i", videoFile.Name(), "-q:v", "2", "-vframes", "1", thumbnailFilename)
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		log.Println("warning: error generating thumbnail with ffpmeg", err)
	}

	log.Println(videoFile.Name(), '\n')

	fi, err := videoFile.Stat()
	if err != nil {
		log.Println("warning: Could not determine file size. This goal will not be stored in the database.")
		return
	}

	fmt.Printf("file size: %d bytes long", fi.Size())

	if fi.Size() < 1 {
		log.Println("warning: Empty file. This goal will not be stored in the database.")
		return
	}

	// Insert goal into db and upload the mp4 file to s3
	redditFullName := post.Kind + "_" + post.Data.Id
	createdAt := convertRedditCreatedAtToTime(post)

	goal := top90.Goal{
		RedditFullname:      redditFullName,
		RedditPostCreatedAt: createdAt,
		RedditPostTitle:     post.Data.Title,
		RedditLinkUrl:       post.Data.URL,
	}

	firstTeamNameFromPost, _ := GetTeamName(post.Data.Title)

	teams, err1 := poller.Dao.GetTeams(db.GetTeamsFilter{})
	if err1 != nil {
		log.Println("error: could not get teams from db")
	}

	team, err2 := GetTeamForTeamName(firstTeamNameFromPost, teams)
	if err2 != nil {
		log.Printf("team name %s could not be found in db", firstTeamNameFromPost)
	}

	// Try to link the fixture
	if err1 == nil && err2 == nil {
		fixtures, _ := poller.Dao.GetFixtures(db.GetFixuresFilter{Date: createdAt})
		fixture, err := GetFixtureForTeamName(firstTeamNameFromPost, team.Aliases, fixtures)

		if err != nil {
			log.Println("warning:", "no fixture for", team.Name, "on date", goal.RedditPostCreatedAt)
		} else {
			goal.FixtureId = fixture.Id
		}
	}

	err = poller.insertAndUpload(goal, videoFile.Name(), thumbnailFilename)
	if err == sql.ErrNoRows {
		log.Printf("Already stored goal for fullname %s", redditFullName)
	} else if err != nil {
		log.Printf("Failed to insert goal for fullname %s due to %v", redditFullName, err)
	}
}

func (poller *GoalPoller) insertAndUpload(goal top90.Goal, videoFilename string, thumbnailFilename string) error {
	videoKey := createKey("mp4")
	goal.S3ObjectKey = videoKey

	thumbnailKey := createKey("jpg")
	goal.ThumbnailS3Key = thumbnailKey

	log.Println("inserting goal...", goal.RedditFullname)
	createdGoal, err := poller.Dao.InsertGoal(&goal)
	if err != nil {
		return err
	}
	log.Println("Successfully saved goal in db", createdGoal.Id, goal.RedditFullname)

	err = poller.S3Client.UploadFile(videoFilename, videoKey, "video/mp4", poller.BucketName)
	if err != nil {
		log.Println("s3 video upload failed", err)
	} else {
		log.Println("Successfully uploaded video to s3", videoKey)
	}

	err = poller.S3Client.UploadFile(thumbnailFilename, thumbnailKey, "image/jpg", poller.BucketName)
	if err != nil {
		log.Println("s3 thumbanil upload failed", err)
	} else {
		log.Println("Successfully uploaded thumbnail to s3", thumbnailKey)
	}

	return nil
}

func (poller *GoalPoller) getSourceUrl(post reddit.RedditPost) string {
	// Get a direct download link (sourceUrl) by crawling
	sourceUrl := poller.Scraper.GetVideoSourceUrl(post.Data.URL)

	// If couldnt get a source url, try to get it from a juststream mirror
	if sourceUrl == "" {
		httpClient := &http.Client{
			Timeout: time.Second * 10,
		}
		redditClient := reddit.NewRedditClient(httpClient)
		// getComments sorting by oldest gives the mirror link at spot [0]
		comments := redditClient.GetComments(post.Data.Id)
		if len(comments) > 0 {
			mirrorsComment := comments[0]
			log.Println("Mirror replies count:", len(mirrorsComment.Data.Replies.Data.Children))

			if len(mirrorsComment.Data.Replies.Data.Children) > 0 {
				replyIds := mirrorsComment.Data.Replies.Data.Children[0].Data.Children
				sourceUrl = poller.Scraper.GetSourceUrlFromMirrors(replyIds)
			}
		} else {
			log.Println("Post had no comments")
		}
	}

	return sourceUrl
}

func convertRedditCreatedAtToTime(post reddit.RedditPost) time.Time {
	unixTimestamp := post.Data.Created_utc
	postCreatedAt := time.Unix(int64(unixTimestamp), 0).UTC()
	return postCreatedAt
}

func createKey(fileExtension string) string {
	nowUtc := time.Now().UTC()
	yearMonthDayStr := fmt.Sprintf("%d-%02d-%02d",
		nowUtc.Year(), nowUtc.Month(), nowUtc.Day())

	id := uuid.NewString()
	id = strings.Replace(id, "-", "", -1)
	objectKey := yearMonthDayStr + "/" + id + "." + fileExtension
	return objectKey
}

func downloadBlob(referrer string, url string) *os.File {
	file, err := ioutil.TempFile("tmp", "*.mp4")
	if err != nil {
		log.Fatalln(err)
	}

	ffmpegPath := os.Getenv("TOP90_FFMPEG_PATH")

	cmd := exec.Command(
		ffmpegPath,
		"-headers", "Referer: "+referrer,
		"-i", url,
		"-y",
		"-c", "copy", file.Name())

	err = cmd.Run()
	if err != nil {
		log.Println(err)
	}

	mp4, err := os.Open(file.Name())
	if err != nil {
		log.Println(err)
	}

	return mp4
}

func downloadVideo(url string) *os.File {
	file, err := ioutil.TempFile("tmp", "*.mp4")
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}

	req.Header.Add("User-Agent", "jawnt")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed GET url: %s due to: %v", url, err)
		return file
	}

	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Println(err)
	}

	return file
}
