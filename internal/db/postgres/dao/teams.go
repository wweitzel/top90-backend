package dao

import (
	"github.com/lib/pq"
	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/db/postgres/dao/query"
)

func (dao *PostgresDAO) CountTeams() (int, error) {
	query := query.CountTeams()

	var count int
	err := dao.DB.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (dao *PostgresDAO) GetTeams(filter db.GetTeamsFilter) ([]apifootball.Team, error) {
	query, args := query.GetTeams(filter)

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
	query, args := query.GetTeamsForLeagueAndSeason(leagueId, season)

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

func (dao *PostgresDAO) InsertTeam(team *apifootball.Team) (*apifootball.Team, error) {
	query := query.InsertTeamQuery(team)

	row := dao.DB.QueryRow(
		query, team.Id, team.Name, team.Code, team.Country, team.Founded, team.National, team.Logo,
	)

	err := row.Scan(&team.Id, &team.Name, &team.Code, &team.Country, &team.Founded, &team.National, &team.Logo, &team.CreatedAt, pq.Array(&team.Aliases))
	if err != nil {
		return team, err
	}

	return team, nil
}
