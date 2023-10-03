package server

import (
	"log"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/s3"
)

var dao db.Top90DAO
var s3Client s3.S3Client
var config top90.Config

func LoadConfig() {
	config = top90.LoadConfig()
}

func InitS3Client() {
	s3Client = s3.NewClient(config.AwsAccessKey, config.AwsSecretAccessKey)
	err := s3Client.VerifyConnection(config.AwsBucketName)
	if err != nil {
		log.Fatalln("Failed to connect to s3 bucket", err)
	}
}

func InitDao() {
	DB, err := db.NewPostgresDB(config.DbUser, config.DbPassword, config.DbName, config.DbHost, config.DbPort)
	if err != nil {
		log.Fatalf("Could not set up database: %v", err)
	}

	dao = db.NewPostgresDAO(DB)
}
