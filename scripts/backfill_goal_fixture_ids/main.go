package main

import (
	"log"
	"net/http"
	"os"
	"time"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/apifootball"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/poller"
)

func main() {
	log.SetFlags(log.Ltime)

	config := top90.LoadConfig()

	// Setup database
	DB, err := db.NewPostgresDB(config.DbUser, config.DbPassword, config.DbName, config.DbHost, config.DbPort)
	if err != nil {
		log.Fatalf("Could not set up database: %v", err)
	}
	defer DB.Close()

	dao := db.NewPostgresDAO(DB)

	host := os.Getenv("API_FOOTBALL_RAPID_API_HOST")
	apiKey := os.Getenv("API_FOOTBALL_RAPID_API_KEY")
	httpClient := &http.Client{Timeout: 10 * time.Second}
	apifootballClient := apifootball.NewClient(host, apiKey, httpClient)

	worldCupTeams, err := apifootballClient.GetTeams(10, 2022)
	if err != nil {
		log.Fatalln(err)
	}

	goals, err := dao.GetGoals(db.Pagination{Skip: 0, Limit: 100000}, db.GetGoalsFilter{})
	if err != nil {
		log.Fatalf("Failed %v", err)
	}

	notAWorldCupTeam := 0
	worldCupTeam := 0
	foundWorldCupFixture := 0
	couldNotDetermineTeamName := 0
	couldNotDetermineFixture := 0

	for _, goal := range goals {
		firstTeamNameFromPost, err := poller.GetTeamName(goal.RedditPostTitle)
		if err != nil {
			couldNotDetermineTeamName++
			continue
		}

		team, err := poller.GetTeamForTeamName(firstTeamNameFromPost, worldCupTeams)
		if err != nil {
			notAWorldCupTeam++
			continue
		}

		worldCupTeam++

		fixtures, err := dao.GetFixtures(db.GetFixuresFilter{Date: goal.RedditPostCreatedAt})
		if err != nil {
			log.Fatalln(err)
			continue
		}

		fixture, err := poller.GetFixtureForTeamName(firstTeamNameFromPost, team.Aliases, fixtures)
		if err != nil {
			couldNotDetermineFixture++
			continue
		}

		log.Println(fixture.LeagueId)
		log.Println("-------------------------------------------------------------------------------------")
		log.Println(goal.RedditPostTitle)

		foundWorldCupFixture++

		if goal.FixtureId == 0 {
			// _, err := dao.UpdateGoal(goal.Id, top90.Goal{FixtureId: fixture.Id})
			if err != nil {
				log.Fatalln(err)
			}
			log.Println("Success!")
		} else {
			log.Println("Already determined fixture")
		}

	}

	log.Println("=====================================================================================")
	log.Println("Total:", len(goals))
	log.Println("Total World cup team games:", worldCupTeam)
	log.Println("Total Not World cup team games:", notAWorldCupTeam)
	log.Println("Total Could not determine Fixtures:", couldNotDetermineFixture)
	log.Println("Total Found World Cup Fixtures:", foundWorldCupFixture)
}
