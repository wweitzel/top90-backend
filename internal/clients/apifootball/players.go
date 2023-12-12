package apifootball

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	db "github.com/wweitzel/top90/internal/db/models"
)

const playersUrl = baseUrl + "players"

func (c *Client) GetPlayer(id int, season int) (db.Player, error) {
	query := url.Values{}
	query.Set("id", strconv.Itoa(id))
	query.Set("season", strconv.Itoa(season))

	resp, err := c.doGet(playersUrl, query)
	if err != nil {
		return db.Player{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return db.Player{}, errors.New(resp.Status)
	}

	r := GetPlayersResponse{}
	err = json.NewDecoder(resp.Body).Decode(&r)

	players := r.toPlayers()
	if len(players) < 1 {
		return db.Player{}, fmt.Errorf("no player returned from apifootball")
	}
	return r.toPlayers()[0], err
}

func (c *Client) GetPlayers(teamId, season int) ([]db.Player, error) {
	var players []db.Player
	r, err := c.getPlayers(teamId, season, 1)
	if err != nil {
		return nil, err
	}
	players = append(players, r.toPlayers()...)

	for r.Paging.Current < r.Paging.Total {
		r, err = c.getPlayers(teamId, season, r.Paging.Current)
		if err != nil {
			return nil, err
		}
		players = append(players, r.toPlayers()...)
		r.Paging.Current++
	}
	return players, nil
}

func (c *Client) getPlayers(teamId, season int, page int) (GetPlayersResponse, error) {
	query := url.Values{}
	query.Set("team", strconv.Itoa(teamId))
	query.Set("season", strconv.Itoa(season))
	query.Set("page", strconv.Itoa(page))

	resp, err := c.doGet(playersUrl, query)
	if err != nil {
		return GetPlayersResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return GetPlayersResponse{}, errors.New(resp.Status)
	}

	r := &GetPlayersResponse{}
	err = json.NewDecoder(resp.Body).Decode(r)
	return *r, err
}
