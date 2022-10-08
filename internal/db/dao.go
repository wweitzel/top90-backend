package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/apifootball"
)

type Top90DAO interface {
	CountGoals(GetGoalsFilter) (int, error)
	CountTeams() (int, error)
	GetGoals(pagination Pagination, filter GetGoalsFilter) ([]top90.Goal, error)
	GetLeagues() ([]apifootball.League, error)
	GetNewestGoal() (top90.Goal, error)
	GetTeams() ([]apifootball.Team, error)
	InsertGoal(*top90.Goal) (*top90.Goal, error)
	InsertLeague(*apifootball.League) (*apifootball.League, error)
	InsertTeam(*apifootball.Team) (*apifootball.Team, error)
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

func (dao *PostgresDAO) CountTeams() (int, error) {
	query := fmt.Sprintf("SELECT count(*) FROM %s", tableNames.Teams)

	var count int
	err := dao.DB.QueryRow(query).Scan(&count)
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

func (dao *PostgresDAO) GetLeagues() ([]apifootball.League, error) {
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY %s ASC", tableNames.Leagues, leagueColumns.Name)

	var leagues []apifootball.League
	rows, err := dao.DB.Query(query)
	if err != nil {
		return leagues, err
	}
	defer rows.Close()

	for rows.Next() {
		var league apifootball.League
		err := rows.Scan(&league.Id, &league.Name, &league.Type, &league.Logo, &league.CreatedAt)
		if err != nil {
			return leagues, err
		}
		leagues = append(leagues, league)
	}

	return leagues, nil
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

func (dao *PostgresDAO) GetTeams() ([]apifootball.Team, error) {
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY %s ASC", tableNames.Teams, teamColumns.Name)

	var teams []apifootball.Team
	rows, err := dao.DB.Query(query)
	if err != nil {
		return teams, err
	}
	defer rows.Close()

	for rows.Next() {
		var team apifootball.Team
		err := rows.Scan(&team.Id, &team.Name, &team.Code, &team.Country, &team.Founded, &team.National, &team.Logo, &team.CreatedAt)
		if err != nil {
			return teams, err
		}
		teams = append(teams, team)
	}

	return teams, nil
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

func (dao *PostgresDAO) InsertLeague(league *apifootball.League) (*apifootball.League, error) {
	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES ($1, $2, $3, $4) ON CONFLICT (%s) DO NOTHING RETURNING *",
		tableNames.Leagues,
		leagueColumns.Id, leagueColumns.Name, leagueColumns.Type, leagueColumns.Logo,
		leagueColumns.Id,
	)

	row := dao.DB.QueryRow(
		query, league.Id, league.Name, league.Type, league.Logo,
	)

	err := row.Scan(&league.Id, &league.Name, &league.Type, &league.Logo, &league.CreatedAt)
	if err != nil {
		return league, err
	}

	return league, nil
}

func (dao *PostgresDAO) InsertTeam(team *apifootball.Team) (*apifootball.Team, error) {
	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (%s) DO NOTHING RETURNING *",
		tableNames.Teams,
		teamColumns.Id, teamColumns.Name, teamColumns.Code, teamColumns.Country, teamColumns.Founded, teamColumns.National, teamColumns.Logo,
		leagueColumns.Id,
	)

	row := dao.DB.QueryRow(
		query, team.Id, team.Name, team.Code, team.Country, team.Founded, team.National, team.Logo,
	)

	err := row.Scan(&team.Id, &team.Name, &team.Code, &team.Country, &team.Founded, &team.National, &team.Logo, &team.CreatedAt)
	if err != nil {
		return team, err
	}

	return team, nil
}
