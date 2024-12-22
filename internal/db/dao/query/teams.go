package query

import (
	"fmt"

	db "github.com/wweitzel/top90/internal/db/models"
)

func CountTeams() string {
	return "SELECT count(*) FROM teams"
}

func GetTeams(filter db.GetTeamsFilter) (string, []any) {
	var args []any
	whereClause, args := getTeamsWhereClause(filter, args)
	query := fmt.Sprintf("SELECT * FROM teams WHERE %s ORDER BY name ASC", whereClause)
	return query, args
}

func GetTeamsForLeagueAndSeason(leagueId, season int) (string, []any) {
	var args []any
	query := "SELECT * from teams WHERE id in ("

	if leagueId != 0 && season != 0 {
		query = query + "SELECT home_team_id FROM fixtures WHERE league_id = $1 AND season = $2 UNION "
		query = query + "SELECT away_team_id FROM fixtures WHERE league_id = $1 AND season = $2"
		args = append(args, leagueId, season)
	}
	if leagueId != 0 && season == 0 {
		query = query + "SELECT home_team_id FROM fixtures WHERE league_id = $1 UNION "
		query = query + "SELECT away_team_id FROM fixtures WHERE league_id = $1"
		args = append(args, leagueId)
	}
	if leagueId == 0 && season != 0 {
		query = query + "SELECT home_team_id FROM fixtures WHERE season = $1 UNION "
		query = query + "SELECT away_team_id FROM fixtures WHERE season = $1"
		args = append(args, season)
	}
	if leagueId == 0 && season == 0 {
		query = query + "SELECT home_team_id FROM fixtures UNION "
		query = query + "SELECT away_team_id FROM fixtures"
	}

	query = query + (")")
	return query, args
}

func InsertTeamQuery(team *db.Team) (string, []any) {
	query := "INSERT INTO teams (id, name, code, country, founded, national, logo, aliases) "
	query = query + "VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (id) DO NOTHING RETURNING *"
	var args []any
	args = append(args, team.Id, team.Name, team.Code, team.Country, team.Founded, team.National, team.Logo, &team.Aliases)
	return query, args
}

func getTeamsWhereClause(filter db.GetTeamsFilter, args []any) (string, []any) {
	p := newParams()
	whereClause := p.next()
	args = append(args, "TRUE")

	if filter.Country != "" {
		whereClause = whereClause + " AND country = " + p.next()
		args = append(args, filter.Country)
	}
	if filter.SearchTerm != "" {
		whereClause = whereClause + " AND name ILIKE " + p.next()
		args = append(args, fmt.Sprintf("%%%s%%", filter.SearchTerm))
	}
	return whereClause, args
}
