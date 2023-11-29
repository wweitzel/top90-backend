package ingest

import (
	"context"
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

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/reddit"
	"github.com/wweitzel/top90/internal/s3"
	"github.com/wweitzel/top90/internal/scrape"
)

type GoalIngest struct {
	dao           db.Top90DAO
	s3client      *s3.S3Client
	redditclient  *reddit.RedditClient
	scraper       *scrape.Scraper
	bucketName    string
	db            *sql.DB
	execCancel    context.CancelFunc
	contextCancel context.CancelFunc
}

func NewGoalIngest(config top90.Config) GoalIngest {
	DB, err := db.NewPostgresDB(config.DbUser, config.DbPassword, config.DbName, config.DbHost, config.DbPort)
	if err != nil {
		log.Fatalf("Could not setup database: %v", err)
	}

	s3Client := s3.NewClient(config.AwsAccessKey, config.AwsSecretAccessKey)
	err = s3Client.VerifyConnection(config.AwsBucketName)
	if err != nil {
		log.Fatalln("Failed to connect to s3 bucket", err)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3830.0 Safari/537.36"),
	)

	ctx, execCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	ctx, contextCancel := chromedp.NewContext(ctx)
	if err := chromedp.Run(ctx); err != nil {
		log.Fatalf("Coult not setup chromedp: %v", err)
	}

	redditClient := reddit.NewRedditClient(&http.Client{Timeout: time.Second * 10})
	scraper := scrape.Scraper{BrowserContext: ctx}
	dao := db.NewPostgresDAO(DB)

	return GoalIngest{
		redditclient:  &redditClient,
		scraper:       &scraper,
		dao:           dao,
		s3client:      &s3Client,
		bucketName:    config.AwsBucketName,
		db:            DB,
		execCancel:    execCancel,
		contextCancel: contextCancel,
	}
}

func (poller *GoalIngest) Run() {
	posts := poller.getNewRedditPosts()
	poller.ingestPosts(posts)
	poller.db.Close()
	poller.execCancel()
	poller.contextCancel()
}

func (poller *GoalIngest) getNewRedditPosts() []reddit.RedditPost {
	// Get reddit posts that have been submitted since the last run
	newestGoal, err := poller.dao.GetNewestGoal()
	if err != nil {
		log.Fatal(err)
	}

	startEpoch := time.Now().AddDate(0, 0, -1).Unix()

	if newestGoal.Id != "" {
		startEpoch = newestGoal.RedditPostCreatedAt.Unix()
	}

	return poller.redditclient.GetNewPosts(startEpoch)
}

func (poller *GoalIngest) ingestPosts(posts []reddit.RedditPost) {
	var wg sync.WaitGroup

	wg.Add(len(posts))

	for _, post := range posts {
		// Sleep here prevents getting rate limited
		time.Sleep(200 * time.Millisecond)
		go poller.ingest(&wg, post)
	}

	wg.Wait()
}

func (poller *GoalIngest) ingest(wg *sync.WaitGroup, post reddit.RedditPost) {
	defer wg.Done()
	log.Println("\nprocessing...", post.Data.Id)

	if len(post.Data.Title) > 110 {
		log.Println("skipping processing. post title does not look like the title of a goal post.")
		return
	}

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
	thumbnailFilename := fmt.Sprintf("tmp/%s.avif", randomId)
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

	allTeams, err1 := poller.dao.GetTeams(db.GetTeamsFilter{})
	if err1 != nil {
		log.Println("error: could not get teams from db")
	}

	// Try to link the fixture
	if err1 == nil {
		fixtures, _ := poller.dao.GetFixtures(db.GetFixuresFilter{Date: createdAt})
		fixture, err := FindFixture(post.Data.Title, allTeams, fixtures)

		if err != nil {
			log.Println("warning:", "no fixture for", post.Data.Title, "on date", goal.RedditPostCreatedAt)
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

func (poller *GoalIngest) insertAndUpload(goal top90.Goal, videoFilename string, thumbnailFilename string) error {
	videoKey := createKey("mp4")
	goal.S3ObjectKey = videoKey

	thumbnailKey := createKey("avif")
	goal.ThumbnailS3Key = thumbnailKey

	log.Println("inserting goal...", goal.RedditFullname)
	createdGoal, err := poller.dao.InsertGoal(&goal)
	if err != nil {
		return err
	}
	log.Println("Successfully saved goal in db", createdGoal.Id, goal.RedditFullname)

	err = poller.s3client.UploadFile(videoFilename, videoKey, "video/mp4", poller.bucketName)
	if err != nil {
		log.Println("s3 video upload failed", err)
	} else {
		log.Println("Successfully uploaded video to s3", videoKey)
	}

	err = poller.s3client.UploadFile(thumbnailFilename, thumbnailKey, "image/avif", poller.bucketName)
	if err != nil {
		log.Println("s3 thumbanil upload failed", err)
	} else {
		log.Println("Successfully uploaded thumbnail to s3", thumbnailKey)
	}

	return nil
}

func (poller *GoalIngest) getSourceUrl(post reddit.RedditPost) string {
	// Get a direct download link (sourceUrl) by crawling
	sourceUrl := poller.scraper.GetVideoSourceUrl(post.Data.URL)

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
				sourceUrl = poller.scraper.GetVideoSourceUrlFromMirrors(replyIds)
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
