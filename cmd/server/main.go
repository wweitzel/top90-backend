package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/s3"
)

var dao db.Top90DAO
var s3Client s3.S3Client
var config top90.Config

func main() {
	config = top90.LoadConfig()
	initS3Client()
	initDao()

	r := initRouter()
	http.Handle("/", r)

	port := ":7171"
	log.Println("Listening on http://127.0.0.1" + port)
	http.ListenAndServe(port, nil)
}

func initS3Client() {
	s3Client = s3.NewClient(config.AwsAccessKey, config.AwsSecretAccessKey)
	err := s3Client.VerifyConnection(config.AwsBucketName)
	if err != nil {
		log.Fatalln("Failed to connect to s3 bucket", err)
	}
}

func initDao() {
	DB, err := db.NewPostgresDB(config.DbUser, config.DbPassword, config.DbName, config.DbHost, config.DbPort)
	if err != nil {
		log.Fatalf("Could not set up database: %v", err)
	}

	dao = db.NewPostgresDAO(DB)
}

func initRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", GetApiInfoHandler)
	r.HandleFunc("/fixtures", GetFixturesHandler)
	r.HandleFunc("/goals", GetGoalsHandler)
	r.HandleFunc("/goals/{id}", GetGoalHandler)
	r.HandleFunc("/goals_crawl", GetGoalsCrawlHandler)
	r.HandleFunc("/leagues", GetLeaguesHandler)
	r.HandleFunc("/teams", GetTeamsHandler)
	r.HandleFunc("/message_preview/{id}", MessagePreviewHandler)
	return r
}
