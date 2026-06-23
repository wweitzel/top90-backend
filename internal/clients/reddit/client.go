package reddit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Client struct {
	http   http.Client
	token  AccessToken
	logger *slog.Logger
}

type Config struct {
	Timeout time.Duration
	Logger  *slog.Logger
}

type AccessToken struct {
	Token     string `json:"access_token"`
	Type      string `json:"token_type"`
	DeviceId  string `json:"device_id"`
	ExpiresIn int    `json:"expires_in"`
	Scope     string `json:"scope"`
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	c := http.Client{Timeout: cfg.Timeout}

	url := "https://www.reddit.com/api/v1/access_token"
	body := bytes.NewBuffer([]byte("grant_type=client_credentials"))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "browser:top90:v0.1 (by /u/top90app)")
	req.Header.Add("Authorization", "redditbasicauth")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed getting reddit access token: %v", err)
	}
	defer resp.Body.Close()

	token := &AccessToken{}
	err = json.NewDecoder(resp.Body).Decode(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode reddit access token")
	}

	return &Client{
		http:   c,
		token:  *token,
		logger: cfg.Logger,
	}, nil
}

func (c *Client) doGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "browser:top90:v0.1 (by /r/top90app)")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token.Token))

	c.logger.Debug("Making request", "token", c.token.Token, "url", url)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	// Log the full response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", "error", err)
		return resp, nil
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	c.logger.Debug("Response received",
		"status_code", resp.Status,
		"status", resp.StatusCode,
		"headers", resp.Header,
		"body", string(respBody))

	return resp, nil
}

func (c *Client) doBrowserGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="149", "Chromium";v="149", "Not)A;Brand";v="24"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36")

	c.logger.Debug("Making browser-style request", "url", url)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read browser response body", "error", err)
		return resp, nil
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	c.logger.Debug("Browser response received",
		"status_code", resp.Status,
		"status", resp.StatusCode,
		"headers", resp.Header,
		"body", string(respBody))

	return resp, nil
}
