package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/db/postgres/dao"
)

var pgDAO db.Top90DAO
var s3Client *s3.S3Client

func main() {
	log.SetFlags(log.Ltime)

	config := config.Load()

	DB, err := db.NewPostgresDB(config.DbUser, config.DbPassword, config.DbName, config.DbHost, config.DbPort)
	if err != nil {
		log.Fatalf("Could not set up database: %v", err)
	}
	defer DB.Close()

	s3Client, err = s3.NewClient(config.AwsAccessKey, config.AwsSecretAccessKey)
	if err != nil {
		log.Fatalln("Failed to create s3 client", err)
	}
	err = s3Client.VerifyConnection(config.AwsBucketName)
	if err != nil {
		log.Fatalln("Failed to connect to s3 bucket", err)
	}

	pgDAO = dao.NewPostgresDAO(DB)

	goalsCount, err := pgDAO.CountGoals(db.GetGoalsFilter{})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Total Goals:", goalsCount)

	goals, err := pgDAO.GetGoals(db.Pagination{Skip: 0, Limit: 10}, db.GetGoalsFilter{})
	if err != nil {
		log.Fatal(err)
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
		log.Println(err)
		return err
	}

	err = s3Client.UploadFile(thumbnailFilename, goal.ThumbnailS3Key, "image/jpg", bucketName)
	if err != nil {
		log.Println("s3 upload failed", err)
	} else {
		log.Println("Successfully updated video in s3", goal.ThumbnailS3Key)
	}

	return nil
}
