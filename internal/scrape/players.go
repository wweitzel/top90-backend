package scrape

import (
	"fmt"

	"github.com/wweitzel/top90/internal/clients/apifootball"
	db "github.com/wweitzel/top90/internal/db/models"
)

func (s *Scraper) linkPlayerWithApiFootball(redditPostTitle string, fixtureId int) (db.Player, apifootball.Event, error) {
	event, err := s.getEventFromApiFootball(redditPostTitle, fixtureId)
	if err != nil {
		return db.Player{},
			apifootball.Event{},
			err
	}

	playerExists, err := s.dao.PlayerExists(event.Player.ID)
	if err != nil {
		return db.Player{},
			apifootball.Event{},
			fmt.Errorf("failed checking if player exists in db: %v", err)
	}

	if playerExists {
		player, err := s.dao.GetPlayer(event.Player.ID)
		if err != nil {
			return db.Player{},
				apifootball.Event{},
				fmt.Errorf("failed getting existing player from db")
		}
		return player, event, nil
	}

	player, err := s.apifbClient.GetPlayer(event.Player.ID, s.apifbClient.CurrentSeason)
	if err != nil {
		return db.Player{},
			apifootball.Event{},
			fmt.Errorf("failed getting player from apifootball")
	}
	insertedPlayer, err := s.dao.UpsertPlayer(player)
	if err != nil {
		return db.Player{},
			apifootball.Event{},
			fmt.Errorf("failed upserting player into db")
	}
	return insertedPlayer, event, nil
}
