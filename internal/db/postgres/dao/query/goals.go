package query

import (
	"fmt"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/db"
)

func CountGoals(filter db.GetGoalsFilter) (string, []any) {
	filter.SearchTerm = "%" + filter.SearchTerm + "%"

	query := fmt.Sprintf("SELECT count(*) FROM %s", tableNames.Goals)

	var variableCount int
	var args []any
	query, args = addGetGoalsJoinAndWhere(query, args, filter, &variableCount)

	return query, args
}

func GetGoal(id string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s = $1", tableNames.Goals, goalColumns.Id)
}

func GetGoals(pagination db.Pagination, filter db.GetGoalsFilter) (string, []any) {
	filter.SearchTerm = "%" + filter.SearchTerm + "%"

	if pagination.Limit == 0 {
		pagination.Limit = 10
	}

	query := fmt.Sprintf("SELECT %s.* FROM %s", tableNames.Goals, tableNames.Goals)

	var variableCount int
	var args []any
	query, args = addGetGoalsJoinAndWhere(query, args, filter, &variableCount)

	variableCount++
	query = query + fmt.Sprintf(" ORDER BY %s.%s DESC OFFSET $%d LIMIT $%d", tableNames.Goals, goalColumns.RedditPostCreatedAt, variableCount, variableCount+1)
	args = append(args, pagination.Skip)
	args = append(args, pagination.Limit)

	return query, args
}

func GoalExists(redditFullname string) (string, []any) {
	query := fmt.Sprintf("SELECT count(*) FROM %s where %s = $1", tableNames.Goals, goalColumns.RedditFullname)
	var args []any
	args = append(args, redditFullname)
	return query, args
}

func InsertGoal(goal *top90.Goal) string {
	return fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (%s) DO UPDATE SET %s = $9, %s = $10 RETURNING *",
		tableNames.Goals,
		goalColumns.Id, goalColumns.RedditFullname, goalColumns.RedditLinkUrl, goalColumns.RedditPostTitle, goalColumns.RedditPostCreatedAt, goalColumns.S3ObjectKey, goalColumns.FixtureId, goalColumns.ThumbnailS3Key,
		goalColumns.RedditFullname,
		goalColumns.S3ObjectKey, goalColumns.ThumbnailS3Key,
	)
}

func UpdateGoal(id string, goalUpdate top90.Goal) (string, []any) {
	var args []any

	query := fmt.Sprintf("UPDATE %s SET ", tableNames.Goals)

	variableCount := 0

	if goalUpdate.FixtureId != 0 {
		variableCount += 1
		query = query + fmt.Sprintf("%s = $%d", goalColumns.FixtureId, variableCount)
		args = append(args, goalUpdate.FixtureId)
	}

	if goalUpdate.ThumbnailS3Key != "" {
		variableCount += 1
		if variableCount == 1 {
			query = query + fmt.Sprintf("%s = $%d", goalColumns.ThumbnailS3Key, variableCount)
		} else {
			query = query + fmt.Sprintf(", %s = $%d", goalColumns.ThumbnailS3Key, variableCount)
		}
		args = append(args, goalUpdate.ThumbnailS3Key)
	}

	variableCount += 1
	query = query + fmt.Sprintf(" WHERE %s = $%d", goalColumns.Id, variableCount)
	args = append(args, id)

	query = query + " RETURNING *"

	return query, args
}

func addGetGoalsJoinAndWhere(query string, args []any, filter db.GetGoalsFilter, variableCount *int) (string, []any) {
	if filter.LeagueId != 0 || filter.Season != 0 || filter.TeamId != 0 || filter.FixtureId != 0 {
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

	if filter.FixtureId != 0 {
		*variableCount++
		query = query + fmt.Sprintf(" AND %s.%s = $%d", tableNames.Fixtures, fixtureColumns.Id, *variableCount)
		args = append(args, filter.FixtureId)
	}

	return query, args
}
