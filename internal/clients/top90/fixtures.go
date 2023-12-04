package top90

import (
	"encoding/json"
	"net/url"

	"github.com/wweitzel/top90/internal/api/handlers"
)

const fixturesUrl = "/fixtures"

func (c *Client) GetFixtures(req handlers.GetFixturesRequest) (*handlers.GetFixturesResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	query := "?json=" + url.QueryEscape(string(jsonData))
	url := apiUrl + fixturesUrl + query

	resp, err := c.http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := &handlers.GetFixturesResponse{}
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		return nil, err
	}

	return r, nil
}
