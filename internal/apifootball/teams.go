package apifootball

import (
	"errors"
	"strconv"
)

const teamsUrl = baseUrl + "teams"

type Team struct {
	Id        int
	Name      string
	Code      string
	Country   string
	Founded   int
	National  bool
	Logo      string
	CreatedAt string
}

func (client *Client) GetTeams(league, season int) ([]Team, error) {
	req, err := client.newRequest("GET", teamsUrl)
	if err != nil {
		return nil, err
	}

	queryParams := req.URL.Query()
	queryParams.Set("league", strconv.Itoa(league))
	queryParams.Set("season", strconv.Itoa(season))

	req.URL.RawQuery = queryParams.Encode()

	getTeamsResponse := &GetTeamsResponse{}
	resp, err := client.do(req, getTeamsResponse)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	var teams = toTeams(getTeamsResponse)

	return teams, nil
}

func toTeams(response *GetTeamsResponse) []Team {
	var teams []Team

	for _, t := range response.Data {
		team := Team{}
		team.Id = t.Team.ID
		team.Name = t.Team.Name
		team.Code = t.Team.Code
		team.Country = t.Team.Country
		team.Founded = t.Team.Founded
		team.National = t.Team.National
		team.Logo = t.Team.Logo
		teams = append(teams, team)
	}

	return teams
}
