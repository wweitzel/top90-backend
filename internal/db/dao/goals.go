package dao

import (
	"strings"

	"github.com/google/uuid"
	"github.com/wweitzel/top90/internal/db/dao/query"
	db "github.com/wweitzel/top90/internal/db/models"
)

func (dao *PostgresDAO) CountGoals(filter db.GetGoalsFilter) (int, error) {
	query, args := query.CountGoals(filter)
	var count int
	err := dao.DB.Get(&count, query, args...)
	return count, err
}

func (dao *PostgresDAO) GoalExists(redditFullname string) (bool, error) {
	query, args := query.GoalExists(redditFullname)
	var count int
	err := dao.DB.Get(&count, query, args...)
	return count > 0, err
}

func (dao *PostgresDAO) GetGoals(pagination db.Pagination, filter db.GetGoalsFilter) ([]db.Goal, error) {
	query, args := query.GetGoals(pagination, filter)
	var goals []db.Goal
	err := dao.DB.Select(&goals, query, args...)
	return goals, err
}

func (dao *PostgresDAO) GetGoal(id string) (db.Goal, error) {
	query, args := query.GetGoal(id)
	var goal db.Goal
	err := dao.DB.Get(&goal, query, args...)
	return goal, err
}

func (dao *PostgresDAO) GetNewestGoal() (db.Goal, error) {
	pagination := db.Pagination{
		Skip:  0,
		Limit: 1,
	}
	newestDbGoals, err := dao.GetGoals(pagination, db.GetGoalsFilter{})
	if err != nil {
		return db.Goal{}, err
	}

	var newestDbGoal db.Goal
	if len(newestDbGoals) > 0 {
		newestDbGoal = newestDbGoals[0]
	}
	return newestDbGoal, nil
}

func (dao *PostgresDAO) InsertGoal(goal *db.Goal) (*db.Goal, error) {
	id := uuid.NewString()
	id = strings.Replace(id, "-", "", -1)
	goal.Id = id
	query, args := query.InsertGoal(goal)

	var insertedGoal db.Goal
	err := dao.DB.QueryRowx(query, args...).StructScan(&insertedGoal)
	return &insertedGoal, err
}

// UpdateGoal updates the goal with primary key = id.
// It will update any fields that are set on goalUpdate that it can update.
// You should only set fields on goalUpdate that you actually want to be updated.
func (dao *PostgresDAO) UpdateGoal(id string, goalUpdate db.Goal) (db.Goal, error) {
	query, args := query.UpdateGoal(id, goalUpdate)

	var updatedGoal db.Goal
	err := dao.DB.QueryRowx(query, args...).StructScan(&updatedGoal)
	return updatedGoal, err
}
