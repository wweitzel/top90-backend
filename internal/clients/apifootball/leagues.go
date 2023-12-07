package apifootball

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
)

const leaguesUrl = baseUrl + "leagues"

type League struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Logo          string `json:"logo"`
	CreatedAt     string `json:"createdAt"`
	CurrentSeason int    `json:"currentSeason"`
}

func (c *Client) GetLeague(Id int) (*League, error) {
	query := url.Values{}
	query.Set("id", strconv.Itoa(Id))

	resp, err := c.doGet(leaguesUrl, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	r := &GetLeaguesResponse{}
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		return nil, err
	}

	leagues := r.toLeagues()
	if len(leagues) > 0 {
		return &leagues[0], nil
	}

	return nil, nil
}
