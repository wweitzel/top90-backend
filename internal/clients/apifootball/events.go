package apifootball

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
)

const eventsUrl = baseUrl + "fixtures/events"

func (c *Client) GetEvents(fixture int) ([]Event, error) {
	query := url.Values{}
	query.Set("fixture", strconv.Itoa(fixture))

	resp, err := c.doGet(eventsUrl, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	r := &GetEventsResponse{}
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		return nil, err
	}
	return r.Data, err
}
