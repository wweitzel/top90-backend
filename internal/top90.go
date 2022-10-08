package top90

import (
	"log"
	"os"
	"time"

	"github.com/wweitzel/top90/internal/dotenv"
)

type Goal struct {
	Id                  string
	RedditFullname      string
	RedditLinkUrl       string
	RedditPostTitle     string
	RedditPostCreatedAt time.Time
	S3ObjectKey         string
	PresignedUrl        string
	CreatedAt           string
}

type Config struct {
	DbUser             string
	DbPassword         string
	DbName             string
	DbHost             string
	DbPort             string
	AwsAccessKey       string
	AwsSecretAccessKey string
	AwsBucketName      string
}

func LoadConfig(fileNames ...string) Config {
	// Export all environment variables in .env file
	err := dotenv.Load(fileNames...)
	if err != nil {
		log.Println("Could not load env file:", err)
	}

	// Extract the loaded environment variables into config struct
	return Config{
		DbUser:     os.Getenv("TOP90_DB_USER"),
		DbPassword: os.Getenv("TOP90_DB_PASSWORD"),
		DbName:     os.Getenv("TOP90_DB_NAME"),
		DbHost:     os.Getenv("TOP90_DB_HOST"),
		DbPort:     os.Getenv("TOP90_DB_PORT"),

		AwsAccessKey:       os.Getenv("TOP90_AWS_ACCESS_KEY"),
		AwsSecretAccessKey: os.Getenv("TOP90_AWS_SECRET_ACCESS_KEY"),
		AwsBucketName:      os.Getenv("TOP90_AWS_BUCKET_NAME"),
	}
}
