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

func GetGoal(id string) (string, []any) {
	query := "SELECT * FROM goals WHERE id = $1"
	return query, []any{id}
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
		INSERT INTO goals (id, reddit_fullname, reddit_link_url, reddit_post_title, reddit_post_created_at, s3_object_key, fixture_id, thumbnail_s3_key, type, type_detail, player_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (reddit_fullname) DO NOTHING RETURNING *`
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
		goal.Type,
		goal.TypeDetail,
		goal.PlayerId)
	return query, args
}

func UpdateGoal(id string, goalUpdate db.Goal) (string, []any) {
	var args []any
	query := "UPDATE goals SET "

	variableCount := 0
	if goalUpdate.FixtureId != 0 {
		variableCount++
		query = query + fmt.Sprintf("fixture_id = $%d", variableCount)
		args = append(args, goalUpdate.FixtureId)
	}
	if goalUpdate.ThumbnailS3Key != "" {
		variableCount++
		query = addComma(query, variableCount != 1)
		query = query + fmt.Sprintf("thumbnail_s3_key = $%d", variableCount)
		args = append(args, goalUpdate.ThumbnailS3Key)
	}
	if goalUpdate.Type != "" {
		variableCount++
		query = addComma(query, variableCount != 1)
		query = query + fmt.Sprintf("type = $%d", variableCount)
		args = append(args, goalUpdate.Type)
	}
	if goalUpdate.TypeDetail != "" {
		variableCount++
		query = addComma(query, variableCount != 1)
		query = query + fmt.Sprintf("type_detail = $%d", variableCount)
		args = append(args, goalUpdate.TypeDetail)
	}
	if goalUpdate.PlayerId != 0 {
		variableCount++
		query = addComma(query, variableCount != 1)
		query = query + fmt.Sprintf("player_id = $%d", variableCount)
		args = append(args, goalUpdate.PlayerId)
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

	if filter.PlayerId != 0 {
		*variableCount++
		query = query + fmt.Sprintf(" AND goals.player_id = $%d", *variableCount)
		args = append(args, filter.PlayerId)
	}

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

func addComma(query string, condition bool) string {
	if condition {
		query = query + ", "
	}
	return query
}
