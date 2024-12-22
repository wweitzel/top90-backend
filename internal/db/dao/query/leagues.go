package query

import (
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
	p := newParams()

	if leagueUpdate.CurrentSeason != 0 {
		query += p.nextUpdate("current_season")
		args = append(args, leagueUpdate.CurrentSeason)
	}

	query += " WHERE id = " + p.next()
	args = append(args, id)

	query += " RETURNING *"
	return query, args
}
