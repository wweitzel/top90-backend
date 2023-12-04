package apifootball

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
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

func (c *Client) GetFixtures(league, season int) ([]Fixture, error) {
	query := url.Values{}
	query.Set("league", strconv.Itoa(league))
	query.Set("season", strconv.Itoa(season))

	resp, err := c.doGet(fixturesUrl, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	r := &GetFixturesResponse{}
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		log.Println(err)
	}

	fixtures := r.toFixtures()
	return fixtures, nil
}
