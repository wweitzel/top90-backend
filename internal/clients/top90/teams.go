package top90

import (
	"encoding/json"
	"net/url"

	"github.com/wweitzel/top90/internal/api/handlers"
)

const teamsUrl = "/teams"

func (c Client) GetTeams(req handlers.GetTeamsRequest) (*handlers.GetTeamsResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	encoded := url.QueryEscape(string(jsonData))

	url := apiUrl + teamsUrl + "?json=" + encoded
	body, err := c.doGet(url)
	if err != nil {
		return nil, err
	}

	var response handlers.GetTeamsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
