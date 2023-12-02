package config

import (
	"log"
	"os"

	"github.com/wweitzel/top90/internal/config/dotenv"
)

type Config struct {
	DbUser                  string
	DbPassword              string
	DbName                  string
	DbHost                  string
	DbPort                  string
	AwsAccessKey            string
	AwsSecretAccessKey      string
	AwsBucketName           string
	AwsS3Endpoint           string
	RedditBasicAuth         string
	FFmpegPath              string
	ApiFootballRapidApiHost string
	ApiFootballRapidApiKey  string
}

// Load env file into struct
func Load(fileNames ...string) Config {
	err := dotenv.Load(fileNames...)
	if err != nil {
		log.Println("No local .env found. Will use existing system environment variables.")
	}

	return Config{
		DbUser:     os.Getenv("TOP90_DB_USER"),
		DbPassword: os.Getenv("TOP90_DB_PASSWORD"),
		DbName:     os.Getenv("TOP90_DB_NAME"),
		DbHost:     os.Getenv("TOP90_DB_HOST"),
		DbPort:     os.Getenv("TOP90_DB_PORT"),

		AwsAccessKey:       os.Getenv("TOP90_AWS_ACCESS_KEY"),
		AwsSecretAccessKey: os.Getenv("TOP90_AWS_SECRET_ACCESS_KEY"),
		AwsBucketName:      os.Getenv("TOP90_AWS_BUCKET_NAME"),
		AwsS3Endpoint:      os.Getenv("TOP90_AWS_S3_ENDPOINT"),

		RedditBasicAuth: os.Getenv("TOP90_REDDIT_BASIC_AUTH"),
		FFmpegPath:      os.Getenv("TOP90_FFMPEG_PATH"),

		ApiFootballRapidApiHost: os.Getenv("API_FOOTBALL_RAPID_API_HOST"),
		ApiFootballRapidApiKey:  os.Getenv("API_FOOTBALL_RAPID_API_KEY"),
	}
}
