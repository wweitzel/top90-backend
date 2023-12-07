package dao

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/db/postgres/dao/query"
)

func (dao *PostgresDAO) CountGoals(filter db.GetGoalsFilter) (int, error) {
	query, args := query.CountGoals(filter)

	var count int
	err := dao.DB.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (dao *PostgresDAO) GoalExists(redditFullname string) (bool, error) {
	query, args := query.GoalExists(redditFullname)

	var count int
	err := dao.DB.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (dao *PostgresDAO) GetGoals(pagination db.Pagination, filter db.GetGoalsFilter) ([]top90.Goal, error) {
	query, args := query.GetGoals(pagination, filter)

	var goals []top90.Goal
	rows, err := dao.DB.Query(query, args...)
	if err != nil {
		return goals, err
	}
	defer rows.Close()

	for rows.Next() {
		var fixtureId sql.NullInt64
		var thumbnailS3Key sql.NullString
		var goal top90.Goal

		err := rows.Scan(&goal.Id, &goal.RedditFullname, &goal.RedditLinkUrl, &goal.RedditPostTitle, &goal.RedditPostCreatedAt, &goal.S3ObjectKey, &goal.CreatedAt, &fixtureId, &thumbnailS3Key)
		if err != nil {
			return goals, err
		}

		goal.FixtureId = int(fixtureId.Int64)
		goal.ThumbnailS3Key = thumbnailS3Key.String
		goals = append(goals, goal)
	}

	return goals, nil
}

func (dao *PostgresDAO) GetGoal(id string) (top90.Goal, error) {
	query := query.GetGoal(id)

	var goal top90.Goal
	row := dao.DB.QueryRow(query, id)

	var fixtureId sql.NullInt64
	var thumbnailS3Key sql.NullString

	err := row.Scan(&goal.Id, &goal.RedditFullname, &goal.RedditLinkUrl, &goal.RedditPostTitle, &goal.RedditPostCreatedAt, &goal.S3ObjectKey, &goal.CreatedAt, &fixtureId, &thumbnailS3Key)
	if err != nil {
		return goal, err
	}

	goal.FixtureId = int(fixtureId.Int64)
	goal.ThumbnailS3Key = thumbnailS3Key.String

	return goal, nil
}

func (dao *PostgresDAO) GetNewestGoal() (top90.Goal, error) {
	pagination := db.Pagination{
		Skip:  0,
		Limit: 1,
	}
	newestDbGoals, err := dao.GetGoals(pagination, db.GetGoalsFilter{})
	if err != nil {
		return top90.Goal{}, err
	}

	var newestDbGoal top90.Goal
	if len(newestDbGoals) > 0 {
		newestDbGoal = newestDbGoals[0]
	}

	return newestDbGoal, nil
}

func (dao *PostgresDAO) InsertGoal(goal *top90.Goal) (*top90.Goal, error) {
	id := uuid.NewString()
	id = strings.Replace(id, "-", "", -1)

	query := query.InsertGoal(goal)

	fixtureId := sql.NullInt64{
		Int64: int64(goal.FixtureId),
		Valid: goal.FixtureId != 0,
	}

	thumbnailS3Key := sql.NullString{
		String: goal.ThumbnailS3Key,
		Valid:  goal.ThumbnailS3Key != "",
	}

	row := dao.DB.QueryRow(
		query, id, goal.RedditFullname, goal.RedditLinkUrl, goal.RedditPostTitle, goal.RedditPostCreatedAt, goal.S3ObjectKey, fixtureId, thumbnailS3Key,
		goal.S3ObjectKey, goal.ThumbnailS3Key,
	)

	err := row.Scan(&goal.Id, &goal.RedditFullname, &goal.RedditLinkUrl, &goal.RedditPostTitle, &goal.RedditPostCreatedAt, &goal.S3ObjectKey, &goal.CreatedAt, &fixtureId, &thumbnailS3Key)
	if err != nil {
		return goal, err
	}

	goal.FixtureId = int(fixtureId.Int64)

	return goal, nil
}

// UpdateGoal updates the goal with primary key = id.
// It will update any fields that are set on goalUpdate that it can update.
// You should only set fields on goalUpdate that you actually want to be updated.
func (dao *PostgresDAO) UpdateGoal(id string, goalUpdate top90.Goal) (top90.Goal, error) {
	query, args := query.UpdateGoal(id, goalUpdate)
	row := dao.DB.QueryRow(query, args...)

	var fixtureId sql.NullInt64
	var thumbanilS3Key sql.NullString
	var updatedGoal top90.Goal

	err := row.Scan(&updatedGoal.Id, &updatedGoal.RedditFullname, &updatedGoal.RedditLinkUrl, &updatedGoal.RedditPostTitle, &updatedGoal.RedditPostCreatedAt, &updatedGoal.S3ObjectKey, &updatedGoal.CreatedAt, &fixtureId, &thumbanilS3Key)
	if err != nil {
		return updatedGoal, err
	}

	updatedGoal.FixtureId = int(fixtureId.Int64)
	updatedGoal.ThumbnailS3Key = thumbanilS3Key.String
	return updatedGoal, nil
}
