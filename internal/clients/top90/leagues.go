package top90

import (
	"encoding/json"

	"github.com/wweitzel/top90/internal/api/handlers"
)

const leaguesUrl = "/leagues"

func (c *Client) GetLeagues() (*handlers.GetLeaguesResponse, error) {
	url := apiUrl + leaguesUrl + "?json={}"

	resp, err := c.http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := &handlers.GetLeaguesResponse{}
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		return nil, err
	}

	return r, nil
}
