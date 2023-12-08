package apifootball

import (
	"time"

	db "github.com/wweitzel/top90/internal/db/models"
)

type GetFixturesResponse struct {
	Get        string `json:"get"`
	Parameters struct {
		League string `json:"league"`
		Season string `json:"season"`
	} `json:"parameters"`
	Errors  []interface{} `json:"errors"`
	Results int           `json:"results"`
	Paging  struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"paging"`
	Data []struct {
		Fixture struct {
			ID        int       `json:"id"`
			Referee   string    `json:"referee"`
			Timezone  string    `json:"timezone"`
			Date      time.Time `json:"date"`
			Timestamp int64     `json:"timestamp"`
			Periods   struct {
				First  int `json:"first"`
				Second int `json:"second"`
			} `json:"periods"`
			Venue struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
				City string `json:"city"`
			} `json:"venue"`
			Status struct {
				Long    string `json:"long"`
				Short   string `json:"short"`
				Elapsed int    `json:"elapsed"`
			} `json:"status"`
		} `json:"fixture"`
		League struct {
			Id      int    `json:"id"`
			Name    string `json:"name"`
			Country string `json:"country"`
			Logo    string `json:"logo"`
			Flag    string `json:"flag"`
			Season  int    `json:"season"`
			Round   string `json:"round"`
		} `json:"league"`
		Teams struct {
			Home struct {
				Id     int    `json:"id"`
				Name   string `json:"name"`
				Logo   string `json:"logo"`
				Winner bool   `json:"winner"`
			} `json:"home"`
			Away struct {
				Id     int    `json:"id"`
				Name   string `json:"name"`
				Logo   string `json:"logo"`
				Winner bool   `json:"winner"`
			} `json:"away"`
		} `json:"teams"`
		Goals struct {
			Home int `json:"home"`
			Away int `json:"away"`
		} `json:"goals"`
		Score struct {
			Halftime struct {
				Home int `json:"home"`
				Away int `json:"away"`
			} `json:"halftime"`
			Fulltime struct {
				Home int `json:"home"`
				Away int `json:"away"`
			} `json:"fulltime"`
			Extratime struct {
				Home interface{} `json:"home"`
				Away interface{} `json:"away"`
			} `json:"extratime"`
			Penalty struct {
				Home interface{} `json:"home"`
				Away interface{} `json:"away"`
			} `json:"penalty"`
		} `json:"score"`
	} `json:"response"`
}

func (resp *GetFixturesResponse) toFixtures() []db.Fixture {
	var fixtures []db.Fixture

	for _, f := range resp.Data {
		fixture := db.Fixture{}
		fixture.Id = f.Fixture.ID
		fixture.Timestamp = f.Fixture.Timestamp
		fixture.Date = f.Fixture.Date
		fixture.Referee = f.Fixture.Referee
		fixture.Teams.Home.Id = f.Teams.Home.Id
		fixture.Teams.Home.Name = f.Teams.Home.Name
		fixture.Teams.Home.Logo = f.Teams.Home.Logo
		fixture.Teams.Away.Id = f.Teams.Away.Id
		fixture.Teams.Away.Name = f.Teams.Away.Name
		fixture.Teams.Away.Logo = f.Teams.Away.Logo
		fixture.LeagueId = f.League.Id
		fixture.Season = f.League.Season
		fixtures = append(fixtures, fixture)
	}

	return fixtures
}
