package apifootball

import (
	"encoding/json"
	"net/http"
)

type Client struct {
	httpClient *http.Client
	host       string
	apiKey     string
}

const baseUrl = "https://api-football-v1.p.rapidapi.com/v3/"

func NewClient(host, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &Client{
		httpClient: httpClient,
		host:       host,
		apiKey:     apiKey,
	}
}

func (client *Client) newRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-RapidAPI-Host", client.host)
	req.Header.Add("X-RapidAPI-Key", client.apiKey)
	return req, nil
}

func (client *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}
