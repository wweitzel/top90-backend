package reddit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type RedditClient struct {
	http            *http.Client
	accessTokenInfo RedditAccessTokenInfo
}

type RedditAccessTokenInfo struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	DeviceId    string `json:"device_id"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

const redditAccessTokenUrl = `https://www.reddit.com/api/v1/access_token`

var redditBasicAuth = os.Getenv("TOP90_REDDIT_BASIC_AUTH")

func NewRedditClient(httpClient *http.Client) RedditClient {
	client := &http.Client{}

	reqBody := "grant_type=client_credentials"

	req, err := http.NewRequest("POST", redditAccessTokenUrl, bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		log.Println(err)
	}

	req.Header.Add("User-Agent", "browser:top90:v0.0 (by /u/top90app)")
	req.Header.Add("Authorization", redditBasicAuth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	var accessTokenInfo RedditAccessTokenInfo
	err = json.Unmarshal(body, &accessTokenInfo)
	if err != nil {
		log.Fatal(err)
	}

	return RedditClient{
		http:            client,
		accessTokenInfo: accessTokenInfo,
	}
}

func (client *RedditClient) doGet(url string) (body []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}

	req.Header.Add("User-Agent", "browser:top90:v0.0 (by /r/top90app)")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", client.accessTokenInfo.AccessToken))

	resp, err := client.http.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
