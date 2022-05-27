package apifootball

const leaguesUrl = baseUrl + "leagues"

func GetLeagues(client *Client) (*GetLeaguesResponse, error) {
	req, err := client.newRequest("GET", leaguesUrl)
	if err != nil {
		return nil, err
	}

	getLeaguesResponse := &GetLeaguesResponse{}
	_, err = client.do(req, getLeaguesResponse)
	if err != nil {
		return nil, err
	}

	return getLeaguesResponse, nil
}
