package scrape

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/db"
)

type Loader struct {
	dao          db.Top90DAO
	s3Client     s3.S3Client
	s3BucketName string
}

// Downloads the video, extracts the thumbnail, stores in db + s3
func (l *Loader) Load(srcUrl string, goal top90.Goal) {
	video := download(srcUrl)
	defer video.Close()
	defer os.Remove(video.Name())

	fi, err := video.Stat()

	log.Printf("file size: %d bytes long", fi.Size())

	if err != nil || fi.Size() < 1 {
		log.Println("warning: Could not determine file size. This goal will not be stored in the database.")
		return
	}

	thumbnail := extractThumbnail(video)
	defer os.Remove(thumbnail)

	err = l.insertAndUpload(goal, video.Name(), thumbnail)
	if err == sql.ErrNoRows {
		log.Printf("Already stored goal for fullname %s", goal.RedditFullname)
	} else if err != nil {
		log.Printf("Failed to insert goal for fullname %s due to %v", goal.RedditFullname, err)
	}
}

func (l *Loader) insertAndUpload(goal top90.Goal, videoFilename string, thumbnailFilename string) error {
	videoKey := createKey("mp4")
	goal.S3ObjectKey = videoKey

	thumbnailKey := createKey("avif")
	goal.ThumbnailS3Key = thumbnailKey

	log.Println("inserting goal...", goal.RedditFullname)
	createdGoal, err := l.dao.InsertGoal(&goal)
	if err != nil {
		return err
	}
	log.Println("Successfully saved goal in db", createdGoal.Id, goal.RedditFullname)

	err = l.s3Client.UploadFile(videoFilename, videoKey, "video/mp4", l.s3BucketName)
	if err != nil {
		log.Println("s3 video upload failed", err)
	} else {
		log.Println("Successfully uploaded video to s3", videoKey)
	}

	err = l.s3Client.UploadFile(thumbnailFilename, thumbnailKey, "image/avif", l.s3BucketName)
	if err != nil {
		log.Println("s3 thumbanil upload failed", err)
	} else {
		log.Println("Successfully uploaded thumbnail to s3", thumbnailKey)
	}

	return nil
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
