package apifootball

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"strconv"
)

const playersUrl = baseUrl + "players"

type Player struct{}

func (c *Client) GetPlayers(league, season int) ([]Player, error) {
	query := url.Values{}
	query.Set("league", strconv.Itoa(league))
	query.Set("season", strconv.Itoa(season))

	resp, err := c.doGet(playersUrl, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	r := &GetPlayersResponse{}
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		log.Println(err)
	}

	players := r.toPlayers()
	return players, nil
}
