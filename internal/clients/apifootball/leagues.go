package apifootball

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	db "github.com/wweitzel/top90/internal/db/models"
)

const leaguesUrl = baseUrl + "leagues"

func (c *Client) GetLeague(Id int) (*db.League, error) {
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
