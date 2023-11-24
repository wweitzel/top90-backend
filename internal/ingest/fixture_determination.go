package ingest

import (
	"errors"
	"strings"

	"github.com/wweitzel/top90/internal/apifootball"
)

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
