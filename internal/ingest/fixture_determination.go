package ingest

import (
	"errors"
	"strings"

	"github.com/wweitzel/top90/internal/apifootball"
)

func GetTeamName(redditPostTitle string) (string, error) {
	// TODO: There has to be a better way than this
	index0 := strings.IndexByte(redditPostTitle, '0')
	index1 := strings.IndexByte(redditPostTitle, '1')
	index2 := strings.IndexByte(redditPostTitle, '2')
	index3 := strings.IndexByte(redditPostTitle, '3')
	index4 := strings.IndexByte(redditPostTitle, '4')
	index5 := strings.IndexByte(redditPostTitle, '5')
	index6 := strings.IndexByte(redditPostTitle, '6')
	index7 := strings.IndexByte(redditPostTitle, '7')
	index8 := strings.IndexByte(redditPostTitle, '8')
	index9 := strings.IndexByte(redditPostTitle, '9')
	indexBracket := strings.IndexByte(redditPostTitle, '[')

	lowestIndex := 9999999
	if index0 != -1 && index0 < lowestIndex {
		lowestIndex = index0
	}
	if index1 != -1 && index1 < lowestIndex {
		lowestIndex = index1
	}
	if index2 != -1 && index2 < lowestIndex {
		lowestIndex = index2
	}
	if index3 != -1 && index3 < lowestIndex {
		lowestIndex = index3
	}
	if index4 != -1 && index4 < lowestIndex {
		lowestIndex = index4
	}
	if index5 != -1 && index5 < lowestIndex {
		lowestIndex = index5
	}
	if index6 != -1 && index6 < lowestIndex {
		lowestIndex = index6
	}
	if index7 != -1 && index7 < lowestIndex {
		lowestIndex = index7
	}
	if index8 != -1 && index8 < lowestIndex {
		lowestIndex = index8
	}
	if index9 != -1 && index9 < lowestIndex {
		lowestIndex = index9
	}
	if indexBracket != -1 && indexBracket < lowestIndex {
		lowestIndex = indexBracket
	}

	if lowestIndex == 9999999 {
		return "", errors.New("could not determine team name")
	}

	teamName := strings.TrimSpace(redditPostTitle[:lowestIndex])
	return teamName, nil
}

func GetTeamForTeamName(teamName string, teams []apifootball.Team) (apifootball.Team, error) {
	for _, t := range teams {
		if t.Name == teamName {
			return t, nil
		}

		for _, alias := range t.Aliases {
			if alias == teamName {
				return t, nil
			}
		}
	}

	return apifootball.Team{}, errors.New("could not determine team")
}

func GetFixtureForTeamName(teamName string, aliases []string, fixtures []apifootball.Fixture) (apifootball.Fixture, error) {
	var fixture apifootball.Fixture
	var foundFixture bool

	for _, f := range fixtures {
		if f.Teams.Home.Name == teamName || f.Teams.Away.Name == teamName {
			fixture = f
			foundFixture = true
			break
		}

		for _, alias := range aliases {
			if f.Teams.Home.Name == alias || f.Teams.Away.Name == alias {
				fixture = f
				foundFixture = true
				break
			}
		}
	}

	if !foundFixture {
		return apifootball.Fixture{}, errors.New("could not determine fixture")
	}

	return fixture, nil
}
