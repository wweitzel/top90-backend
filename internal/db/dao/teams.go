package dao

import (
	"github.com/wweitzel/top90/internal/db/dao/query"
	db "github.com/wweitzel/top90/internal/db/models"
)

func (dao *PostgresDAO) CountTeams() (int, error) {
	query := query.CountTeams()
	var count int
	err := dao.DB.Get(&count, query)
	return count, err
}

func (dao *PostgresDAO) GetTeams(filter db.GetTeamsFilter) ([]db.Team, error) {
	query, args := query.GetTeams(filter)
	var teams []db.Team
	err := dao.DB.Select(&teams, query, args...)
	return teams, err
}

func (dao *PostgresDAO) GetTeamsForLeagueAndSeason(leagueId, season int) ([]db.Team, error) {
	query, args := query.GetTeamsForLeagueAndSeason(leagueId, season)
	var teams []db.Team
	err := dao.DB.Select(&teams, query, args...)
	return teams, err
}

func (dao *PostgresDAO) InsertTeam(team *db.Team) (*db.Team, error) {
	query, args := query.InsertTeamQuery(team)
	var insertedTeam db.Team
	err := dao.DB.QueryRowx(query, args...).StructScan(&insertedTeam)
	return &insertedTeam, err
}
