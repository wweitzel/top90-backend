package dao

import (
	"github.com/wweitzel/top90/internal/db/dao/query"
	db "github.com/wweitzel/top90/internal/db/models"
)

func (dao *PostgresDAO) GetLeagues() ([]db.League, error) {
	query := query.GetLeagues()
	var leagues []db.League
	err := dao.DB.Select(&leagues, query)
	return leagues, err
}

func (dao *PostgresDAO) InsertLeague(league *db.League) (*db.League, error) {
	query, args := query.InsertLeague(league)
	var insertedLeague db.League
	err := dao.DB.QueryRowx(query, args...).StructScan(&insertedLeague)
	return &insertedLeague, err
}

// UpdateLeague updates the league with primary key = id.
// It will update fields that are set on leagueUpdate that it can update.
// You should only set fields on goalUpdate that you actually want to be updated.
func (dao *PostgresDAO) UpdateLeague(id int, leagueUpdate db.League) (db.League, error) {
	query, args := query.UpdateLeague(id, leagueUpdate)
	var updatedLeague db.League
	err := dao.DB.QueryRowx(query, args...).StructScan(&updatedLeague)
	return updatedLeague, err
}
