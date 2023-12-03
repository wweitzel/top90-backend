package dao

import (
	"database/sql"

	"github.com/wweitzel/top90/internal/db"
)

type PostgresDAO struct {
	DB *sql.DB
}

func NewPostgresDAO(db *sql.DB) db.Top90DAO {
	return &PostgresDAO{
		DB: db,
	}
}
