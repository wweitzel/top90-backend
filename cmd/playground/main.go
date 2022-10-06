package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/wweitzel/top90/internal/apifootball"
	"github.com/wweitzel/top90/internal/dotenv"
)

func main() {
	dotenv.Load()
	host := os.Getenv("API_FOOTBALL_RAPID_API_HOST")
	apiKey := os.Getenv("API_FOOTBALL_RAPID_API_KEY")
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Instantiate client for querying apifootball api
	client := apifootball.NewClient(host, apiKey, httpClient)

	country := "England"
	leagueName := "Premier League"
	season := 2022

	// Get a specific league
	league, err := client.GetLeague(country, leagueName)
	if err != nil {
		log.Fatalf("Could not get league for country %s, name %s \n%v", country, leagueName, err)
	}

	// Get fixtures for that league and season
	fixtures, err := client.GetFixtures(league.ID, season)
	if err != nil {
		log.Fatalf("Could not get fixtures for leagueId %d, season %d \n%v", league.ID, season, err)
	}

	// TODO: Store these fixtures in our own database
	for _, fixture := range fixtures {
		fmt.Println("------------------------------")
		fmt.Println("id: ", fixture.ID)
		fmt.Println("date: ", fixture.Date)
		fmt.Println("timestamp: ", fixture.Timestamp)
		fmt.Println("home team: ", fixture.Teams.Home.Name)
		fmt.Println("away team: ", fixture.Teams.Away.Name)
		fmt.Println("status: ", fixture.Status.Long)
	}
	fmt.Println()
}
