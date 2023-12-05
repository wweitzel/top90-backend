package top90

import (
	"encoding/json"
	"net/url"

	"github.com/wweitzel/top90/internal/api/handlers"
)

const teamsUrl = "/teams"

func (c *Client) GetTeams(req handlers.GetTeamsRequest) (*handlers.GetTeamsResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	query := "?json=" + url.QueryEscape(string(jsonData))
	url := apiUrl + teamsUrl + query

	resp, err := c.http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := &handlers.GetTeamsResponse{}
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		return nil, err
	}

	return r, nil
}
