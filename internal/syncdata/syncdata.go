package syncdata

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/db/dao"
	db "github.com/wweitzel/top90/internal/db/models"
)

type SyncData struct {
	dao    dao.Top90DAO
	source apifootball.Client
	logger *slog.Logger
}

func New(dao dao.Top90DAO, client apifootball.Client, logger *slog.Logger) SyncData {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return SyncData{
		dao:    dao,
		source: client,
		logger: logger,
	}
}

func (s *SyncData) Leagues() error {
	dbLeagues, err := s.dao.GetLeagues()
	if err != nil {
		return fmt.Errorf("error getting leagues from db %v", err)
	}

	for _, dbLeague := range dbLeagues {
		sourceLeague, err := s.source.GetLeague(dbLeague.Id)
		if err != nil {
			return fmt.Errorf("error getting leagues from apifootball %v", err)
		}
		_, err = s.dao.UpdateLeague(dbLeague.Id, db.League{
			CurrentSeason: sourceLeague.CurrentSeason,
		})
		if err != nil {
			return fmt.Errorf("error updating league in db %v", err)
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}

func (s *SyncData) Teams() error {
	leagues, err := s.dao.GetLeagues()
	if err != nil {
		return fmt.Errorf("error getting leagues from db %v", err)
	}

	for _, league := range leagues {
		sourceTeams, err := s.source.GetTeams(league.Id, int(league.CurrentSeason))
		if err != nil {
			return fmt.Errorf("error getting teams from apifootball %v", err)
		}
		for _, sourceTeam := range sourceTeams {
			_, err := s.dao.InsertTeam(&sourceTeam)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("error updating team in db %v", err)
			}
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}

func (s *SyncData) Fixtures() error {
	leagues, err := s.dao.GetLeagues()
	if err != nil {
		return fmt.Errorf("error getting leagues from db %v", err)
	}

	for _, league := range leagues {
		sourceFixtures, err := s.source.GetFixtures(league.Id, int(league.CurrentSeason))
		if err != nil {
			return fmt.Errorf("error getting fixtures from apifootball %v", err)
		}
		for _, fixture := range sourceFixtures {
			_, err := s.dao.InsertFixture(&fixture)
			if err != nil {
				return fmt.Errorf("error upserting fixtures in db %v", err)
			}
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}
