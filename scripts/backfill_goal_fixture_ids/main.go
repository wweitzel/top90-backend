package main

import (
	"log"

	top90 "github.com/wweitzel/top90/internal"
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

	premierLeagueTeams, err := dao.GetTeams(db.GetTeamsFilter{Country: "England"})
	if err != nil {
		log.Fatalln(err)
	}

	goals, err := dao.GetGoals(db.Pagination{Skip: 0, Limit: 100000}, db.GetGoalsFilter{})
	if err != nil {
		log.Fatalf("Failed %v", err)
	}

	notAPremierLeagueTeam := 0
	couldNotDetermineTeamName := 0
	couldNotDetermineFixture := 0
	actualPremierLeagueMatch := 0

	for _, goal := range goals {
		firstTeamNameFromPost, err := poller.GetTeamName(goal.RedditPostTitle)
		if err != nil {
			couldNotDetermineTeamName++
			continue
		}

		team, err := poller.GetTeamForTeamName(firstTeamNameFromPost, premierLeagueTeams)
		if err != nil {
			notAPremierLeagueTeam++
			continue
		}

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

		if fixture.LeagueId == 39 {
			log.Println("-------------------------------------------------------------------------------------")
			log.Println(goal.RedditPostTitle)

			actualPremierLeagueMatch++

			_, err := dao.UpdateGoal(goal.Id, top90.Goal{FixtureId: fixture.Id})
			if err != nil {
				log.Fatalln(err)
			}

			log.Println("Success!")
		}
	}

	log.Println("=====================================================================================")
	log.Println("Total:", len(goals))
	log.Println("Total Premier League:", actualPremierLeagueMatch)
}
