package query

import (
	"fmt"

	db "github.com/wweitzel/top90/internal/db/models"
)

func CountGoals(filter db.GetGoalsFilter) (string, []any) {
	filter.SearchTerm = "%" + filter.SearchTerm + "%"
	query := "SELECT count(*) FROM goals"

	var variableCount int
	var args []any
	query, args = addGetGoalsJoinAndWhere(query, args, filter, &variableCount)
	return query, args
}

func GetGoal(id string) string {
	return "SELECT * FROM goals WHERE id = $1"
}

func GetGoals(pagination db.Pagination, filter db.GetGoalsFilter) (string, []any) {
	filter.SearchTerm = "%" + filter.SearchTerm + "%"
	if pagination.Limit == 0 {
		pagination.Limit = 10
	}

	query := "SELECT goals.* FROM goals"

	var variableCount int
	var args []any
	query, args = addGetGoalsJoinAndWhere(query, args, filter, &variableCount)

	variableCount++
	query = query + fmt.Sprintf(" ORDER BY goals.reddit_post_created_at DESC OFFSET $%d LIMIT $%d", variableCount, variableCount+1)
	args = append(args, pagination.Skip)
	args = append(args, pagination.Limit)
	return query, args
}

func GoalExists(redditFullname string) (string, []any) {
	query := "SELECT count(*) FROM goals where reddit_fullname = $1"
	var args []any
	args = append(args, redditFullname)
	return query, args
}

func InsertGoal(goal *db.Goal) (string, []any) {
	query := `
		INSERT INTO goals (id, reddit_fullname, reddit_link_url, reddit_post_title, reddit_post_created_at, s3_object_key, fixture_id, thumbnail_s3_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (reddit_fullname) DO UPDATE SET s3_object_key = $9, thumbnail_s3_key = $10
		RETURNING *`
	var args []any
	args = append(args,
		goal.Id,
		goal.RedditFullname,
		goal.RedditLinkUrl,
		goal.RedditPostTitle,
		goal.RedditPostCreatedAt,
		goal.S3ObjectKey,
		goal.FixtureId,
		goal.ThumbnailS3Key,
		goal.S3ObjectKey,
		goal.ThumbnailS3Key)
	return query, args
}

func UpdateGoal(id string, goalUpdate db.Goal) (string, []any) {
	var args []any
	query := "UPDATE goals SET "

	variableCount := 0
	if goalUpdate.FixtureId != 0 {
		variableCount += 1
		query = query + fmt.Sprintf("fixture_id = $%d", variableCount)
		args = append(args, goalUpdate.FixtureId)
	}
	if goalUpdate.ThumbnailS3Key != "" {
		variableCount += 1
		if variableCount == 1 {
			query = query + fmt.Sprintf("thumbnail_s3_key = $%d", variableCount)
		} else {
			query = query + fmt.Sprintf(", thumbnail_s3_key = $%d", variableCount)
		}
		args = append(args, goalUpdate.ThumbnailS3Key)
	}

	variableCount += 1
	query = query + fmt.Sprintf(" WHERE id = $%d", variableCount)
	args = append(args, id)
	query = query + " RETURNING *"
	return query, args
}

func addGetGoalsJoinAndWhere(query string, args []any, filter db.GetGoalsFilter, variableCount *int) (string, []any) {
	if filter.LeagueId != 0 || filter.Season != 0 || filter.TeamId != 0 || filter.FixtureId != 0 {
		query = query + " JOIN fixtures on goals.fixture_id = fixtures.id"
	}

	*variableCount = 1
	query = query + fmt.Sprintf(" WHERE goals.reddit_post_title ILIKE $%d", *variableCount)
	args = append(args, filter.SearchTerm)

	if filter.LeagueId != 0 {
		*variableCount++
		query = query + fmt.Sprintf(" AND fixtures.league_id = $%d", *variableCount)
		args = append(args, filter.LeagueId)
	}
	if filter.Season != 0 {
		*variableCount++
		query = query + fmt.Sprintf(" AND fixtures.season = $%d", *variableCount)
		args = append(args, filter.Season)
	}
	if filter.TeamId != 0 {
		*variableCount++
		query = query + fmt.Sprintf(" AND (fixtures.home_team_id = $%d OR fixtures.away_team_id = $%d)", *variableCount, *variableCount+1)
		*variableCount++
		args = append(args, filter.TeamId)
		args = append(args, filter.TeamId)
	}
	if filter.FixtureId != 0 {
		*variableCount++
		query = query + fmt.Sprintf(" AND fixtures.id = $%d", *variableCount)
		args = append(args, filter.FixtureId)
	}
	return query, args
}
