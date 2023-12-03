package db

import (
	"database/sql"
	"time"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/clients/apifootball"
)

type Top90DAO interface {
	CountGoals(GetGoalsFilter) (int, error)
	CountTeams() (int, error)
	GetFixture(id int) (apifootball.Fixture, error)
	GetFixtures(filter GetFixuresFilter) ([]apifootball.Fixture, error)
	GetGoals(pagination Pagination, filter GetGoalsFilter) ([]top90.Goal, error)
	GetGoal(id string) (top90.Goal, error)
	GetLeagues() ([]apifootball.League, error)
	GetNewestGoal() (top90.Goal, error)
	GetTeams(filter GetTeamsFilter) ([]apifootball.Team, error)
	GetTeamsForLeagueAndSeason(leagueId, season int) ([]apifootball.Team, error)
	GoalExists(redditFullname string) (bool, error)
	InsertFixture(*apifootball.Fixture) (*apifootball.Fixture, error)
	InsertGoal(*top90.Goal) (*top90.Goal, error)
	InsertLeague(*apifootball.League) (*apifootball.League, error)
	InsertTeam(*apifootball.Team) (*apifootball.Team, error)
	UpdateGoal(id string, goalUpdate top90.Goal) (updatedGoal top90.Goal, err error)
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
	StartDate  string
	LeagueId   int
	Season     int
	TeamId     int
	FixtureId  int
}

type GetFixuresFilter struct {
	LeagueId int
	Date     time.Time
}

type GetTeamsFilter struct {
	Country    string
	SearchTerm string
}

func NewPostgresDAO(db *sql.DB) Top90DAO {
	return &PostgresDAO{
		DB: db,
	}
}
