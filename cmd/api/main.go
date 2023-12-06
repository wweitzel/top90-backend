package main

import (
	"net/http"

	"github.com/wweitzel/top90/internal/api"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/cmd"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/jsonlogger"
)

func main() {
	config := config.Load()
	logger := jsonlogger.New(&jsonlogger.Options{
		Level:    config.LogLevel,
		Colorize: config.LogColor,
	})

	init := cmd.NewInit(logger)

	s3Client := init.S3Client(
		s3.Config{
			AccessKey:       config.AwsAccessKey,
			SecretAccessKey: config.AwsSecretAccessKey,
			Endpoint:        config.AwsS3Endpoint,
			Logger:          logger,
		},
		config.AwsBucketName)

	dao := init.Dao(
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbHost,
		config.DbPort,
	)

	s := api.NewServer(
		dao,
		s3Client,
		config,
		logger)

	port := ":7171"
	logger.Info("Listening on http://127.0.0.1" + port)
	http.ListenAndServe(port, s)
}
