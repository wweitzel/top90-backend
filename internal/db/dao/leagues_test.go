package dao

import (
	"testing"

	db "github.com/wweitzel/top90/internal/db/models"
	"gotest.tools/v3/assert"
)

func TestGetLeagues(t *testing.T) {
	t.Parallel()

	dao, pool, resource, err := createTestDb()
	assert.NilError(t, err)
	defer pool.Purge(resource)

	_, err = dao.InsertLeague(&db.League{
		Id:   1,
		Name: "premier league",
	})
	assert.NilError(t, err)

	_, err = dao.InsertLeague(&db.League{
		Id:            2,
		Name:          "la liga",
		CurrentSeason: 2024,
	})
	assert.NilError(t, err)

	leagues, err := dao.GetLeagues()
	assert.NilError(t, err)
	assert.Equal(t, len(leagues), 2)
}

func TestUpdateLeague(t *testing.T) {
	t.Parallel()

	dao, pool, resource, err := createTestDb()
	assert.NilError(t, err)
	defer pool.Purge(resource)

	league, err := dao.InsertLeague(&db.League{
		Id:   1,
		Name: "premier league",
	})
	assert.NilError(t, err)

	leagueUpdate := db.League{CurrentSeason: 2024}
	updatedLeague, err := dao.UpdateLeague(league.Id, leagueUpdate)
	assert.NilError(t, err)
	assert.Equal(t, updatedLeague.CurrentSeason, 2024)

	leagueUpdate.CurrentSeason = 2025
	updatedLeague, err = dao.UpdateLeague(league.Id, leagueUpdate)
	assert.NilError(t, err)
	assert.Equal(t, updatedLeague.CurrentSeason, 2025)
}
