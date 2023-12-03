package top90

import (
	"encoding/json"
	"net/url"

	"github.com/wweitzel/top90/internal/api/handlers"
)

const fixturesUrl = "/fixtures"

func (c Client) GetFixtures(req handlers.GetFixturesRequest) (*handlers.GetFixturesResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	encoded := url.QueryEscape(string(jsonData))

	url := apiUrl + fixturesUrl + "?json=" + encoded
	body, err := c.doGet(url)
	if err != nil {
		return nil, err
	}

	var response handlers.GetFixturesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
