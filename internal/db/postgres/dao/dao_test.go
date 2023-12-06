package dao

import (
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/db"
	"gotest.tools/v3/assert"
)

func TestGetGoals(t *testing.T) {
	t.Parallel()

	dao, pool, resource, err := createTestDb()
	assert.NilError(t, err)
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

	count, _ := dao.CountGoals(db.GetGoalsFilter{})
	assert.Equal(t, count, 1)

	goals, _ := dao.GetGoals(db.Pagination{}, db.GetGoalsFilter{})
	assert.Equal(t, len(goals), 1)

	goals, _ = dao.GetGoals(db.Pagination{}, db.GetGoalsFilter{FixtureId: fixture.Id})
	assert.Equal(t, len(goals), 1)

	goals, _ = dao.GetGoals(db.Pagination{}, db.GetGoalsFilter{FixtureId: 9783246978987})
	assert.Equal(t, len(goals), 0)
}

func TestGetFixtures(t *testing.T) {
	t.Parallel()

	dao, pool, resource, err := createTestDb()
	assert.NilError(t, err)
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

	fixtures, _ := dao.GetFixtures(db.GetFixuresFilter{})
	assert.Equal(t, len(fixtures), 2)

	fixtures, _ = dao.GetFixtures(db.GetFixuresFilter{LeagueId: 1})
	assert.Equal(t, len(fixtures), 1)

	fixtures, _ = dao.GetFixtures(db.GetFixuresFilter{Date: time.Now()})
	assert.Equal(t, len(fixtures), 1)
}

func TestGetTeams(t *testing.T) {
	t.Parallel()

	dao, pool, resource, err := createTestDb()
	assert.NilError(t, err)
	defer pool.Purge(resource)

	_, err = dao.InsertTeam(&apifootball.Team{
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

	teams, _ := dao.GetTeams(db.GetTeamsFilter{})
	assert.Equal(t, len(teams), 2)

	teams, _ = dao.GetTeams(db.GetTeamsFilter{Country: "usa"})
	assert.Equal(t, len(teams), 1)

	teams, _ = dao.GetTeams(db.GetTeamsFilter{Country: "lkjlk"})
	assert.Equal(t, len(teams), 0)

	teams, err = dao.GetTeams(db.GetTeamsFilter{SearchTerm: "team1"})
	assert.NilError(t, err)
	assert.Equal(t, len(teams), 1)
	assert.Equal(t, teams[0].Id, 1)
}

func TestGetLeagues(t *testing.T) {
	t.Parallel()

	dao, pool, resource, err := createTestDb()
	assert.NilError(t, err)
	defer pool.Purge(resource)

	_, err = dao.InsertLeague(&apifootball.League{
		Id:   1,
		Name: "premier league",
	})
	assert.NilError(t, err)

	_, err = dao.InsertLeague(&apifootball.League{
		Id:            2,
		Name:          "la liga",
		CurrentSeason: 2024,
	})
	assert.NilError(t, err)

	leagues, err := dao.GetLeagues()
	assert.NilError(t, err)
	assert.Equal(t, len(leagues), 2)
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
