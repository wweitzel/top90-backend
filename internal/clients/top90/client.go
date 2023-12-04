package top90

import (
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
