package db

import "github.com/wweitzel/top90/internal/clients/apifootball"

func (dao *PostgresDAO) GetLeagues() ([]apifootball.League, error) {
	query := getLeaguesQuery()

	var leagues []apifootball.League
	rows, err := dao.DB.Query(query)
	if err != nil {
		return leagues, err
	}
	defer rows.Close()

	for rows.Next() {
		var league apifootball.League
		err := rows.Scan(&league.Id, &league.Name, &league.Type, &league.Logo, &league.CreatedAt)
		if err != nil {
			return leagues, err
		}
		leagues = append(leagues, league)
	}

	return leagues, nil
}

func (dao *PostgresDAO) InsertLeague(league *apifootball.League) (*apifootball.League, error) {
	query := insertLeagueQuery(league)

	row := dao.DB.QueryRow(
		query, league.Id, league.Name, league.Type, league.Logo,
	)

	err := row.Scan(&league.Id, &league.Name, &league.Type, &league.Logo, &league.CreatedAt)
	if err != nil {
		return league, err
	}

	return league, nil
}
