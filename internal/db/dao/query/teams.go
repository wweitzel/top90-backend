package query

import (
	"fmt"

	db "github.com/wweitzel/top90/internal/db/models"
)

func CountTeams() string {
	return fmt.Sprintf("SELECT count(*) FROM %s", tableNames.Teams)
}

func GetTeams(filter db.GetTeamsFilter) (string, []any) {
	var args []any
	whereClause, args := getTeamsWhereClause(filter, args)
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s ORDER BY %s ASC", tableNames.Teams, whereClause, teamColumns.Name)
	return query, args
}

func GetTeamsForLeagueAndSeason(leagueId, season int) (string, []any) {
	var args []any

	// Union of home teams ids and away team ids for a given league season
	query := fmt.Sprintf("SELECT * from %s WHERE %s in (", tableNames.Teams, teamColumns.Id)

	// Teams for league id and season
	if leagueId != 0 && season != 0 {
		query = query + fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1 AND %s = $2", fixtureColumns.HomeTeamId, tableNames.Fixtures, fixtureColumns.LeagueId, fixtureColumns.Season)
		query = query + " UNION "
		query = query + fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1 AND %s = $2", fixtureColumns.AwayTeamId, tableNames.Fixtures, fixtureColumns.LeagueId, fixtureColumns.Season)
		args = append(args, leagueId, season)
	}

	// Teams for league id
	if leagueId != 0 && season == 0 {
		query = query + fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", fixtureColumns.HomeTeamId, tableNames.Fixtures, fixtureColumns.LeagueId)
		query = query + " UNION "
		query = query + fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", fixtureColumns.AwayTeamId, tableNames.Fixtures, fixtureColumns.LeagueId)
		args = append(args, leagueId)
	}

	// Teams for season
	if leagueId == 0 && season != 0 {
		query = query + fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", fixtureColumns.HomeTeamId, tableNames.Fixtures, fixtureColumns.Season)
		query = query + " UNION "
		query = query + fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", fixtureColumns.AwayTeamId, tableNames.Fixtures, fixtureColumns.Season)
		args = append(args, season)
	}

	// All teams
	if leagueId == 0 && season == 0 {
		query = query + fmt.Sprintf("SELECT %s FROM %s", fixtureColumns.HomeTeamId, tableNames.Fixtures)
		query = query + " UNION "
		query = query + fmt.Sprintf("SELECT %s FROM %s", fixtureColumns.AwayTeamId, tableNames.Fixtures)
	}

	query = query + (")")

	return query, args
}

func InsertTeamQuery(team *db.Team) (string, []any) {
	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (%s) DO NOTHING RETURNING *",
		tableNames.Teams,
		teamColumns.Id, teamColumns.Name, teamColumns.Code, teamColumns.Country, teamColumns.Founded, teamColumns.National, teamColumns.Logo, teamColumns.Aliases,
		leagueColumns.Id,
	)
	var args []any
	args = append(args, team.Id, team.Name, team.Code, team.Country, team.Founded, team.National, team.Logo, &team.Aliases)
	return query, args
}

func getTeamsWhereClause(filter db.GetTeamsFilter, args []any) (string, []any) {
	whereClause := ""

	whereClause = whereClause + "$1"
	args = append(args, "TRUE")

	if filter.Country != "" {
		whereClause = whereClause + fmt.Sprintf(" AND %s = $%d", teamColumns.Country, len(args)+1)
		args = append(args, filter.Country)
	}

	if filter.SearchTerm != "" {
		whereClause = whereClause + fmt.Sprintf(" AND %s ILIKE $%d", teamColumns.Name, len(args)+1)
		args = append(args, fmt.Sprintf("%%%s%%", filter.SearchTerm))
	}

	return whereClause, args
}
