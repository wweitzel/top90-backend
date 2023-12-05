package scrape

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/db"
)

type loader struct {
	dao      db.Top90DAO
	s3Client s3.S3Client
	s3Bucket string
	logger   *slog.Logger
}

func NewLoader(dao db.Top90DAO, s3Client s3.S3Client, s3Bucket string, logger *slog.Logger) loader {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return loader{
		dao:      dao,
		s3Client: s3Client,
		s3Bucket: s3Bucket,
		logger:   logger,
	}
}

// Downloads the video, extracts the thumbnail, stores in db + s3
func (l *loader) Load(srcUrl string, goal top90.Goal) error {
	video, err := download(srcUrl)
	if err != nil {
		return err
	}
	defer video.Close()
	defer os.Remove(video.Name())

	fi, err := video.Stat()
	if err != nil {
		return err
	}
	if fi.Size() < 1 {
		return fmt.Errorf("empty video file")
	}

	thumbnail, err := extractThumbnail(video)
	if err != nil {
		return fmt.Errorf("error extracting thumbnail: %v", err)
	}
	defer os.Remove(thumbnail)

	err = l.insertAndUpload(goal, video.Name(), thumbnail)
	if err != nil {
		return fmt.Errorf("failed to insert goal for fullname %v: %v", goal.RedditFullname, err)
	}

	return nil
}

func (l *loader) insertAndUpload(goal top90.Goal, videoFilename string, thumbnailFilename string) error {
	videoKey := createKey("mp4")
	goal.S3ObjectKey = videoKey

	thumbnailKey := createKey("jpg")
	goal.ThumbnailS3Key = thumbnailKey

	_, err := l.dao.InsertGoal(&goal)
	if err != nil {
		return err
	}

	err = l.s3Client.UploadFile(videoFilename, videoKey, "video/mp4", l.s3Bucket)
	if err != nil {
		return fmt.Errorf("s3 video upload failed: %v", err)
	}

	err = l.s3Client.UploadFile(thumbnailFilename, thumbnailKey, "image/jpg", l.s3Bucket)
	if err != nil {
		return fmt.Errorf("s3 thumbnail upload failed: %v", err)
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
