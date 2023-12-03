package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/db/postgres/dao"
)

type App struct {
	client *apifootball.Client
	dao    db.Top90DAO
}

func loadApp() (app App, conn *sql.DB) {
	// Load config from .env into environment variables
	config := config.Load()

	// Connect to database
	DB, err := db.NewPostgresDB(config.DbUser, config.DbPassword, config.DbName, config.DbHost, config.DbPort)
	if err != nil {
		log.Fatalf("Could not setup database: %v", err)
	}

	// Create dao for accessing the db
	dao := dao.NewPostgresDAO(DB)

	host := os.Getenv("API_FOOTBALL_RAPID_API_HOST")
	apiKey := os.Getenv("API_FOOTBALL_RAPID_API_KEY")
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Instantiate an apifootball api client
	client := apifootball.NewClient(host, apiKey, httpClient)

	return App{
		client: client,
		dao:    dao,
	}, DB
}
