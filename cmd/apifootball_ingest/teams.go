package main

import (
	"database/sql"
	"log"
)

func (app *App) IngestTeams() {
	// TODO: Take from command line input
	const SEASON = 2022

	leagues, err := app.dao.GetLeagues()

	if err != nil {
		log.Fatalln("Could not get leagues from database")
	}

	for _, league := range leagues {
		teams, err := app.client.GetTeams(league.Id, SEASON)
		if err != nil {
			log.Fatalf("Could not get teams for leagueId %d, season %d due to %v\n", league.Id, SEASON, err)
		}

		for _, team := range teams {
			createdTeam, err := app.dao.InsertTeam(&team)
			if err == sql.ErrNoRows {
				log.Printf("Already stored team for id %d, name %s", team.Id, team.Name)
			} else if err != nil {
				log.Fatalf("Failed to insert team: %v", err)
			} else {
				log.Println("Successfully inserted team")
				log.Println("id:", createdTeam.Id)
				log.Println("country:", createdTeam.Country)
				log.Println("name:", createdTeam.Name)
				log.Println("founded:", createdTeam.Founded)
				log.Println("logo:", createdTeam.Logo)
				log.Println()
			}
		}
	}
}
