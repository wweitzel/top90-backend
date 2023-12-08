package query

import (
	"fmt"

	db "github.com/wweitzel/top90/internal/db/models"
)

func GetLeagues() string {
	return fmt.Sprintf("SELECT * FROM %s ORDER BY %s ASC", tableNames.Leagues, leagueColumns.Name)
}

func InsertLeague(league *db.League) string {
	return fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (%s) DO NOTHING RETURNING *",
		tableNames.Leagues,
		leagueColumns.Id, leagueColumns.Name, leagueColumns.Type, leagueColumns.Logo, leagueColumns.CurrentSeason,
		leagueColumns.Id,
	)
}

func UpdateLeague(id int, leagueUpdate db.League) (string, []any) {
	var args []any
	query := fmt.Sprintf("UPDATE %s SET ", tableNames.Leagues)

	variableCount := 0

	if leagueUpdate.CurrentSeason != 0 {
		variableCount += 1
		query = query + fmt.Sprintf("%s = $%d", leagueColumns.CurrentSeason, variableCount)
		args = append(args, leagueUpdate.CurrentSeason)
	}

	variableCount += 1
	query = query + fmt.Sprintf(" WHERE %s = $%d", leagueColumns.Id, variableCount)
	args = append(args, id)

	query = query + " RETURNING *"

	return query, args
}
