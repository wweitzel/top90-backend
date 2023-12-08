package apifootball

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	db "github.com/wweitzel/top90/internal/db/models"
)

const fixturesUrl = baseUrl + "fixtures"

func (c *Client) GetFixtures(league, season int) ([]db.Fixture, error) {
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
		return nil, err
	}

	fixtures := r.toFixtures()
	return fixtures, nil
}
