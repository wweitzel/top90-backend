package db

import (
	"fmt"

	"github.com/wweitzel/top90/internal/apifootball"
)

func getLeaguesQuery() string {
	return fmt.Sprintf("SELECT * FROM %s ORDER BY %s ASC", tableNames.Leagues, leagueColumns.Name)
}

func insertLeagueQuery(league *apifootball.League) string {
	return fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES ($1, $2, $3, $4) ON CONFLICT (%s) DO NOTHING RETURNING *",
		tableNames.Leagues,
		leagueColumns.Id, leagueColumns.Name, leagueColumns.Type, leagueColumns.Logo,
		leagueColumns.Id,
	)
}
