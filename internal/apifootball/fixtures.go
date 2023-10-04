package apifootball

import (
	"errors"
	"strconv"
	"time"
)

const fixturesUrl = baseUrl + "fixtures"

type Teams struct {
	Home Team `json:"home"`
	Away Team `json:"away"`
}

type Fixture struct {
	Id        int       `json:"id"`
	Referee   string    `json:"referee"`
	Date      time.Time `json:"date"`
	Timestamp int64     `json:"timestamp"`
	Teams     Teams     `json:"teams"`
	LeagueId  int       `json:"leagueId"`
	Season    int       `json:"season"`
	CreatedAt string    `json:"createdAt"`
}

func (client *Client) GetFixtures(league, season int) ([]Fixture, error) {
	req, err := client.newRequest("GET", fixturesUrl)
	if err != nil {
		return nil, err
	}

	queryParams := req.URL.Query()
	queryParams.Set("league", strconv.Itoa(league))
	queryParams.Set("season", strconv.Itoa(season))

	req.URL.RawQuery = queryParams.Encode()

	getFixturesResponse := &GetFixturesResponse{}
	resp, err := client.do(req, getFixturesResponse)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	var fixtures = toFixtures(getFixturesResponse)

	return fixtures, nil
}

func toFixtures(response *GetFixturesResponse) []Fixture {
	var fixtures []Fixture

	for _, f := range response.Data {
		fixture := Fixture{}
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
