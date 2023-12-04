package scrape

import (
	"errors"
	"strings"
	"time"

	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/clients/reddit"
	"github.com/wweitzel/top90/internal/db"
)

func (s *Scraper) findFixture(p reddit.Post) (apifootball.Fixture, error) {
	createdAt := time.Unix(int64(p.Data.Created_utc), 0).UTC()
	dbFixtures, _ := s.dao.GetFixtures(db.GetFixuresFilter{Date: createdAt})
	teams, _ := s.findTeams(p)

	for _, f := range dbFixtures {
		for _, t := range teams {
			if f.Teams.Home.Id == t.Id || f.Teams.Away.Id == t.Id {
				return f, nil
			}
		}
	}

	return apifootball.Fixture{}, errors.New("could not determine fixture")
}

func (s *Scraper) findTeams(p reddit.Post) ([]apifootball.Team, error) {
	dbTeams, err := s.dao.GetTeams(db.GetTeamsFilter{})
	if err != nil {
		return nil, errors.New("could not get teams from db")
	}

	var teamsFound []apifootball.Team
	for _, team := range dbTeams {
		if strings.Contains(p.Data.Title, team.Name) {
			teamsFound = append(teamsFound, team)
		} else {
			for _, alias := range team.Aliases {
				if strings.Contains(p.Data.Title, alias) {
					teamsFound = append(teamsFound, team)
				}
			}
		}
	}
	return teamsFound, nil
}
