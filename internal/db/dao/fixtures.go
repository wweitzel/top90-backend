package dao

import (
	"github.com/wweitzel/top90/internal/db/dao/query"
	db "github.com/wweitzel/top90/internal/db/models"
)

func (dao *PostgresDAO) GetFixtures(filter db.GetFixturesFilter) ([]db.Fixture, error) {
	query, args := query.GetFixtures(filter)

	var fixtures []db.Fixture
	rows, err := dao.DB.Query(query, args...)
	if err != nil {
		return fixtures, err
	}
	defer rows.Close()

	for rows.Next() {
		var fixture db.Fixture
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

func (dao *PostgresDAO) GetFixture(id int) (db.Fixture, error) {
	query := query.GetFixture(id)

	var fixture db.Fixture
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
	return fixture, err
}

func (dao *PostgresDAO) InsertFixture(fixture *db.Fixture) (*db.Fixture, error) {
	query, args := query.InsertFixture(fixture)

	var insertedFixture db.Fixture
	row := dao.DB.QueryRow(query, args...)
	err := row.Scan(
		&insertedFixture.Id,
		&insertedFixture.Referee,
		&insertedFixture.Date,
		&insertedFixture.Teams.Home.Id,
		&insertedFixture.Teams.Away.Id,
		&insertedFixture.LeagueId,
		&insertedFixture.Season,
		&insertedFixture.CreatedAt)
	return &insertedFixture, err
}
