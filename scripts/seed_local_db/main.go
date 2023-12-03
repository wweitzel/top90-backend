package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/wweitzel/top90/internal/api/handlers"
	"github.com/wweitzel/top90/internal/clients/top90"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db"
)

type seed struct {
	dao    db.Top90DAO
	client top90.Client
}

func main() {
	config := config.Load()

	DB, err := db.NewPostgresDB("admin", "admin", "redditsoccergoals", "localhost", config.DbPort)
	if err != nil {
		log.Fatalln("Could not setup database:", err)
	}

	driver, err := postgres.WithInstance(DB, &postgres.Config{})
	if err != nil {
		log.Fatalln("Could not setup database driver:", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/db/migrations",
		"postgres", driver)
	if err != nil {
		log.Fatalln("Could not instantiate migrate:", err)
	}

	m.Down()
	m.Up()

	dao := db.NewPostgresDAO(DB)

	top90Client := top90.NewClient(top90.Config{
		Timeout: 10 * time.Second,
	})

	seed := seed{dao, top90Client}

	seed.createS3Bucket("reddit-soccer-goals", "us-east-1", "test", "test", "http://localhost:4566")
	seed.createLeagues(top90Client, dao)
	seed.createTeams(top90Client, dao)
	seed.createFixtures(top90Client, dao)
}

func (seed) createLeagues(client top90.Client, dao db.Top90DAO) {
	resp, err := client.GetLeagues()
	if err != nil {
		log.Fatalln(err)
	}

	for _, league := range resp.Leagues {
		log.Println(league.Name)
		_, err := dao.InsertLeague(&league)
		if err != nil && err != sql.ErrNoRows {
			log.Fatalln("Could not insert league:", err)
		}
	}
}

func (seed) createTeams(client top90.Client, dao db.Top90DAO) {
	resp, err := client.GetTeams(handlers.GetTeamsRequest{})
	if err != nil {
		log.Fatalln(err)
	}

	for _, team := range resp.Teams {
		log.Println(team.Name)
		_, err := dao.InsertTeam(&team)
		if err != nil && err != sql.ErrNoRows {
			log.Fatalln("Could not insert team:", err)
		}
	}
}

func (seed) createFixtures(client top90.Client, dao db.Top90DAO) {
	resp, err := client.GetFixtures(handlers.GetFixturesRequest{
		TodayOnly: true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	for _, fixture := range resp.Fixtures {
		log.Println(fixture.Date)
		_, err := dao.InsertFixture(&fixture)
		if err != nil && err != sql.ErrNoRows {
			log.Fatalln("Could not insert fixture:", err)
		}
	}
}

func (seed) createS3Bucket(bucketName, region, accessKey, secretKey, endpoint string) {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
	})
	if err != nil {
		log.Fatalln("Could not create s3 bucket:", err)
	}

	s3Client := s3.New(sess)

	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	_, err = s3Client.CreateBucket(input)
	if err != nil {
		log.Fatalln("Could not create s3 bucket:", err)
	}
}
