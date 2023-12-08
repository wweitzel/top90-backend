package dao

import (
	"database/sql"

	"github.com/wweitzel/top90/internal/db/dao/query"
	db "github.com/wweitzel/top90/internal/db/models"
)

func (dao *PostgresDAO) GetLeagues() ([]db.League, error) {
	query := query.GetLeagues()

	var leagues []db.League
	rows, err := dao.DB.Query(query)
	if err != nil {
		return leagues, err
	}
	defer rows.Close()

	for rows.Next() {
		var currentSeason sql.NullInt64
		var league db.League
		err := rows.Scan(&league.Id, &league.Name, &league.Type, &league.Logo, &league.CreatedAt, &currentSeason)
		if err != nil {
			return leagues, err
		}
		league.CurrentSeason = int(currentSeason.Int64)
		leagues = append(leagues, league)
	}

	return leagues, nil
}

func (dao *PostgresDAO) InsertLeague(league *db.League) (*db.League, error) {
	query := query.InsertLeague(league)

	currentSeason := sql.NullInt64{
		Int64: int64(league.CurrentSeason),
		Valid: league.CurrentSeason != 0,
	}

	row := dao.DB.QueryRow(
		query, league.Id, league.Name, league.Type, league.Logo, currentSeason,
	)

	err := row.Scan(&league.Id, &league.Name, &league.Type, &league.Logo, &league.CreatedAt, &currentSeason)
	if err != nil {
		return league, err
	}

	league.CurrentSeason = int(currentSeason.Int64)

	return league, nil
}

// UpdateLeague updates the league with primary key = id.
// It will update fields that are set on leagueUpdate that it can update.
// You should only set fields on goalUpdate that you actually want to be updated.
func (dao *PostgresDAO) UpdateLeague(id int, leagueUpdate db.League) (db.League, error) {
	query, args := query.UpdateLeague(id, leagueUpdate)
	row := dao.DB.QueryRow(query, args...)

	var currentSeason sql.NullInt64
	var updatedLeague db.League

	err := row.Scan(&updatedLeague.Id, &updatedLeague.Name, &updatedLeague.Type, &updatedLeague.Logo, &updatedLeague.CreatedAt, &currentSeason)
	if err != nil {
		return updatedLeague, err
	}

	updatedLeague.CurrentSeason = int(currentSeason.Int64)
	return updatedLeague, nil
}
