package scrape

import (
	"errors"
	"strings"
	"time"

	"github.com/wweitzel/top90/internal/clients/reddit"
	db "github.com/wweitzel/top90/internal/db/models"
)

func (s *Scraper) findFixture(p reddit.Post) (*db.Fixture, error) {
	createdAt := time.Unix(int64(p.Data.Created_utc), 0).UTC()
	dbFixtures, err := s.dao.GetFixtures(db.GetFixturesFilter{Date: createdAt})
	if err != nil {
		return nil, err
	}

	teams, _ := s.findTeams(p)

	for _, f := range dbFixtures {
		for _, t := range teams {
			if f.Teams.Home.Id == t.Id || f.Teams.Away.Id == t.Id {
				return &f, nil
			}
		}
	}

	return nil, nil
}

func (s *Scraper) findTeams(p reddit.Post) ([]db.Team, error) {
	dbTeams, err := s.dao.GetTeams(db.GetTeamsFilter{})
	if err != nil {
		return nil, errors.New("could not get teams from db")
	}

	var teamsFound []db.Team
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
