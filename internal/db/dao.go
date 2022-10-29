package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/apifootball"
)

type Top90DAO interface {
	CountGoals(GetGoalsFilter) (int, error)
	CountTeams() (int, error)
	GetFixtures(filter GetFixuresFilter) ([]apifootball.Fixture, error)
	GetGoals(pagination Pagination, filter GetGoalsFilter) ([]top90.Goal, error)
	GetGoal(id string) (top90.Goal, error)
	GetLeagues() ([]apifootball.League, error)
	GetNewestGoal() (top90.Goal, error)
	GetTeams(filter GetTeamsFilter) ([]apifootball.Team, error)
	GetTeamsForLeagueAndSeason(leagueId, season int) ([]apifootball.Team, error)
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
}

type GetFixuresFilter struct {
	LeagueId int
	Date     time.Time
}

type GetTeamsFilter struct {
	Country string
}

func NewPostgresDAO(db *sql.DB) Top90DAO {
	return &PostgresDAO{
		DB: db,
	}
}

func (dao *PostgresDAO) CountGoals(filter GetGoalsFilter) (int, error) {
	filter.SearchTerm = "%" + filter.SearchTerm + "%"

	query := fmt.Sprintf("SELECT count(*) FROM %s", tableNames.Goals)

	var variableCount int
	var args []any
	query, args = addGetGoalsJoinAndWhere(query, args, filter, &variableCount)

	var count int
	err := dao.DB.QueryRow(query, args...).Scan(&count)
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

func (dao *PostgresDAO) GetFixtures(filter GetFixuresFilter) ([]apifootball.Fixture, error) {
	var args []any

	whereClause, args := getFixturesWhereClause(filter, args)

	query := fmt.Sprintf(
		`SELECT
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			%s,
			home_teams.name as home_team_name,
			home_teams.logo as home_team_logo,
			away_teams.name as away_team_name,
			away_teams.logo as away_team_logo
			FROM %s
		JOIN %s home_teams ON home_teams.%s=%s
		JOIN %s away_teams ON away_teams.%s=%s
		WHERE %s ORDER BY %s ASC`,
		tableNames.Fixtures+"."+fixtureColumns.Id,
		tableNames.Fixtures+"."+fixtureColumns.Referee,
		tableNames.Fixtures+"."+fixtureColumns.Date,
		tableNames.Fixtures+"."+fixtureColumns.HomeTeamId,
		tableNames.Fixtures+"."+fixtureColumns.AwayTeamId,
		tableNames.Fixtures+"."+fixtureColumns.LeagueId,
		tableNames.Fixtures+"."+fixtureColumns.Season,
		tableNames.Fixtures+"."+fixtureColumns.CreatedAt,
		tableNames.Fixtures,
		tableNames.Teams,
		teamColumns.Id,
		tableNames.Fixtures+"."+fixtureColumns.HomeTeamId,
		tableNames.Teams,
		teamColumns.Id,
		tableNames.Fixtures+"."+fixtureColumns.AwayTeamId,
		whereClause,
		fixtureColumns.Date)

	var fixtures []apifootball.Fixture
	rows, err := dao.DB.Query(query, args...)
	if err != nil {
		return fixtures, err
	}
	defer rows.Close()

	for rows.Next() {
		var fixture apifootball.Fixture
		err := rows.Scan(
			&fixture.Id,
			&fixture.Referee,
			&fixture.Date,
			&fixture.Teams.Home.Id,
			&fixture.Teams.Away.Id,
			&fixture.LeagueId,
			&fixture.Season,
			&fixture.CreatedAt,
			&fixture.Teams.Home.Name,
			&fixture.Teams.Home.Logo,
			&fixture.Teams.Away.Name,
			&fixture.Teams.Away.Logo)
		if err != nil {
			return fixtures, err
		}
		fixture.Timestamp = fixture.Date.Unix()
		fixtures = append(fixtures, fixture)
	}

	return fixtures, nil
}

func (dao *PostgresDAO) GetGoals(pagination Pagination, filter GetGoalsFilter) ([]top90.Goal, error) {
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
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = $1", tableNames.Goals, goalColumns.Id)

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
	newestDbGoals, err := dao.GetGoals(pagination, GetGoalsFilter{})
	if err != nil {
		return top90.Goal{}, err
	}

	var newestDbGoal top90.Goal
	if len(newestDbGoals) > 0 {
		newestDbGoal = newestDbGoals[0]
	}

	return newestDbGoal, nil
}

func (dao *PostgresDAO) GetTeams(filter GetTeamsFilter) ([]apifootball.Team, error) {
	var args []any

	whereClause, args := getTeamsWhereClause(filter, args)

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s ORDER BY %s ASC", tableNames.Teams, whereClause, teamColumns.Name)

	var teams []apifootball.Team
	rows, err := dao.DB.Query(query, args...)
	if err != nil {
		return teams, err
	}
	defer rows.Close()

	for rows.Next() {
		var team apifootball.Team
		err := rows.Scan(&team.Id, &team.Name, &team.Code, &team.Country, &team.Founded, &team.National, &team.Logo, &team.CreatedAt, pq.Array(&team.Aliases))
		if err != nil {
			return teams, err
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (dao *PostgresDAO) GetTeamsForLeagueAndSeason(leagueId, season int) ([]apifootball.Team, error) {
	var args []any

	// Union of home teams ids and away team ids for a given league season
	query := fmt.Sprintf("SELECT * from %s WHERE %s in (", tableNames.Teams, teamColumns.Id)

	// Teams for league id and season
	if leagueId != 0 && season != 0 {
		query = query + fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1 AND %s = $2", fixtureColumns.HomeTeamId, tableNames.Fixtures, fixtureColumns.LeagueId, fixtureColumns.Season)
		query = query + " UNION "
		query = query + fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1 AND %s = $2", fixtureColumns.AwayTeamId, tableNames.Fixtures, fixtureColumns.LeagueId, fixtureColumns.Season)
		args = append(args, leagueId, season)
	}

	// Teams for league id
	if leagueId != 0 && season == 0 {
		query = query + fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", fixtureColumns.HomeTeamId, tableNames.Fixtures, fixtureColumns.LeagueId)
		query = query + " UNION "
		query = query + fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", fixtureColumns.AwayTeamId, tableNames.Fixtures, fixtureColumns.LeagueId)
		args = append(args, leagueId)
	}

	// Teams for season
	if leagueId == 0 && season != 0 {
		query = query + fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", fixtureColumns.HomeTeamId, tableNames.Fixtures, fixtureColumns.Season)
		query = query + " UNION "
		query = query + fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", fixtureColumns.AwayTeamId, tableNames.Fixtures, fixtureColumns.Season)
		args = append(args, season)
	}

	// All teams
	if leagueId == 0 && season == 0 {
		query = query + fmt.Sprintf("SELECT %s FROM %s", fixtureColumns.HomeTeamId, tableNames.Fixtures)
		query = query + " UNION "
		query = query + fmt.Sprintf("SELECT %s FROM %s", fixtureColumns.AwayTeamId, tableNames.Fixtures)
	}

	query = query + (")")

	var teams []apifootball.Team
	rows, err := dao.DB.Query(query, args...)
	if err != nil {
		return teams, err
	}
	defer rows.Close()

	for rows.Next() {
		var team apifootball.Team
		err := rows.Scan(&team.Id, &team.Name, &team.Code, &team.Country, &team.Founded, &team.National, &team.Logo, &team.CreatedAt, pq.Array(&team.Aliases))
		if err != nil {
			return teams, err
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (dao *PostgresDAO) InsertFixture(fixture *apifootball.Fixture) (*apifootball.Fixture, error) {
	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (%s) DO NOTHING RETURNING *",
		tableNames.Fixtures,
		fixtureColumns.Id, fixtureColumns.Referee, fixtureColumns.Date, fixtureColumns.HomeTeamId, fixtureColumns.AwayTeamId, fixtureColumns.LeagueId, fixtureColumns.Season,
		fixtureColumns.Id,
	)

	row := dao.DB.QueryRow(
		query, fixture.Id, fixture.Referee, time.Unix(fixture.Timestamp, 0), fixture.Teams.Home.Id, fixture.Teams.Away.Id, fixture.LeagueId, fixture.Season,
	)

	err := row.Scan(&fixture.Id, &fixture.Referee, &fixture.Date, &fixture.Teams.Home.Id, &fixture.Teams.Away.Id, &fixture.LeagueId, &fixture.Season, &fixture.CreatedAt)
	if err != nil {
		return fixture, err
	}

	return fixture, nil
}

func (dao *PostgresDAO) InsertGoal(goal *top90.Goal) (*top90.Goal, error) {
	id := uuid.NewString()
	id = strings.Replace(id, "-", "", -1)

	query := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (%s) DO NOTHING RETURNING *",
		tableNames.Goals,
		goalColumns.Id, goalColumns.RedditFullname, goalColumns.RedditLinkUrl, goalColumns.RedditPostTitle, goalColumns.RedditPostCreatedAt, goalColumns.S3ObjectKey, goalColumns.FixtureId, goalColumns.ThumbnailS3Key,
		goalColumns.RedditFullname,
	)

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
	)

	err := row.Scan(&goal.Id, &goal.RedditFullname, &goal.RedditLinkUrl, &goal.RedditPostTitle, &goal.RedditPostCreatedAt, &goal.S3ObjectKey, &goal.CreatedAt, &fixtureId, &thumbnailS3Key)
	if err != nil {
		return goal, err
	}

	goal.FixtureId = int(fixtureId.Int64)

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

	err := row.Scan(&team.Id, &team.Name, &team.Code, &team.Country, &team.Founded, &team.National, &team.Logo, &team.CreatedAt, pq.Array(&team.Aliases))
	if err != nil {
		return team, err
	}

	return team, nil
}

// UpdateGoal updates the goal with primary key = id.
// The function will update any fields that are set on goalUpdate.
// This means you should only set fields on goalUpdate that you actually
// want to be updated.
func (dao *PostgresDAO) UpdateGoal(id string, goalUpdate top90.Goal) (top90.Goal, error) {
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
