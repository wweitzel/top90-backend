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
	{Country: "USA", Name: "Major League Soccer"},
	// TODO: Leaving these uncommented causes too many requests error with the seed.
	// Need to update the seed to not have to hit apifootball api
	// {Country: "World", Name: "World Cup - Qualification Intercontinental Play-offs"},
	// {Country: "World", Name: "World Cup - Qualification CONCACAF"},
	// {Country: "World", Name: "World Cup - Qualification Europe"},
	// {Country: "World", Name: "World Cup - Qualification Oceania"},
	// {Country: "World", Name: "World Cup - Qualification South America"},
	// {Country: "World", Name: "World Cup - Qualification Africa"},
	// {Country: "World", Name: "World Cup - Qualification Asia"},
	// {Country: "World", Name: "Friendlies"},
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
			log.Fatalf("Could not get league for country %s, name %s due to %v\n", leagueToIngest.Country, leagueToIngest.Name, err)
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
