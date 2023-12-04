package main

import (
	"database/sql"
	"log"
)

func (app *App) IngestFixtures() {
	// TODO: Take from command line input
	const SEASON = 2023

	leagues, err := app.dao.GetLeagues()

	if err != nil {
		log.Fatalln("Could not get leagues from database")
	}

	for _, league := range leagues {
		fixtures, err := app.client.GetFixtures(78, SEASON)
		if err != nil {
			log.Fatalf("Could not get teams for leagueId %d, season %d\n due to %v\n", league.Id, SEASON, err)
		}

		for _, fixture := range fixtures {
			createdFixture, err := app.dao.InsertFixture(&fixture)
			if err == sql.ErrNoRows {
				log.Printf("Already stored fixture for id %d", fixture.Id)
			} else if err != nil {
				log.Fatalf("Failed to insert fixture: %v", err)
			} else {
				log.Println("Successfully inserted fixture")
				log.Println("id:", createdFixture.Id)
				log.Println("referree:", createdFixture.Referee)
				log.Println("date:", createdFixture.Date)
				log.Println("home team:", createdFixture.Teams.Home.Name)
				log.Println("away team:", createdFixture.Teams.Away.Name)
				log.Println()
			}
		}
	}
}
