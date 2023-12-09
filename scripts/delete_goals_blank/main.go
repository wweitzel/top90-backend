package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/cmd"
	"github.com/wweitzel/top90/internal/config"
	db "github.com/wweitzel/top90/internal/db/models"
	"github.com/wweitzel/top90/internal/jsonlogger"
)

func main() {
	config := config.Load()
	logger := jsonlogger.New(&jsonlogger.Options{
		Level:    config.LogLevel,
		Colorize: config.LogColor,
	})

	init := cmd.NewInit(logger)

	dbConn := init.DB(
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbHost,
		config.DbPort)

	s3Client := init.S3Client(
		s3.Config{
			AccessKey:       config.AwsAccessKey,
			SecretAccessKey: config.AwsSecretAccessKey,
			Endpoint:        config.AwsS3Endpoint,
			Logger:          logger,
		},
		config.AwsBucketName)

	keys, err := s3Client.ListAllObjects(config.AwsBucketName)
	if err != nil {
		init.Exit("Failed listing objects", err)
	}
	logger.Info("Total objects: " + fmt.Sprint(len(keys)))

	for _, key := range keys {
		query := "SELECT * FROM goals WHERE s3_object_key = $1 or thumbnail_s3_key = $2"
		args := []any{key, key}
		var goal db.Goal
		err := dbConn.Get(&goal, query, args...)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			init.Exit("Failed getting goal for key "+key, err)
		}

		if err != nil && errors.Is(err, sql.ErrNoRows) {
			logger.Info("No goal found for key " + key)
			err = s3Client.DeleteObject(key, config.AwsBucketName)
			if err != nil {
				init.Exit("Failed deleting key "+key, err)
			}
			logger.Info("Successfully deleted object " + key)
			continue
		}

		out, err := s3Client.HeadObject(key, config.AwsBucketName)
		if err != nil {
			init.Exit("Failed heading object "+key, err)
		}
		if aws.Int64Value(out.ContentLength) == 0 {
			logger.Info("Found a 0 byte object: " + key)

			if err == nil && goal != (db.Goal{}) {
				logger.Info("Goal with 0 byte file", "goal", goal)
				if len(string(goal.ThumbnailS3Key)) > 0 {
					err = s3Client.DeleteObject(string(goal.ThumbnailS3Key), config.AwsBucketName)
					if err != nil {
						init.Exit("Failed deleting thumbnail file", err)
					}
				}
				err := s3Client.DeleteObject(goal.S3ObjectKey, config.AwsBucketName)
				if err != nil {
					init.Exit("Failed deleting video file", err)
				}
				query := "DELETE FROM goals WHERE id = $1"
				args := []any{key}
				var deletedGoal db.Goal
				err = dbConn.QueryRowx(query, args...).StructScan(&goal)
				if err != nil {
					init.Exit("Failed deleting goal from db", err)
				}
				logger.Info("Successfully deleted goal and associated objects", "goal", deletedGoal)
			}
		}
	}
}
