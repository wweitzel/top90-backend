package query

import (
	"fmt"

	db "github.com/wweitzel/top90/internal/db/models"
)

func GetLeagues() string {
	return "SELECT * FROM leagues ORDER BY name ASC"
}

func InsertLeague(league *db.League) (string, []any) {
	query := "INSERT INTO leagues (id, name, type, logo, current_season) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING RETURNING *"
	var args []any
	args = append(args, league.Id, league.Name, league.Type, league.Logo, league.CurrentSeason)
	return query, args
}

func UpdateLeague(id int, leagueUpdate db.League) (string, []any) {
	var args []any
	query := "UPDATE leagues SET "

	variableCount := 0
	if leagueUpdate.CurrentSeason != 0 {
		variableCount += 1
		query = query + fmt.Sprintf("current_season = $%d", variableCount)
		args = append(args, leagueUpdate.CurrentSeason)
	}

	variableCount += 1
	query = query + fmt.Sprintf(" WHERE id = $%d", variableCount)
	args = append(args, id)

	query = query + " RETURNING *"
	return query, args
}
