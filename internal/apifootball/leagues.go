package apifootball

import "errors"

const leaguesUrl = baseUrl + "leagues"

type League struct {
	Id        int
	Name      string
	Type      string
	Logo      string
	CreatedAt string
}

func (client *Client) GetLeague(country, leagueName string) (League, error) {
	req, err := client.newRequest("GET", leaguesUrl)
	if err != nil {
		return League{}, err
	}

	queryParams := req.URL.Query()
	queryParams.Set("country", country)
	queryParams.Set("name", leagueName)

	req.URL.RawQuery = queryParams.Encode()

	getLeaguesResponse := &GetLeaguesResponse{}
	resp, err := client.do(req, getLeaguesResponse)
	if err != nil {
		return League{}, err
	}

	if resp.StatusCode != 200 {
		return League{}, errors.New(resp.Status)
	}

	return toLeagues(*getLeaguesResponse)[0], nil
}

func toLeagues(response GetLeaguesResponse) []League {
	var leagues []League

	for _, l := range response.Data {
		league := League{}
		league.Id = l.League.ID
		league.Logo = l.League.Logo
		league.Name = l.League.Name
		league.Type = l.League.Type
		leagues = append(leagues, league)
	}

	return leagues
}
