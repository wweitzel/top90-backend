package top90

import (
	"io"
	"net/http"
	"time"
)

type Client struct {
	http http.Client
}

type Config struct {
	Timeout time.Duration
}

const apiUrl = "https://api.top90.io"

func NewClient(cfg Config) Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	c := http.Client{Timeout: cfg.Timeout}

	return Client{
		http: c,
	}
}

func (c Client) doGet(url string) (body []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
