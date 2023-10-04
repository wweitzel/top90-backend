package main

import (
	"log"
	"net/http"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/s3"
	"github.com/wweitzel/top90/internal/server"
)

func main() {
	config := top90.LoadConfig()

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

	s := server.NewServer(
		dao,
		s3Client,
		config)

	port := ":7171"
	log.Println("Listening on http://127.0.0.1" + port)
	http.ListenAndServe(port, s)
}

func initS3Client(accessKey, secretAccessKey, bucketName string) s3.S3Client {
	s3Client := s3.NewClient(accessKey, secretAccessKey)
	err := s3Client.VerifyConnection(bucketName)
	if err != nil {
		log.Fatalln("Failed to connect to s3 bucket", err)
	}
	return s3Client
}

func initDao(user, password, name, host, port string) db.Top90DAO {
	DB, err := db.NewPostgresDB(user, password, name, host, port)
	if err != nil {
		log.Fatalf("Could not set up database: %v", err)
	}

	return db.NewPostgresDAO(DB)
}
