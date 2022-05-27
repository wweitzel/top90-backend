package apifootball

import (
	"strconv"
)

const fixturesUrl = baseUrl + "fixtures"

func (client *Client) GetFixtures(league, season int) (*GetFixturesResponse, error) {
	req, err := client.newRequest("GET", fixturesUrl)
	if err != nil {
		return nil, err
	}

	queryParams := req.URL.Query()
	queryParams.Set("league", strconv.Itoa(league))
	queryParams.Set("season", strconv.Itoa(season))

	req.URL.RawQuery = queryParams.Encode()

	getFixturesResponse := &GetFixturesResponse{}
	_, err = client.do(req, getFixturesResponse)
	if err != nil {
		return nil, err
	}

	return getFixturesResponse, nil
}
