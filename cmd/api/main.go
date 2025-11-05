package main

import (
	"net/http"
	"os"

	"github.com/newrelic/go-agent/v3/newrelic"
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

	s3Client := init.S3Client(s3.Config{
		AccessKey:       config.AwsAccessKey,
		SecretAccessKey: config.AwsSecretAccessKey,
		Endpoint:        config.AwsS3Endpoint,
		Logger:          logger,
	}, config.AwsBucketName)

	db := init.DB(
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbHost,
		config.DbPort)

	db.SetMaxOpenConns(50)

	dao := init.Dao(db)

	metrics, err := newrelic.NewApplication(
		newrelic.ConfigAppName("top90"),
		newrelic.ConfigLicense(config.NewRelicLicenseKey),
	)
	if err != nil {
		logger.Error("Failed to start New Relic", "error", err)
		os.Exit(1)
	}

	s := api.NewServer(
		dao,
		s3Client,
		config,
		logger,
		metrics)

	port := ":7171"
	logger.Info("Listening on http://127.0.0.1" + port)
	http.ListenAndServe(port, s)
}
