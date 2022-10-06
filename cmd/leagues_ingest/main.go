package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/apifootball"
	"github.com/wweitzel/top90/internal/db"
)

var LEAGUES_TO_INGEST = []LeagueCountryAndName{
	{Country: "England", Name: "Premier League"},
	{Country: "Italy", Name: "Serie A"},
	{Country: "Spain", Name: "La Liga"},
	{Country: "Germany", Name: "Bundesliga"},
	{Country: "France", Name: "Ligue 1"},
	{Country: "World", Name: "UEFA Champions League"},
	{Country: "World", Name: "UEFA Europa League"},
	{Country: "World", Name: "World Cup"},
}

type LeagueCountryAndName struct {
	Country string
	Name    string
}

func main() {
	log.SetFlags(log.Ltime)

	// Load config from .env into environment variables
	config := top90.LoadConfig()

	// Connect to database
	DB, err := db.NewPostgresDB(config.DbUser, config.DbPassword, config.DbName, config.DbHost, config.DbPort)
	if err != nil {
		log.Fatalf("Could not setup database: %v", err)
	}
	defer DB.Close()

	// Create dao for accessing the db
	dao := db.NewPostgresDAO(DB)

	host := os.Getenv("API_FOOTBALL_RAPID_API_HOST")
	apiKey := os.Getenv("API_FOOTBALL_RAPID_API_KEY")
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Instantiate an apifootball api client
	client := apifootball.NewClient(host, apiKey, httpClient)

	for _, leagueToIngest := range LEAGUES_TO_INGEST {
		// Get the league from apifootball
		league, err := client.GetLeague(leagueToIngest.Country, leagueToIngest.Name)
		if err != nil {
			log.Fatalf("Could not get league for country %s, name %s \n%v", leagueToIngest.Country, leagueToIngest.Name, err)
		}

		// Insert league into the db
		createdLeague, err := dao.InsertLeague(&league)
		if err == sql.ErrNoRows {
			log.Printf("Already stored league for id %d", league.Id)
		} else if err != nil {
			log.Fatalf("Failed to insert league: %v", err)
		} else {
			log.Println("Successfully inserted league")
			log.Println("id:", createdLeague.Id)
			log.Println("name:", createdLeague.Name)
			log.Println("type:", createdLeague.Type)
			log.Println("logo:", createdLeague.Logo)
			log.Println()
		}
	}
}
