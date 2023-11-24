package main

import (
	"log"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/ingest"
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
	terminateIfError(err)

	allGoals, err := dao.GetGoals(db.Pagination{Skip: 0, Limit: 100000}, db.GetGoalsFilter{})
	terminateIfError(err)

	successCount := 0

	for _, goal := range allGoals {
		if goal.FixtureId != 0 {
			log.Println("Already determined fixture")
			continue
		}

		fixtures, err := dao.GetFixtures(db.GetFixuresFilter{Date: goal.RedditPostCreatedAt})
		terminateIfError(err)

		fixture, err := ingest.FindFixture(goal.RedditPostTitle, allTeams, fixtures)
		if err != nil {
			log.Printf("%v", err)
			continue
		}

		_, err = dao.UpdateGoal(goal.Id, top90.Goal{FixtureId: fixture.Id})
		terminateIfError(err)

		log.Println("-------------------------------------------------------------------------------------")
		log.Println(goal.RedditPostTitle)
		log.Println("Success!")
		successCount++
	}

	log.Println("=====================================================================================")
	log.Println("Processed: ", len(allGoals))
	log.Println("Successes: ", successCount)
}

func terminateIfError(err error) {
	if err != nil {
		log.Fatalf("Failed %v", err)
	}
}
