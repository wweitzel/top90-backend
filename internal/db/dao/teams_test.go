package dao

import (
	"database/sql"
	"testing"

	db "github.com/wweitzel/top90/internal/db/models"
	"gotest.tools/v3/assert"
)

func TestGetTeams(t *testing.T) {
	t.Parallel()

	dao, pool, resource, err := createTestDb()
	assert.NilError(t, err)
	defer pool.Purge(resource)

	var aliases []string
	aliases = append(aliases, "united states")
	_, err = dao.InsertTeam(&db.Team{
		Id:      1,
		Name:    "team1",
		Country: "usa",
		Aliases: aliases,
	})
	assert.NilError(t, err)

	_, err = dao.InsertTeam(&db.Team{
		Id:      2,
		Name:    "team2",
		Country: "mexico",
	})
	assert.NilError(t, err)

	teams, err := dao.GetTeams(db.GetTeamsFilter{})
	assert.NilError(t, err)
	assert.Equal(t, len(teams), 2)

	teams, err = dao.GetTeams(db.GetTeamsFilter{Country: "usa"})
	assert.NilError(t, err)
	assert.Equal(t, len(teams), 1)
	assert.Equal(t, teams[0].Aliases[0], "united states")

	teams, err = dao.GetTeams(db.GetTeamsFilter{Country: "lkjlk"})
	assert.NilError(t, err)
	assert.Equal(t, len(teams), 0)

	teams, err = dao.GetTeams(db.GetTeamsFilter{SearchTerm: "team1"})
	assert.NilError(t, err)
	assert.Equal(t, len(teams), 1)
	assert.Equal(t, teams[0].Id, 1)
}

func TestInsertTeamTwice(t *testing.T) {
	t.Parallel()

	dao, pool, resource, err := createTestDb()
	assert.NilError(t, err)
	defer pool.Purge(resource)

	_, err = dao.InsertTeam(&db.Team{
		Id: 1,
	})
	assert.NilError(t, err)

	_, err = dao.InsertTeam(&db.Team{
		Id: 1,
	})
	assert.ErrorIs(t, err, sql.ErrNoRows)
}
