package query

import (
	"fmt"
	"time"

	db "github.com/wweitzel/top90/internal/db/models"
)

func GetFixture(id int) string {
	whereClause := fmt.Sprintf("%s.%s = $1", tableNames.Fixtures, fixtureColumns.Id)
	query := fmt.Sprintf(
		`SELECT
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			home_teams.name as home_team_name,
			home_teams.logo as home_team_logo,
			away_teams.name as away_team_name,
			away_teams.logo as away_team_logo
			FROM %s
		JOIN %s home_teams ON home_teams.%s=%s
		JOIN %s away_teams ON away_teams.%s=%s
		WHERE %s ORDER BY %s ASC`,
		tableNames.Fixtures+"."+fixtureColumns.Id,
		tableNames.Fixtures+"."+fixtureColumns.Referee,
		tableNames.Fixtures+"."+fixtureColumns.Date,
		tableNames.Fixtures+"."+fixtureColumns.HomeTeamId,
		tableNames.Fixtures+"."+fixtureColumns.AwayTeamId,
		tableNames.Fixtures+"."+fixtureColumns.LeagueId,
		tableNames.Fixtures+"."+fixtureColumns.Season,
		tableNames.Fixtures+"."+fixtureColumns.CreatedAt,
		tableNames.Fixtures,
		tableNames.Teams,
		teamColumns.Id,
		tableNames.Fixtures+"."+fixtureColumns.HomeTeamId,
		tableNames.Teams,
		teamColumns.Id,
		tableNames.Fixtures+"."+fixtureColumns.AwayTeamId,
		whereClause,
		fixtureColumns.Date)
	return query
}

func GetFixtures(filter db.GetFixturesFilter) (string, []any) {
	var args []any
	whereClause, args := getFixturesWhereClause(filter, args)
	query := fmt.Sprintf(
		`SELECT
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			home_teams.name as home_team_name,
			home_teams.logo as home_team_logo,
			away_teams.name as away_team_name,
			away_teams.logo as away_team_logo
			FROM %s
		JOIN %s home_teams ON home_teams.%s=%s
		JOIN %s away_teams ON away_teams.%s=%s
		WHERE %s ORDER BY %s ASC`,
		tableNames.Fixtures+"."+fixtureColumns.Id,
		tableNames.Fixtures+"."+fixtureColumns.Referee,
		tableNames.Fixtures+"."+fixtureColumns.Date,
		tableNames.Fixtures+"."+fixtureColumns.HomeTeamId,
		tableNames.Fixtures+"."+fixtureColumns.AwayTeamId,
		tableNames.Fixtures+"."+fixtureColumns.LeagueId,
		tableNames.Fixtures+"."+fixtureColumns.Season,
		tableNames.Fixtures+"."+fixtureColumns.CreatedAt,
		tableNames.Fixtures,
		tableNames.Teams,
		teamColumns.Id,
		tableNames.Fixtures+"."+fixtureColumns.HomeTeamId,
		tableNames.Teams,
		teamColumns.Id,
		tableNames.Fixtures+"."+fixtureColumns.AwayTeamId,
		whereClause,
		fixtureColumns.Date)
	return query, args
}

func InsertFixture(fixture *db.Fixture) (string, []any) {
	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (%s) DO UPDATE SET %s = $8 RETURNING *",
		tableNames.Fixtures,
		fixtureColumns.Id, fixtureColumns.Referee, fixtureColumns.Date, fixtureColumns.HomeTeamId, fixtureColumns.AwayTeamId, fixtureColumns.LeagueId, fixtureColumns.Season,
		fixtureColumns.Id,
		fixtureColumns.Date,
	)
	var args []any
	args = append(args, fixture.Id, fixture.Referee, time.Unix(fixture.Timestamp, 0), fixture.Teams.Home.Id, fixture.Teams.Away.Id, fixture.LeagueId, fixture.Season, fixture.Date)
	return query, args
}

func getFixturesWhereClause(filter db.GetFixturesFilter, args []any) (string, []any) {
	whereClause := ""

	if filter.LeagueId != 0 {
		whereClause = whereClause + fmt.Sprintf(" %s = $1", fixtureColumns.LeagueId)
		args = append(args, filter.LeagueId)
	} else {
		whereClause = whereClause + " $1"
		args = append(args, "TRUE")
	}

	if !filter.Date.IsZero() {
		searchStartDate := filter.Date.Add(-12 * time.Hour)
		searchEndtDate := filter.Date.Add(12 * time.Hour)

		whereClause = whereClause + fmt.Sprintf(" AND %s BETWEEN $2 AND $3",
			tableNames.Fixtures+"."+fixtureColumns.Date,
		)
		args = append(args, searchStartDate)
		args = append(args, searchEndtDate)
	}
	return whereClause, args
}
