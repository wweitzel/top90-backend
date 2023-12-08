package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

func NewPostgresDB(username, password, database, host, port string) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, database)

	db, err := sqlx.Connect("postgres", dsn)
	return db, err
}
