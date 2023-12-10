package dao

import (
	"github.com/wweitzel/top90/internal/db/dao/query"
	db "github.com/wweitzel/top90/internal/db/models"
)

func (dao *PostgresDAO) UpsertPlayer(player db.Player) (db.Player, error) {
	query, args := query.UpsertPlayer(player)
	var insertedPlayer db.Player
	err := dao.DB.QueryRowx(query, args...).StructScan(&insertedPlayer)
	return insertedPlayer, err
}
