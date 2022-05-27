package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	top90 "github.com/wweitzel/top90/internal"
)

type Top90DAO interface {
	CountGoals(GetGoalsFilter) (int, error)
	GetGoals(pagination Pagination, filter GetGoalsFilter) ([]top90.Goal, error)
	GetNewestGoal() (top90.Goal, error)
	InsertGoal(*top90.Goal) (*top90.Goal, error)
}

type PostgresDAO struct {
	DB *sql.DB
}

type Pagination struct {
	Skip  int
	Limit int
}

type GetGoalsFilter struct {
	SearchTerm string
}

func NewPostgresDAO(db *sql.DB) Top90DAO {
	return &PostgresDAO{
		DB: db,
	}
}

func (dao *PostgresDAO) CountGoals(filter GetGoalsFilter) (int, error) {
	filter.SearchTerm = "%" + filter.SearchTerm + "%"

	query := fmt.Sprintf("SELECT count(*) FROM %s WHERE reddit_post_title ILIKE $1", tableNames.Goals)

	var count int
	err := dao.DB.QueryRow(query, filter.SearchTerm).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (dao *PostgresDAO) GetGoals(pagination Pagination, filter GetGoalsFilter) ([]top90.Goal, error) {
	filter.SearchTerm = "%" + filter.SearchTerm + "%"

	if pagination.Limit == 0 {
		pagination.Limit = 10
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE reddit_post_title ILIKE $1 ORDER BY %s DESC OFFSET $2 LIMIT $3",
		tableNames.Goals, goalColumns.RedditPostCreatedAt)

	var list []top90.Goal
	rows, err := dao.DB.Query(query, filter.SearchTerm, pagination.Skip, pagination.Limit)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		var video top90.Goal
		err := rows.Scan(&video.Id, &video.RedditFullname, &video.RedditLinkUrl, &video.RedditPostTitle, &video.RedditPostCreatedAt, &video.S3ObjectKey, &video.CreatedAt)
		if err != nil {
			return list, err
		}
		list = append(list, video)
	}

	return list, nil
}

func (dao *PostgresDAO) GetNewestGoal() (top90.Goal, error) {
	pagination := Pagination{
		Skip:  0,
		Limit: 1,
	}
	newestDbVideos, err := dao.GetGoals(pagination, GetGoalsFilter{})
	if err != nil {
		return top90.Goal{}, err
	}

	var newestDbVideo top90.Goal
	if len(newestDbVideos) > 0 {
		newestDbVideo = newestDbVideos[0]
	}

	return newestDbVideo, nil
}

func (dao *PostgresDAO) InsertGoal(goal *top90.Goal) (*top90.Goal, error) {
	id := uuid.NewString()
	id = strings.Replace(id, "-", "", -1)

	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (%s) DO NOTHING RETURNING *",
		tableNames.Goals,
		goalColumns.Id, goalColumns.RedditFullname, goalColumns.RedditLinkUrl, goalColumns.RedditPostTitle, goalColumns.RedditPostCreatedAt, goalColumns.S3ObjectKey,
		goalColumns.RedditFullname,
	)

	row := dao.DB.QueryRow(
		query, id, goal.RedditFullname, goal.RedditLinkUrl, goal.RedditPostTitle, goal.RedditPostCreatedAt, goal.S3ObjectKey,
	)

	err := row.Scan(&goal.Id, &goal.RedditFullname, &goal.RedditLinkUrl, &goal.RedditPostTitle, &goal.RedditPostCreatedAt, &goal.S3ObjectKey, &goal.CreatedAt)
	if err != nil {
		return goal, err
	}

	return goal, nil
}
