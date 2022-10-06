package apifootball

import (
	"strconv"
	"time"
)

const fixturesUrl = baseUrl + "fixtures"

type Teams struct {
	Home struct {
		ID     int
		Name   string
		Logo   string
		Winner bool
	}
	Away struct {
		ID     int
		Name   string
		Logo   string
		Winner bool
	}
}

type Fixture struct {
	ID        int
	Referee   string
	Timezone  string
	Date      time.Time
	Timestamp int
	Periods   struct {
		First  int
		Second int
	}
	Venue struct {
		ID   int
		Name string
		City string
	}
	Status struct {
		Long    string
		Short   string
		Elapsed int
	}
	Teams Teams
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
	_, err = client.do(req, getFixturesResponse)
	if err != nil {
		return nil, err
	}

	var fixtures = toFixtures(getFixturesResponse)

	return fixtures, nil
}

func toFixtures(response *GetFixturesResponse) []Fixture {
	var fixtures []Fixture

	for _, f := range response.Data {
		fixture := Fixture{}
		fixture.ID = f.Fixture.ID
		fixture.Date = f.Fixture.Date
		fixture.Status = struct {
			Long    string
			Short   string
			Elapsed int
		}(f.Fixture.Status)
		fixture.Timestamp = f.Fixture.Timestamp
		fixture.Referee = f.Fixture.Referee
		fixture.Periods = struct {
			First  int
			Second int
		}(f.Fixture.Periods)
		fixture.Venue = struct {
			ID   int
			Name string
			City string
		}(f.Fixture.Venue)
		fixture.Teams = Teams(f.Teams)
		fixtures = append(fixtures, fixture)
	}

	return fixtures
}
