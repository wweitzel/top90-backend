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

	allTeams, err := dao.GetTeams(db.GetTeamsFilter{})
	allGoals, err := dao.GetGoals(db.Pagination{Skip: 0, Limit: 100000}, db.GetGoalsFilter{})

	if err != nil {
		log.Fatalf("Failed %v", err)
	}

	noTeamFound := 0
	foundFixture := 0
	couldNotDetermineTeamName := 0
	couldNotDetermineFixture := 0

	for _, goal := range allGoals {
		firstTeamNameFromPost, err := poller.GetTeamName(goal.RedditPostTitle)
		if err != nil {
			couldNotDetermineTeamName++
			continue
		}

		team, err := poller.GetTeamForTeamName(firstTeamNameFromPost, allTeams)
		if err != nil {
			noTeamFound++
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

		foundFixture++

		log.Println(fixture.LeagueId)
		log.Println("-------------------------------------------------------------------------------------")
		log.Println(goal.RedditPostTitle)

		if goal.FixtureId == 0 {
			_, err := dao.UpdateGoal(goal.Id, top90.Goal{FixtureId: fixture.Id})
			if err != nil {
				log.Fatalln(err)
			} else {
				log.Println("Success!")
			}
		} else {
			log.Println("Already determined fixture")
		}

	}

	log.Println("=====================================================================================")
	log.Println("Processed: ", len(allGoals))
	log.Println("No Team Found:", noTeamFound)
	log.Println("Total Could not determine Fixtures:", couldNotDetermineFixture)
	log.Println("Total Found World Cup Fixtures:", foundFixture)
}
