package dao

import (
	"time"

	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/db/postgres/dao/query"
)

func (dao *PostgresDAO) GetFixtures(filter db.GetFixuresFilter) ([]apifootball.Fixture, error) {
	query, args := query.GetFixtures(filter)

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

func (dao *PostgresDAO) GetFixture(id int) (apifootball.Fixture, error) {
	query := query.GetFixture(id)

	var fixture apifootball.Fixture
	row := dao.DB.QueryRow(query, id)

	err := row.Scan(
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
		return fixture, err
	}

	return fixture, nil
}

func (dao *PostgresDAO) InsertFixture(fixture *apifootball.Fixture) (*apifootball.Fixture, error) {
	query := query.InsertFixture(fixture)

	row := dao.DB.QueryRow(
		query, fixture.Id, fixture.Referee, time.Unix(fixture.Timestamp, 0), fixture.Teams.Home.Id, fixture.Teams.Away.Id, fixture.LeagueId, fixture.Season, fixture.Date,
	)

	err := row.Scan(&fixture.Id, &fixture.Referee, &fixture.Date, &fixture.Teams.Home.Id, &fixture.Teams.Away.Id, &fixture.LeagueId, &fixture.Season, &fixture.CreatedAt)
	if err != nil {
		return fixture, err
	}

	return fixture, nil
}
