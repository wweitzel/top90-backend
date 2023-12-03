package top90

import (
	"encoding/json"

	"github.com/wweitzel/top90/internal/api/handlers"
)

const leaguesUrl = "/leagues"

func (c Client) GetLeagues() (*handlers.GetLeaguesResponse, error) {
	url := apiUrl + leaguesUrl + "?json={}"
	body, err := c.doGet(url)
	if err != nil {
		return nil, err
	}

	var response handlers.GetLeaguesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
