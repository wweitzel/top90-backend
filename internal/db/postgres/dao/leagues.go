package dao

import (
	"database/sql"

	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/db/postgres/dao/query"
)

func (dao *PostgresDAO) GetLeagues() ([]apifootball.League, error) {
	query := query.GetLeagues()

	var leagues []apifootball.League
	rows, err := dao.DB.Query(query)
	if err != nil {
		return leagues, err
	}
	defer rows.Close()

	for rows.Next() {
		var currentSeason sql.NullInt64
		var league apifootball.League
		err := rows.Scan(&league.Id, &league.Name, &league.Type, &league.Logo, &league.CreatedAt, &currentSeason)
		if err != nil {
			return leagues, err
		}
		league.CurrentSeason = int(currentSeason.Int64)
		leagues = append(leagues, league)
	}

	return leagues, nil
}

func (dao *PostgresDAO) InsertLeague(league *apifootball.League) (*apifootball.League, error) {
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
