package dao

import (
	"testing"
	"time"

	db "github.com/wweitzel/top90/internal/db/models"
	"gotest.tools/v3/assert"
)

func TestGetFixtures(t *testing.T) {
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

	league2, err := dao.InsertLeague(&db.League{
		Id:   2,
		Name: "la liga",
	})
	assert.NilError(t, err)

	_, err = dao.InsertFixture(&db.Fixture{
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

	_, err = dao.InsertFixture(&db.Fixture{
		Id:        2,
		Referee:   "jimbob",
		Timestamp: now.AddDate(0, 0, 2).Unix(),
		LeagueId:  league2.Id,
		Teams: db.Teams{
			Home: db.Team{Id: team1.Id},
			Away: db.Team{Id: team2.Id},
		},
	})
	assert.NilError(t, err)

	fixtures, err := dao.GetFixtures(db.GetFixturesFilter{})
	assert.NilError(t, err)
	assert.Equal(t, len(fixtures), 2)

	fixtures, err = dao.GetFixtures(db.GetFixturesFilter{LeagueId: 1})
	assert.NilError(t, err)
	assert.Equal(t, len(fixtures), 1)

	fixtures, err = dao.GetFixtures(db.GetFixturesFilter{Date: time.Now()})
	assert.NilError(t, err)
	assert.Equal(t, len(fixtures), 1)

	newDate := time.Now().AddDate(0, 1, 0)
	updatedFixture, err := dao.InsertFixture(&db.Fixture{
		Id:   2,
		Date: newDate,
	})
	assert.NilError(t, err)
	assert.Equal(t, updatedFixture.Date.Unix(), newDate.Unix())
}
