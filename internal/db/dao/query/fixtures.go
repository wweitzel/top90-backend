package query

import (
	"fmt"
	"time"

	db "github.com/wweitzel/top90/internal/db/models"
)

func GetFixture(id int) string {
	query :=
		`SELECT
			fixtures.id,
			fixtures.referee,
			fixtures.date,
			fixtures.home_team_id,
			fixtures.away_team_id,
			fixtures.league_id,
			fixtures.season,
			fixtures.created_at,
			home_teams.name as home_team_name,
			home_teams.logo as home_team_logo,
			away_teams.name as away_team_name,
			away_teams.logo as away_team_logo
			FROM fixtures
		JOIN teams home_teams ON home_teams.id=fixtures.home_team_id
		JOIN teams away_teams ON away_teams.id=fixtures.away_team_id
		WHERE fixtures.id = $1 ORDER BY date ASC`
	return query
}

func GetFixtures(filter db.GetFixturesFilter) (string, []any) {
	var args []any
	whereClause, args := getFixturesWhereClause(filter, args)
	query := fmt.Sprintf(
		`SELECT
		fixtures.id,
		fixtures.referee,
		fixtures.date,
		fixtures.home_team_id,
		fixtures.away_team_id,
		fixtures.league_id,
		fixtures.season,
		fixtures.created_at,
			home_teams.name as home_team_name,
			home_teams.logo as home_team_logo,
			away_teams.name as away_team_name,
			away_teams.logo as away_team_logo
			FROM fixtures
			JOIN teams home_teams ON home_teams.id=fixtures.home_team_id
			JOIN teams away_teams ON away_teams.id=fixtures.away_team_id
		WHERE %s ORDER BY date ASC`,
		whereClause)
	return query, args
}

func InsertFixture(fixture *db.Fixture) (string, []any) {
	query := "INSERT INTO fixtures (id, referee, date, home_team_id, away_team_id, league_id, season) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (id) DO UPDATE SET date = $8 RETURNING *"
	var args []any
	args = append(args, fixture.Id, fixture.Referee, time.Unix(fixture.Timestamp, 0), fixture.Teams.Home.Id, fixture.Teams.Away.Id, fixture.LeagueId, fixture.Season, fixture.Date)
	return query, args
}

func getFixturesWhereClause(filter db.GetFixturesFilter, args []any) (string, []any) {
	whereClause := ""
	p := newParams()

	if filter.LeagueId != 0 {
		whereClause += " league_id = " + p.next()
		args = append(args, filter.LeagueId)
	} else {
		whereClause += " TRUE"
	}

	if len(filter.LeagueIds) != 0 && filter.LeagueId == 0 {
		whereClause += " AND league_id IN " + p.in(filter.LeagueIds, &args)
	}

	if !filter.Date.IsZero() {
		searchStartDate := filter.Date.Add(-12 * time.Hour)
		searchEndtDate := filter.Date.Add(12 * time.Hour)

		whereClause += " AND fixtures.date BETWEEN " + p.next() + " AND " + p.next()
		args = append(args, searchStartDate)
		args = append(args, searchEndtDate)
	}

	return whereClause, args
}
