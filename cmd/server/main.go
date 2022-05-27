package main

import (
	"database/sql"
	"log"
	"net/http"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/s3"
)

var DB *sql.DB
var dao db.Top90DAO

var s3Client s3.S3Client
var config top90.Config

func main() {
	log.SetFlags(log.Ltime)

	// Load config from .env into environment variables
	config = top90.LoadConfig()

	DB, err := db.NewPostgresDB(config.DbUser, config.DbPassword, config.DbName, config.DbHost, config.DbPort)
	if err != nil {
		log.Fatalf("Could not set up database: %v", err)
	}
	defer DB.Close()

	s3Client = s3.NewClient(config.AwsAccessKey, config.AwsSecretAccessKey)
	err = s3Client.VerifyConnection(config.AwsBucketName)
	if err != nil {
		log.Fatalln("Failed to connect to s3 bucket", err)
	}

	dao = db.NewPostgresDAO(DB)

	http.HandleFunc("/", GetApiInfoHandler)
	http.HandleFunc("/goals", GetGoalsHandler)
	http.HandleFunc("/goals_crawl", GetGoalsCrawlHandler)

	// Start the server
	port := ":7171"
	log.Println("Listening on http://127.0.0.1" + port)
	http.ListenAndServe(port, nil)
}
