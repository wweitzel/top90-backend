package main

import (
	"log/slog"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/wweitzel/top90/internal/api/handlers"
	"github.com/wweitzel/top90/internal/clients/top90"
	"github.com/wweitzel/top90/internal/cmd"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db/dao"
	"github.com/wweitzel/top90/internal/jsonlogger"
)

type seed struct {
	dao    dao.Top90DAO
	client top90.Client
}

var logger *slog.Logger

func main() {
	config := config.Load()
	logger = jsonlogger.New(&jsonlogger.Options{
		Level:    config.LogLevel,
		Colorize: config.LogColor,
	})

	init := cmd.NewInit(logger)

	db := init.DB("admin", "admin", "redditsoccergoals", "localhost", config.DbPort)
	dao := init.Dao(db)
	m := init.Migrate(db.DB)

	m.Down()
	m.Up()

	top90Client := init.Top90Client(10 * time.Second)

	seed := seed{dao, top90Client}
	seed.createS3Bucket("reddit-soccer-goals", "us-east-1", "test", "test", "http://localhost:4566")
	seed.createLeagues(top90Client, dao)
	seed.createTeams(top90Client, dao)
	seed.createFixtures(top90Client, dao)
	err := seed.createPlayers(db)
	if err != nil {
		init.Exit("Error loading players sql file", err)
	}
}

func (seed) createLeagues(client top90.Client, dao dao.Top90DAO) {
	resp, err := client.GetLeagues()
	if err != nil {
		exit("Failed getting leagues", err)
	}

	for _, league := range resp.Leagues {
		logger.Info(league.Name)

		_, err := dao.InsertLeague(&league)
		if err != nil {
			exit("Failed inserting league", err)
		}
	}
}

func (seed) createTeams(client top90.Client, dao dao.Top90DAO) {
	resp, err := client.GetTeams(handlers.GetTeamsRequest{})
	if err != nil {
		exit("Failed getting teams", err)
	}

	for _, team := range resp.Teams {
		logger.Info(team.Name)

		_, err := dao.InsertTeam(&team)
		if err != nil {
			exit("Failed inserting team", err)
		}
	}
}

func (seed) createFixtures(client top90.Client, dao dao.Top90DAO) {
	resp, err := client.GetFixtures(handlers.GetFixturesRequest{
		TodayOnly: true,
	})
	if err != nil {
		exit("Failed getting fixutes", err)
	}

	for _, fixture := range resp.Fixtures {
		logger.Info("Fixture", "teams", fixture.Teams)

		_, err := dao.InsertFixture(&fixture)
		if err != nil {
			exit("Failed inserting fixture", err)
		}
	}
}

func (seed) createPlayers(db *sqlx.DB) error {
	file, err := os.ReadFile("scripts/seed_local_db/players.sql")
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		tx.Rollback()
	}()
	for _, q := range strings.Split(string(file), ";") {
		q := strings.TrimSpace(q)
		q = clean(q)
		if q == "" {
			continue
		}
		if _, err := tx.Exec(q); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func clean(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) {
			return r
		}
		return -1
	}, s)
}

func (seed) createS3Bucket(bucketName, region, accessKey, secretKey, endpoint string) {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
	})
	if err != nil {
		exit("Failed creating s3 session", err)
	}

	s3Client := s3.New(sess)

	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	_, err = s3Client.CreateBucket(input)
	if err != nil {
		exit("Failed creating s3 bucket", err)
	}
}

func exit(msg string, err error) {
	logger.Error(msg, "error", err)
	os.Exit(1)
}
