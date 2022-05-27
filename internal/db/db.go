package db

import (
	"database/sql"
	"fmt"
)

func NewPostgresDB(username, password, database, host, port string) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, database)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return conn, err
	}

	err = conn.Ping()
	if err != nil {
		return conn, err
	}

	return conn, nil
}
