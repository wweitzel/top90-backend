package apifootball

import (
	"strconv"
)

const teamsUrl = baseUrl + "teams"

func (client *Client) GetTeams(league, season int) (*GetTeamsResponse, error) {
	req, err := client.newRequest("GET", teamsUrl)
	if err != nil {
		return nil, err
	}

	queryParams := req.URL.Query()
	queryParams.Set("league", strconv.Itoa(league))
	queryParams.Set("season", strconv.Itoa(season))

	req.URL.RawQuery = queryParams.Encode()

	getTeamsResponse := &GetTeamsResponse{}
	_, err = client.do(req, getTeamsResponse)
	if err != nil {
		return nil, err
	}

	return getTeamsResponse, nil
}
