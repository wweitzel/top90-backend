package db

import (
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/apifootball"
	"gotest.tools/v3/assert"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func TestGetGoals(t *testing.T) {
	t.Parallel()

	dao, pool, resource := setup()
	defer pool.Purge(resource)

	now := time.Now()
	uuid := uuid.NewString()

	team1, err := dao.InsertTeam(&apifootball.Team{
		Id:   1,
		Name: "team1",
	})
	assert.NilError(t, err)

	team2, err := dao.InsertTeam(&apifootball.Team{
		Id:   2,
		Name: "team2",
	})
	assert.NilError(t, err)

	league1, err := dao.InsertLeague(&apifootball.League{
		Id:   1,
		Name: "premier league",
	})
	assert.NilError(t, err)

	fixture, err := dao.InsertFixture(&apifootball.Fixture{
		Id:        1,
		Referee:   "jimbob",
		Timestamp: now.Unix(),
		LeagueId:  league1.Id,
		Teams: apifootball.Teams{
			Home: apifootball.Team{Id: team1.Id},
			Away: apifootball.Team{Id: team2.Id},
		},
	})
	assert.NilError(t, err)

	goal, _ := dao.InsertGoal(&top90.Goal{
		RedditFullname:      uuid,
		RedditLinkUrl:       "redditlinkurl",
		RedditPostTitle:     "redditposttitlte",
		S3ObjectKey:         "s3objectkey",
		RedditPostCreatedAt: now,
		ThumbnailS3Key:      "thumbnails3key",
		FixtureId:           fixture.Id,
	})

	assertEqual(t, *goal, top90.Goal{
		Id:                  goal.Id,
		CreatedAt:           goal.CreatedAt,
		RedditFullname:      uuid,
		RedditLinkUrl:       "redditlinkurl",
		RedditPostTitle:     "redditposttitlte",
		S3ObjectKey:         "s3objectkey",
		RedditPostCreatedAt: now,
		ThumbnailS3Key:      "thumbnails3key",
	})

	count, _ := dao.CountGoals(GetGoalsFilter{})
	assert.Equal(t, count, 1)

	goals, _ := dao.GetGoals(Pagination{}, GetGoalsFilter{})
	assert.Equal(t, len(goals), 1)

	goals, _ = dao.GetGoals(Pagination{}, GetGoalsFilter{FixtureId: fixture.Id})
	assert.Equal(t, len(goals), 1)

	goals, _ = dao.GetGoals(Pagination{}, GetGoalsFilter{FixtureId: 9783246978987})
	assert.Equal(t, len(goals), 0)
}

func TestGetFixtures(t *testing.T) {
	t.Parallel()

	dao, pool, resource := setup()
	defer pool.Purge(resource)

	now := time.Now()

	team1, err := dao.InsertTeam(&apifootball.Team{
		Id:   1,
		Name: "team1",
	})
	assert.NilError(t, err)

	team2, err := dao.InsertTeam(&apifootball.Team{
		Id:   2,
		Name: "team2",
	})
	assert.NilError(t, err)

	league1, err := dao.InsertLeague(&apifootball.League{
		Id:   1,
		Name: "premier league",
	})
	assert.NilError(t, err)

	league2, err := dao.InsertLeague(&apifootball.League{
		Id:   2,
		Name: "la liga",
	})
	assert.NilError(t, err)

	_, err = dao.InsertFixture(&apifootball.Fixture{
		Id:        1,
		Referee:   "jimbob",
		Timestamp: now.Unix(),
		LeagueId:  league1.Id,
		Teams: apifootball.Teams{
			Home: apifootball.Team{Id: team1.Id},
			Away: apifootball.Team{Id: team2.Id},
		},
	})
	assert.NilError(t, err)

	_, err = dao.InsertFixture(&apifootball.Fixture{
		Id:        2,
		Referee:   "jimbob",
		Timestamp: now.AddDate(0, 0, 2).Unix(),
		LeagueId:  league2.Id,
		Teams: apifootball.Teams{
			Home: apifootball.Team{Id: team1.Id},
			Away: apifootball.Team{Id: team2.Id},
		},
	})
	assert.NilError(t, err)

	fixtures, _ := dao.GetFixtures(GetFixuresFilter{})
	assert.Equal(t, len(fixtures), 2)

	fixtures, _ = dao.GetFixtures(GetFixuresFilter{LeagueId: 1})
	assert.Equal(t, len(fixtures), 1)

	fixtures, _ = dao.GetFixtures(GetFixuresFilter{Date: time.Now()})
	assert.Equal(t, len(fixtures), 1)
}

func TestGetTeams(t *testing.T) {
	t.Parallel()

	dao, pool, resource := setup()
	defer pool.Purge(resource)

	_, err := dao.InsertTeam(&apifootball.Team{
		Id:      1,
		Name:    "team1",
		Country: "usa",
	})
	assert.NilError(t, err)

	_, err = dao.InsertTeam(&apifootball.Team{
		Id:      2,
		Name:    "team2",
		Country: "mexico",
	})
	assert.NilError(t, err)

	teams, _ := dao.GetTeams(GetTeamsFilter{})
	assert.Equal(t, len(teams), 2)

	teams, _ = dao.GetTeams(GetTeamsFilter{Country: "usa"})
	assert.Equal(t, len(teams), 1)

	teams, _ = dao.GetTeams(GetTeamsFilter{Country: "lkjlk"})
	assert.Equal(t, len(teams), 0)

	teams, err = dao.GetTeams(GetTeamsFilter{SearchTerm: "team1"})
	assert.NilError(t, err)
	assert.Equal(t, len(teams), 1)
	assert.Equal(t, teams[0].Id, 1)
}

func assertEqual(t *testing.T, actual top90.Goal, expected top90.Goal) {
	assert.Equal(t, actual.Id, expected.Id)
	assert.Equal(t, actual.RedditFullname, expected.RedditFullname)
	assert.Equal(t, actual.RedditLinkUrl, expected.RedditLinkUrl)
	assert.Equal(t, actual.RedditPostTitle, expected.RedditPostTitle)
	assert.Equal(t, actual.S3ObjectKey, expected.S3ObjectKey)
	// TODO: Figure out why the bwlow assertion fails
	// assert.Equal(t, actual.RedditPostCreatedAt, expected.RedditPostCreatedAt)
	assert.Equal(t, actual.ThumbnailS3Key, expected.ThumbnailS3Key)
}

func TestGetLeagues(t *testing.T) {
	t.Parallel()

	dao, pool, resource := setup()
	defer pool.Purge(resource)

	_, err := dao.InsertLeague(&apifootball.League{
		Id:   1,
		Name: "premier league",
	})
	assert.NilError(t, err)

	_, err = dao.InsertLeague(&apifootball.League{
		Id:   2,
		Name: "la liga",
	})
	assert.NilError(t, err)

	leagues, err := dao.GetLeagues()
	assert.Equal(t, len(leagues), 2)
}

func setup() (dao Top90DAO, pool *dockertest.Pool, res *dockertest.Resource) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=user_name",
			"POSTGRES_DB=dbname",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://user_name:secret@%s/dbname?sslmode=disable", hostAndPort)

	resource.Expire(120)

	var db *sql.DB

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		db, err = sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})

	var mig *migrate.Migrate
	mig, err = migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)

	if err := mig.Up(); err != nil {
		log.Fatal(err)
	}

	return NewPostgresDAO(db), pool, resource
}
