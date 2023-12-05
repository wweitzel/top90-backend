package main

import (
	"log"
	"net/http"

	"github.com/wweitzel/top90/internal/api"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/db/postgres/dao"
	"github.com/wweitzel/top90/internal/jsonlogger"
)

func main() {
	config := config.Load()

	logger := jsonlogger.New(&jsonlogger.Options{
		Level:    config.LogLevel,
		Colorize: config.LogColor,
	})

	s3Client := initS3Client(
		config.AwsAccessKey,
		config.AwsSecretAccessKey,
		config.AwsBucketName)

	dao := initDao(
		config.DbUser,
		config.DbPassword,
		config.DbName,
		config.DbHost,
		config.DbPort)

	s := api.NewServer(
		dao,
		s3Client,
		config)

	port := ":7171"
	logger.Info("Listening on http://127.0.0.1" + port)
	http.ListenAndServe(port, s)
}

func initS3Client(accessKey, secretAccessKey, bucketName string) s3.S3Client {
	s3Client, err := s3.NewClient(accessKey, secretAccessKey)
	if err != nil {
		log.Fatalln("Failed to create s3 client", err)
	}

	err = s3Client.VerifyConnection(bucketName)
	if err != nil {
		log.Fatalln("Failed to connect to s3 bucket", err)
	}
	return *s3Client
}

func initDao(user, password, name, host, port string) db.Top90DAO {
	DB, err := db.NewPostgresDB(user, password, name, host, port)
	if err != nil {
		log.Fatalf("Could not set up database: %v", err)
	}

	return dao.NewPostgresDAO(DB)
}
