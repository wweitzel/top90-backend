package apifootball

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	db "github.com/wweitzel/top90/internal/db/models"
)

const teamsUrl = baseUrl + "teams"

func (c *Client) GetTeams(league, season int) ([]db.Team, error) {
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
