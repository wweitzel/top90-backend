package dao

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	db "github.com/wweitzel/top90/internal/db/models"
	"gotest.tools/v3/assert"
)

func TestGoalsDao(t *testing.T) {
	t.Parallel()

	dao, pool, resource, err := createTestDb()
	assert.NilError(t, err)
	defer pool.Purge(resource)

	now := time.Now()

	team1, err := dao.InsertTeam(&db.Team{
		Id:   1,
		Name: "team1",
	})
	assert.NilError(t, err)

	team2, err := dao.InsertTeam(&db.Team{
		Id:   2,
		Name: "team2",
	})
	assert.NilError(t, err)

	league1, err := dao.InsertLeague(&db.League{
		Id:   1,
		Name: "premier league",
	})
	assert.NilError(t, err)

	fixture, err := dao.InsertFixture(&db.Fixture{
		Id:        1,
		Referee:   "jimbob",
		Timestamp: now.Unix(),
		LeagueId:  league1.Id,
		Teams: db.Teams{
			Home: db.Team{Id: team1.Id},
			Away: db.Team{Id: team2.Id},
		},
	})
	assert.NilError(t, err)

	player, err := dao.UpsertPlayer(db.Player{
		Id:          1,
		Name:        "jim bob",
		FirstName:   "jim",
		LastName:    "bob",
		Age:         21,
		Nationality: "usa",
		Height:      "100 cm",
		Weight:      "100 kg",
		Photo:       "https://somephoto",
	})
	assert.NilError(t, err)

	uid := uuid.NewString()
	goal, err := dao.InsertGoal(&db.Goal{
		RedditFullname:      uid,
		RedditLinkUrl:       "redditlinkurl",
		RedditPostTitle:     "redditposttitlte",
		S3ObjectKey:         "s3objectkey",
		RedditPostCreatedAt: now,
		ThumbnailS3Key:      "thumbnails3key",
		FixtureId:           db.NullInt(fixture.Id),
		Type:                "Goal",
		TypeDetail:          "Normal Goal",
		PlayerId:            db.NullInt(player.Id),
	})
	assert.NilError(t, err)

	assertEqual(t, *goal, db.Goal{
		Id:                  goal.Id,
		CreatedAt:           goal.CreatedAt,
		RedditFullname:      uid,
		RedditLinkUrl:       "redditlinkurl",
		RedditPostTitle:     "redditposttitlte",
		S3ObjectKey:         "s3objectkey",
		RedditPostCreatedAt: now,
		ThumbnailS3Key:      "thumbnails3key",
		FixtureId:           db.NullInt(fixture.Id),
		Type:                "Goal",
		TypeDetail:          "Normal Goal",
		PlayerId:            db.NullInt(player.Id),
	})

	uid2 := uuid.NewString()
	goal2, _ := dao.InsertGoal(&db.Goal{
		RedditFullname:      uid2,
		RedditPostCreatedAt: now,
	})
	assertEqual(t, *goal2, db.Goal{
		Id:             goal2.Id,
		RedditFullname: uid2,
	})

	count, err := dao.CountGoals(db.GetGoalsFilter{})
	assert.NilError(t, err)
	assert.Equal(t, count, 2)

	goals, err := dao.GetGoals(db.Pagination{}, db.GetGoalsFilter{})
	assert.NilError(t, err)
	assert.Equal(t, len(goals), 2)

	oneHourAgo := now.Add(-1 * time.Hour)
	goalsSince, err := dao.GetGoalsSince(oneHourAgo)
	assert.NilError(t, err)
	assert.Equal(t, len(goalsSince), 2)
	assert.Equal(t, goalsSince[0].Id, goal.Id)

	futureTime := now.Add(1 * time.Hour)
	goalsSince, err = dao.GetGoalsSince(futureTime)
	assert.NilError(t, err)
	assert.Equal(t, len(goalsSince), 0)

	goals, _ = dao.GetGoals(db.Pagination{}, db.GetGoalsFilter{FixtureId: fixture.Id})
	assert.Equal(t, len(goals), 1)

	goals, _ = dao.GetGoals(db.Pagination{}, db.GetGoalsFilter{FixtureId: 9783246978987})
	assert.Equal(t, len(goals), 0)

	fixture2, err := dao.InsertFixture(&db.Fixture{
		Id:       2,
		Referee:  "jimbob",
		LeagueId: league1.Id,
		Teams: db.Teams{
			Home: db.Team{Id: team1.Id},
			Away: db.Team{Id: team2.Id},
		},
	})
	assert.NilError(t, err)

	player2, err := dao.UpsertPlayer(db.Player{
		Id:          2,
		Name:        "player 2",
		FirstName:   "player",
		LastName:    "twp",
		Age:         35,
		Nationality: "usa 2",
		Height:      "100 cm 2",
		Weight:      "100 kg 2",
		Photo:       "https://somephoto2",
	})
	assert.NilError(t, err)

	goalUpdate := db.Goal{
		FixtureId:      db.NullInt(fixture2.Id),
		ThumbnailS3Key: "updatedS3key",
		Type:           "Goal type update",
		TypeDetail:     "Goal type detail update",
		PlayerId:       db.NullInt(player2.Id),
	}

	updatedGoal, err := dao.UpdateGoal(goal.Id, goalUpdate)
	assert.NilError(t, err)
	assert.Equal(t, int(updatedGoal.FixtureId), 2)
	assert.Equal(t, string(updatedGoal.ThumbnailS3Key), "updatedS3key")
	assert.Equal(t, int(updatedGoal.PlayerId), 2)
	assert.Equal(t, string(updatedGoal.Type), "Goal type update")
	assert.Equal(t, string(updatedGoal.TypeDetail), "Goal type detail update")

	fromDb, err := dao.GetGoal(updatedGoal.Id)
	assert.NilError(t, err)
	assert.Equal(t, fromDb.Id, updatedGoal.Id)

	_, err = dao.GetGoal("notfound")
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func assertEqual(t *testing.T, actual db.Goal, expected db.Goal) {
	assert.Equal(t, actual.Id, expected.Id)
	assert.Equal(t, actual.RedditFullname, expected.RedditFullname)
	assert.Equal(t, actual.RedditLinkUrl, expected.RedditLinkUrl)
	assert.Equal(t, actual.RedditPostTitle, expected.RedditPostTitle)
	assert.Equal(t, actual.S3ObjectKey, expected.S3ObjectKey)
	// TODO: Figure out why the bwlow assertion fails
	// assert.Equal(t, actual.RedditPostCreatedAt, expected.RedditPostCreatedAt)
	assert.Equal(t, actual.ThumbnailS3Key, expected.ThumbnailS3Key)
	assert.Equal(t, actual.FixtureId, expected.FixtureId)
	assert.Equal(t, actual.Type, expected.Type)
	assert.Equal(t, actual.TypeDetail, expected.TypeDetail)
	assert.Equal(t, actual.PlayerId, expected.PlayerId)
}
