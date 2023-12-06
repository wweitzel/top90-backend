package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sync"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/cmd"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/jsonlogger"
)

var dao db.Top90DAO
var s3Client s3.S3Client
var logger = jsonlogger.New(&jsonlogger.Options{
	Level:    slog.LevelDebug,
	Colorize: true,
})

func main() {
	config := config.Load()
	init := cmd.NewInit(logger)

	s3Client = init.S3Client(
		s3.Config{
			AccessKey:       config.AwsAccessKey,
			SecretAccessKey: config.AwsSecretAccessKey,
			Endpoint:        config.AwsS3Endpoint,
			Logger:          logger,
		},
		config.AwsBucketName)

	database := init.DB(
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbHost,
		config.DbPort)

	dao = init.Dao(database)

	goals, err := dao.GetGoals(db.Pagination{Skip: 0, Limit: 10}, db.GetGoalsFilter{})
	if err != nil {
		init.Exit("Failed getting goals from database", err)
	}

	updateThumbnails(goals, config.AwsBucketName)
}

type UpdateThumbnailJob struct {
	Goal       top90.Goal
	BucketName string
	Id         int
}

func updateThumbnails(goals []top90.Goal, bucketName string) {
	jobs := make(chan UpdateThumbnailJob, len(goals))

	var wg sync.WaitGroup

	var worker = func(job chan UpdateThumbnailJob) {
		for job := range jobs {
			func() {
				defer wg.Done()
				updateThumbnail(job.Goal, job.BucketName, job.Id)
			}()
		}
	}

	const workers = 5
	for w := 0; w < workers; w++ {
		go worker(jobs)
	}

	for i, goal := range goals {
		wg.Add(1)
		job := UpdateThumbnailJob{
			Goal:       goal,
			BucketName: bucketName,
			Id:         i,
		}
		jobs <- job
	}

	close(jobs)
	wg.Wait()
}

func updateThumbnail(goal top90.Goal, bucketName string, id int) error {
	videoFilename := fmt.Sprintf("tmp/vid#%d.mp4", id)
	defer os.Remove(videoFilename)
	s3Client.DownloadFile(goal.S3ObjectKey, bucketName, videoFilename)

	thumbnailFilename := fmt.Sprintf("tmp/thumb#%d.jpg", id)
	defer os.Remove(thumbnailFilename)
	ffmpegPath := os.Getenv("TOP90_FFMPEG_PATH")
	cmd := exec.Command(ffmpegPath, "-i", videoFilename, "-q:v", "8", "-vframes", "1", thumbnailFilename)
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		return err
	}

	err = s3Client.UploadFile(thumbnailFilename, goal.ThumbnailS3Key, "image/jpg", bucketName)
	if err != nil {
		logger.Error("Failed uploading to s3", "error", err)
		return err
	}

	logger.Info("Successfully updated video in s3", "title", goal.RedditPostTitle)
	return nil
}
