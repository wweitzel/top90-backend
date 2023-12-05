package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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
		log.Fatalln("Failed to connect to s3 bucket", err)
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

	goals, err := pgDAO.GetGoals(db.Pagination{Skip: 0, Limit: 10000000}, db.GetGoalsFilter{})
	if err != nil {
		log.Fatal(err)
	}

	errorCount := 0

	start := time.Now()

	// Run the thumbanil extraction on a bunch of go routines. Currently
	// tries to kick off 100 go routines every 30 seconds.
	wg := new(sync.WaitGroup)
	for i, goal := range goals {
		if i != 0 && i%100 == 0 {
			log.Println("Sleeping 30 seconds...")
			time.Sleep(30 * time.Second)
		} else {
			time.Sleep(1 * time.Millisecond)
		}
		wg.Add(1)
		go func(goal top90.Goal, wg *sync.WaitGroup, i int) {
			var err error

			if goal.ThumbnailS3Key == "" {
				err = extractThumbnail(goal, s3Client, "reddit-soccer-goals", fmt.Sprintf("tmp/vid#%d.mp4", i), i)
			}

			// var err error
			// if goal.ThumbnailS3Key != "" {
			// 	log.Println(goal.ThumbnailS3Key)
			// 	err = s3Client.DeleteFile(goal.ThumbnailS3Key, "reddit-soccer-goals")
			// }

			if err != nil {
				errorCount += 1
				log.Println("TOP90 ERROR:", errorCount, err)
			} else {
				log.Println("TOP90 SUCCESS:", i)
			}
			defer wg.Done()
		}(goal, wg, i)
	}
	wg.Wait()

	// This will run the processing on a single thread. Uncomment when debugging.
	// for i, goal := range goals {
	// 	err := extractThumbnail(goal, &s3Client, "reddit-soccer-goals", fmt.Sprintf("tmp/vid#%d.mp4", i), i)

	// 	if err != nil {
	// 		errorCount += 1
	// 		log.Println("TOP90_ERROR: ", err)
	// 	}
	// 	log.Println("Current:", i)
	// 	log.Println("Error count:", errorCount)
	// }

	duration := time.Since(start)
	fmt.Println("----------------------------------------")
	fmt.Println("Total Time:", duration)
	log.Println("Total:", len(goals))
	log.Println("Total Errors:", errorCount)
}

func extractThumbnail(goal top90.Goal, s3Client *s3.S3Client, bucketName string, videoFilename string, i int) error {
	// TODO: Figure out how to pass bytes directly to ffmpeg instead of
	// passing a file as input (would be a faster solution). The commented out
	// code is close, but fails for certain videos.
	//
	// cmd := exec.Command("/usr/local/bin/ffmpeg", "-i", "pipe:0", "-vframes", "1", thumbnailFilename)
	// cmd.Stdin = bytes.NewReader(video)

	s3Client.DownloadFile(goal.S3ObjectKey, bucketName, videoFilename)
	thumbnailFilename := fmt.Sprintf("tmp/thumb#%d.jpg", i)

	ffmpegPath := os.Getenv("TOP90_FFMPEG_PATH")
	cmd := exec.Command(ffmpegPath, "-i", videoFilename, "-q:v", "2", "-vframes", "1", thumbnailFilename)
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	os.Remove(videoFilename)
	if err != nil {
		return err
	}

	objectKey, err := uploadFile(thumbnailFilename)
	if err != nil {
		return err
	}

	goalUpdate := top90.Goal{ThumbnailS3Key: objectKey}
	updatedGoal, err := pgDAO.UpdateGoal(goal.Id, goalUpdate)
	os.Remove(thumbnailFilename)
	if err != nil {
		return err
	}

	log.Println("Successfully updated goal:", updatedGoal.ThumbnailS3Key)
	return nil
}

func uploadFile(fileName string) (objectKey string, error error) {
	key := createKey()

	bucketName := "reddit-soccer-goals"
	err := s3Client.UploadFile(fileName, key, "image/png", bucketName)
	if err != nil {
		log.Println("s3 upload failed", err)
		return "", err
	} else {
		log.Println("Successfully uploaded video to s3", key)
	}

	return key, nil
}

func createKey() string {
	nowUtc := time.Now().UTC()
	yearMonthDayStr := fmt.Sprintf("%d-%02d-%02d",
		nowUtc.Year(), nowUtc.Month(), nowUtc.Day())

	id := uuid.NewString()
	id = strings.Replace(id, "-", "", -1)
	objectKey := yearMonthDayStr + "/" + id + ".png"
	return objectKey
}
