package main

import (
	"database/sql"
	"log"
)

// TODO: Take from command line input
var LEAGUES_TO_INGEST = []leagueCountryAndName{
	{Country: "England", Name: "Premier League"},
	{Country: "Italy", Name: "Serie A"},
	{Country: "Spain", Name: "La Liga"},
	{Country: "Germany", Name: "Bundesliga"},
	{Country: "France", Name: "Ligue 1"},
	{Country: "World", Name: "UEFA Champions League"},
	{Country: "World", Name: "UEFA Europa League"},
	{Country: "World", Name: "World Cup"},
}

type leagueCountryAndName struct {
	Country string
	Name    string
}

func (app *App) IngestLeagues() {
	for _, leagueToIngest := range LEAGUES_TO_INGEST {
		// Get the league from apifootball
		league, err := app.client.GetLeague(leagueToIngest.Country, leagueToIngest.Name)
		if err != nil {
			log.Fatalf("Could not get league for country %s, name %s \n%v", leagueToIngest.Country, leagueToIngest.Name, err)
		}

		// Insert league into the db
		createdLeague, err := app.dao.InsertLeague(&league)
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
