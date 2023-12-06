package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/wweitzel/top90/internal/api"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/db/postgres/dao"
	"github.com/wweitzel/top90/internal/jsonlogger"
)

var logger *slog.Logger

func main() {
	config := config.Load()

	logger = jsonlogger.New(&jsonlogger.Options{
		Level:    config.LogLevel,
		Colorize: config.LogColor,
	})

	s3Client := initS3Client(s3.Config{
		AccessKey:       config.AwsAccessKey,
		SecretAccessKey: config.AwsSecretAccessKey,
		Endpoint:        config.AwsS3Endpoint,
		Logger:          logger,
	}, config.AwsBucketName)

	dao := initDao(
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
		logger,
	)

	port := ":7171"
	logger.Info("Listening on http://127.0.0.1" + port)
	http.ListenAndServe(port, s)
}

func initS3Client(cfg s3.Config, bucket string) s3.S3Client {
	s3Client, err := s3.NewClient(cfg)
	if err != nil {
		logger.Error("Could not create s3 client", "error", err)
		os.Exit(1)
	}

	err = s3Client.VerifyConnection(bucket)
	if err != nil {
		logger.Error("Could not connect to s3 bucket", "error", err)
		os.Exit(1)
	}
	return *s3Client
}

func initDao(user, password, name, host, port string) db.Top90DAO {
	DB, err := db.NewPostgresDB(user, password, name, host, port)
	if err != nil {
		logger.Error("Could not set up database: %v", err)
		os.Exit(1)
	}

	return dao.NewPostgresDAO(DB)
}
