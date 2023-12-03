package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db"
)

var dao db.Top90DAO
var s3Client s3.S3Client

func main() {
	log.SetFlags(log.Ltime)

	config := config.Load()

	DB, err := db.NewPostgresDB(config.DbUser, config.DbPassword, config.DbName, config.DbHost, config.DbPort)
	if err != nil {
		log.Fatalf("Could not set up database: %v", err)
	}
	defer DB.Close()

	s3Client = s3.NewClient(config.AwsAccessKey, config.AwsSecretAccessKey)
	err = s3Client.VerifyConnection(config.AwsBucketName)
	if err != nil {
		log.Fatalln("Failed to connect to s3 bucket", err)
	}

	dao = db.NewPostgresDAO(DB)

	goalsCount, err := dao.CountGoals(db.GetGoalsFilter{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Total Goals:", goalsCount)

	goals, err := dao.GetGoals(db.Pagination{Skip: 0, Limit: 100}, db.GetGoalsFilter{})
	if err != nil {
		log.Fatal(err)
	}

	for index, goal := range goals {
		log.Println(goal.RedditPostTitle)
		tmpFilePath := fmt.Sprintf("tmp/vid#%d.mp4", index)
		updateThumnnail(goal, "reddit-soccer-goals", tmpFilePath, index)
	}
}

func updateThumnnail(goal top90.Goal, bucketName string, videoFilename string, i int) error {
	s3Client.DownloadFile(goal.S3ObjectKey, bucketName, videoFilename)
	thumbnailFilename := fmt.Sprintf("tmp/thumb#%d.avif", i)
	defer os.Remove(thumbnailFilename)
	defer os.Remove(videoFilename)

	ffmpegPath := os.Getenv("TOP90_FFMPEG_PATH")
	cmd := exec.Command(ffmpegPath, "-i", videoFilename, "-q:v", "2", "-vframes", "1", thumbnailFilename)
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		log.Println(err)
		return err
	}

	err = s3Client.UploadFile(thumbnailFilename, goal.ThumbnailS3Key, "image/avif", bucketName)
	if err != nil {
		log.Println("s3 upload failed", err)
	} else {
		log.Println("Successfully updated video in s3", goal.ThumbnailS3Key)
	}

	return nil
}
