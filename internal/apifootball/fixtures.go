package apifootball

import (
	"errors"
	"strconv"
	"time"
)

const fixturesUrl = baseUrl + "fixtures"

type Fixture struct {
	Id        int
	Referee   string
	Date      time.Time
	Timestamp int64
	Teams     struct {
		Home struct {
			Id     int
			Name   string
			Logo   string
			Winner bool
		}
		Away struct {
			Id     int
			Name   string
			Logo   string
			Winner bool
		}
	}
	LeagueId  int
	Season    int
	CreatedAt string
}

func (client *Client) GetFixtures(league, season int) ([]Fixture, error) {
	req, err := client.newRequest("GET", fixturesUrl)
	if err != nil {
		return nil, err
	}

	queryParams := req.URL.Query()
	queryParams.Set("league", strconv.Itoa(league))
	queryParams.Set("season", strconv.Itoa(season))

	req.URL.RawQuery = queryParams.Encode()

	getFixturesResponse := &GetFixturesResponse{}
	resp, err := client.do(req, getFixturesResponse)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	var fixtures = toFixtures(getFixturesResponse)

	return fixtures, nil
}

func toFixtures(response *GetFixturesResponse) []Fixture {
	var fixtures []Fixture

	for _, f := range response.Data {
		fixture := Fixture{}
		fixture.Id = f.Fixture.ID
		fixture.Timestamp = f.Fixture.Timestamp
		fixture.Date = f.Fixture.Date
		fixture.Referee = f.Fixture.Referee
		fixture.Teams = struct {
			Home struct {
				Id     int
				Name   string
				Logo   string
				Winner bool
			}
			Away struct {
				Id     int
				Name   string
				Logo   string
				Winner bool
			}
		}(f.Teams)
		fixture.LeagueId = f.League.Id
		fixture.Season = f.League.Season
		fixtures = append(fixtures, fixture)
	}

	return fixtures
}
