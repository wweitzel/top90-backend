package apifootball

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
)

const teamsUrl = baseUrl + "teams"

type Team struct {
	Id        int      `json:"id"`
	Name      string   `json:"name"`
	Aliases   []string `json:"aliases"`
	Code      string   `json:"code"`
	Country   string   `json:"country"`
	Founded   int      `json:"founded"`
	National  bool     `json:"national"`
	Logo      string   `json:"logo"`
	CreatedAt string   `json:"createdAt"`
}

func (c *Client) GetTeams(league, season int) ([]Team, error) {
	query := url.Values{}
	query.Set("league", strconv.Itoa(league))
	query.Set("season", strconv.Itoa(season))

	resp, err := c.doGet(teamsUrl, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	r := &GetTeamsResponse{}
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		return nil, err
	}

	teams := r.toTeams()
	return teams, nil
}
