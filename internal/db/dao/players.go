package dao

import (
	"github.com/wweitzel/top90/internal/db/dao/query"
	db "github.com/wweitzel/top90/internal/db/models"
)

func (dao *PostgresDAO) GetPlayer(id int) (db.Player, error) {
	query, args := query.GetPlayer(id)
	var player db.Player
	err := dao.DB.Get(&player, query, args...)
	return player, err
}

func (dao *PostgresDAO) PlayerExists(id int) (bool, error) {
	query, args := query.PlayerExists(id)
	var count int
	err := dao.DB.Get(&count, query, args...)
	return count > 0, err
}

func (dao *PostgresDAO) UpsertPlayer(player db.Player) (db.Player, error) {
	query, args := query.UpsertPlayer(player)
	var insertedPlayer db.Player
	err := dao.DB.QueryRowx(query, args...).StructScan(&insertedPlayer)
	return insertedPlayer, err
}

func (dao *PostgresDAO) SearchPlayers(searchTerm string) ([]db.Player, error) {
	query, args := query.SearchPlayers(searchTerm)
	var players []db.Player
	err := dao.DB.Select(&players, query, args...)
	return players, err
}

func (dao *PostgresDAO) GetTopScorers() ([]db.Player, error) {
	query := query.GetTopScorers()
	var players []db.Player
	err := dao.DB.Unsafe().Select(&players, query)
	return players, err
}
