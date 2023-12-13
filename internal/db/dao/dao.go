package dao

import (
	"github.com/jmoiron/sqlx"
	db "github.com/wweitzel/top90/internal/db/models"
)

type Top90DAO interface {
	CountGoals(db.GetGoalsFilter) (int, error)
	CountTeams() (int, error)
	GetFixture(id int) (db.Fixture, error)
	GetFixtures(filter db.GetFixturesFilter) ([]db.Fixture, error)
	GetGoals(pagination db.Pagination, filter db.GetGoalsFilter) ([]db.Goal, error)
	GetGoal(id string) (db.Goal, error)
	GetLeagues() ([]db.League, error)
	GetNewestGoal() (db.Goal, error)
	GetPlayer(id int) (db.Player, error)
	GetTeams(filter db.GetTeamsFilter) ([]db.Team, error)
	GetTeamsForLeagueAndSeason(leagueId, season int) ([]db.Team, error)
	GetTopScorers() ([]db.Player, error)
	GoalExists(redditFullname string) (bool, error)
	InsertFixture(*db.Fixture) (*db.Fixture, error)
	InsertGoal(*db.Goal) (*db.Goal, error)
	InsertLeague(*db.League) (*db.League, error)
	InsertTeam(*db.Team) (*db.Team, error)
	PlayerExists(id int) (bool, error)
	SearchPlayers(searchTerm string) ([]db.Player, error)
	UpsertPlayer(db.Player) (db.Player, error)
	UpdateGoal(id string, goalUpdate db.Goal) (updatedGoal db.Goal, err error)
	UpdateLeague(id int, leagueUpdate db.League) (updatedLeague db.League, err error)
}

type PostgresDAO struct {
	DB *sqlx.DB
}

func NewPostgresDAO(db *sqlx.DB) Top90DAO {
	return &PostgresDAO{
		DB: db,
	}
}
