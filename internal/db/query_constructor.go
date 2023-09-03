package db

import (
	"fmt"
	"time"
)

func addGetGoalsJoinAndWhere(query string, args []any, filter GetGoalsFilter, variableCount *int) (string, []any) {
	if filter.LeagueId != 0 || filter.Season != 0 || filter.TeamId != 0 {
		// Join fixtures
		query = query + fmt.Sprintf(" JOIN %s on %s.%s = %s.%s",
			tableNames.Fixtures, tableNames.Goals, goalColumns.FixtureId, tableNames.Fixtures, fixtureColumns.Id)
	}

	*variableCount = 1
	query = query + fmt.Sprintf(" WHERE %s.%s ILIKE $%d", tableNames.Goals, goalColumns.RedditPostTitle, *variableCount)
	args = append(args, filter.SearchTerm)

	if filter.LeagueId != 0 {
		*variableCount++
		query = query + fmt.Sprintf(" AND %s.%s = $%d", tableNames.Fixtures, fixtureColumns.LeagueId, *variableCount)
		args = append(args, filter.LeagueId)
	}

	if filter.Season != 0 {
		*variableCount++
		query = query + fmt.Sprintf(" AND %s.%s = $%d", tableNames.Fixtures, fixtureColumns.Season, *variableCount)
		args = append(args, filter.Season)
	}

	if filter.TeamId != 0 {
		*variableCount++
		query = query + fmt.Sprintf(" AND (%s.%s = $%d OR %s.%s = $%d)", tableNames.Fixtures, fixtureColumns.HomeTeamId, *variableCount, tableNames.Fixtures, fixtureColumns.AwayTeamId, *variableCount+1)
		*variableCount++
		args = append(args, filter.TeamId)
		args = append(args, filter.TeamId)
	}

	return query, args
}

func getFixturesWhereClause(filter GetFixuresFilter, args []any) (string, []any) {
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

		whereClause = whereClause + fmt.Sprintf(" AND %s >= $2 and %s <= $3",
			tableNames.Fixtures+"."+fixtureColumns.Date,
			tableNames.Fixtures+"."+fixtureColumns.Date,
		)
		args = append(args, searchStartDate)
		args = append(args, searchEndtDate)
	}

	return whereClause, args
}

func getTeamsWhereClause(filter GetTeamsFilter, args []any) (string, []any) {
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
