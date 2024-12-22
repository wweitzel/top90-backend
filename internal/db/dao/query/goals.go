package query

import (
	"time"

	db "github.com/wweitzel/top90/internal/db/models"
)

func CountGoals(filter db.GetGoalsFilter) (string, []any) {
	filter.SearchTerm = "%" + filter.SearchTerm + "%"
	query := "SELECT count(*) FROM goals"

	var args []any
	query, args = addGetGoalsJoinAndWhere(query, args, filter)
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

	var args []any
	query, args = addGetGoalsJoinAndWhere(query, args, filter)

	p := newParamsFrom(len(args) + 1)
	query = query + " ORDER BY goals.reddit_post_created_at DESC OFFSET " + p.next() + " LIMIT " + p.next()
	args = append(args, pagination.Skip)
	args = append(args, pagination.Limit)
	return query, args
}

func GetGoalsSince(since time.Time) (string, []any) {
	query := "SELECT * FROM goals WHERE reddit_post_created_at > $1 ORDER BY reddit_post_created_at DESC"
	return query, []any{since}
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
	p := newParams()

	if goalUpdate.FixtureId != 0 {
		query += p.nextUpdate("fixture_id")
		args = append(args, goalUpdate.FixtureId)
	}
	if goalUpdate.ThumbnailS3Key != "" {
		query += p.nextUpdate("thumbnail_s3_key")
		args = append(args, goalUpdate.ThumbnailS3Key)
	}
	if goalUpdate.Type != "" {
		query += p.nextUpdate("type")
		args = append(args, goalUpdate.Type)
	}
	if goalUpdate.TypeDetail != "" {
		query += p.nextUpdate("type_detail")
		args = append(args, goalUpdate.TypeDetail)
	}
	if goalUpdate.PlayerId != 0 {
		query += p.nextUpdate("player_id")
		args = append(args, goalUpdate.PlayerId)
	}

	query += " WHERE id = " + p.next()
	args = append(args, id)
	query += " RETURNING *"
	return query, args
}

func addGetGoalsJoinAndWhere(query string, args []any, filter db.GetGoalsFilter) (string, []any) {
	if filter.LeagueId != 0 || filter.Season != 0 || filter.TeamId != 0 || filter.FixtureId != 0 {
		query = query + " JOIN fixtures on goals.fixture_id = fixtures.id"
	}

	p := newParams()
	query = query + " WHERE goals.reddit_post_title ILIKE " + p.next()
	args = append(args, filter.SearchTerm)

	if filter.PlayerId != 0 {
		query = query + " AND goals.player_id = " + p.next()
		args = append(args, filter.PlayerId)
	}

	if filter.LeagueId != 0 {
		query = query + " AND fixtures.league_id = " + p.next()
		args = append(args, filter.LeagueId)
	}
	if filter.Season != 0 {
		query = query + " AND fixtures.season = " + p.next()
		args = append(args, filter.Season)
	}
	if filter.TeamId != 0 {
		query = query + " AND (fixtures.home_team_id = " + p.next() + " OR fixtures.away_team_id = " + p.next() + ")"
		args = append(args, filter.TeamId)
		args = append(args, filter.TeamId)
	}
	if filter.FixtureId != 0 {
		query = query + " AND fixtures.id = " + p.next()
		args = append(args, filter.FixtureId)
	}
	return query, args
}
