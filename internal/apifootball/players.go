package apifootball

import "strconv"

const playersUrl = baseUrl + "players"

func (client *Client) GetPlayers(league, season int) (*GetPlayersResponse, error) {
	req, err := client.newRequest("GET", playersUrl)
	if err != nil {
		return nil, err
	}

	queryParams := req.URL.Query()
	queryParams.Set("league", strconv.Itoa(league))
	queryParams.Set("season", strconv.Itoa(season))

	req.URL.RawQuery = queryParams.Encode()

	getPlayersResponse := &GetPlayersResponse{}
	_, err = client.do(req, getPlayersResponse)
	if err != nil {
		return nil, err
	}

	return getPlayersResponse, nil
}
