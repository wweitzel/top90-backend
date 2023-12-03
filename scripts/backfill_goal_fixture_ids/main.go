package main

import (
	"errors"
	"log"
	"strings"

	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/config"
	"github.com/wweitzel/top90/internal/db"
)

func main() {
	log.SetFlags(log.Ltime)

	config := config.Load()

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

		fixture, err := FindFixture(goal.RedditPostTitle, allTeams, fixtures)
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

func findTeams(redditPostTile string, teams []apifootball.Team) []apifootball.Team {
	var teamsFound []apifootball.Team
	for _, team := range teams {
		if strings.Contains(redditPostTile, team.Name) {
			teamsFound = append(teamsFound, team)
		} else {
			for _, alias := range team.Aliases {
				if strings.Contains(redditPostTile, alias) {
					teamsFound = append(teamsFound, team)
				}
			}
		}
	}
	return teamsFound
}

func FindFixture(redditPostTitle string, allTeams []apifootball.Team, fixtures []apifootball.Fixture) (apifootball.Fixture, error) {
	teams := findTeams(redditPostTitle, allTeams)

	for _, f := range fixtures {
		for _, t := range teams {
			if f.Teams.Home.Id == t.Id || f.Teams.Away.Id == t.Id {
				return f, nil
			}
		}
	}

	return apifootball.Fixture{}, errors.New("could not determine fixture")
}
