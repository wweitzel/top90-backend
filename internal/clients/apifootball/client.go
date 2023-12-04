package apifootball

import (
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	http   http.Client
	host   string
	apiKey string
}

type Config struct {
	Host    string
	ApiKey  string
	Timeout time.Duration
}

const baseUrl = "https://api-football-v1.p.rapidapi.com/v3/"

func NewClient(cfg Config) Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	c := http.Client{Timeout: cfg.Timeout}

	return Client{
		http:   c,
		host:   cfg.Host,
		apiKey: cfg.ApiKey,
	}
}

func (c *Client) doGet(url string, query url.Values) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query.Encode()
	req.Header.Add("X-RapidAPI-Host", c.host)
	req.Header.Add("X-RapidAPI-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
