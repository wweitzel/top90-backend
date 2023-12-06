package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/wweitzel/top90/internal/config/dotenv"
	"github.com/wweitzel/top90/internal/jsonlogger"
)

type Config struct {
	LogLevel                slog.Leveler
	LogColor                bool
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

// Load env file into system env variables and struct
func Load(fileNames ...string) Config {
	err := dotenv.Load(fileNames...)
	if err != nil {
		jsonlogger.New(nil).Info("No local .env found. Will use existing system environment variables.")
	}

	return Config{
		LogLevel: logLevel(),
		LogColor: logColor(),

		DbUser:     os.Getenv("TOP90_DB_USER"),
		DbPassword: os.Getenv("TOP90_DB_PASSWORD"),
		DbName:     os.Getenv("TOP90_DB_NAME"),
		DbHost:     os.Getenv("TOP90_DB_HOST"),
		DbPort:     os.Getenv("TOP90_DB_PORT"),

		AwsAccessKey:       os.Getenv("TOP90_AWS_ACCESS_KEY"),
		AwsSecretAccessKey: os.Getenv("TOP90_AWS_SECRET_ACCESS_KEY"),
		AwsBucketName:      os.Getenv("TOP90_AWS_BUCKET_NAME"),
		AwsS3Endpoint:      os.Getenv("TOP90_AWS_S3_ENDPOINT"),

		FFmpegPath: os.Getenv("TOP90_FFMPEG_PATH"),

		ApiFootballRapidApiHost: os.Getenv("API_FOOTBALL_RAPID_API_HOST"),
		ApiFootballRapidApiKey:  os.Getenv("API_FOOTBALL_RAPID_API_KEY"),
	}
}

func logColor() bool {
	logColor := os.Getenv("TOP90_LOG_COLOR")
	if logColor == "" {
		return false
	}
	color, err := strconv.ParseBool(logColor)
	if err != nil {
		return false
	}
	return color
}

func logLevel() slog.Leveler {
	level := os.Getenv("TOP90_LOG_LEVEL")
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
