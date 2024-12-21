package db

import "time"

type GetFixturesFilter struct {
	LeagueId  int
	LeagueIds []int
	TeamId    int
	Date      time.Time
}

type Teams struct {
	Home Team `json:"home"`
	Away Team `json:"away"`
}

type Fixture struct {
	Id        int       `json:"id"`
	Referee   string    `json:"referee"`
	Date      time.Time `json:"date"`
	Timestamp int64     `json:"timestamp"`
	Teams     Teams     `json:"teams"`
	LeagueId  int       `json:"leagueId"`
	Season    int       `json:"season"`
	CreatedAt string    `json:"createdAt"`
}
