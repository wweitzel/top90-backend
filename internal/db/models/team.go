package db

import "github.com/lib/pq"

type GetTeamsFilter struct {
	Country    string
	SearchTerm string
}

type Team struct {
	Id        int            `json:"id" db:"id"`
	Name      string         `json:"name" db:"name"`
	Aliases   pq.StringArray `json:"aliases" db:"aliases"`
	Code      string         `json:"code" db:"code"`
	Country   string         `json:"country" db:"country"`
	Founded   int            `json:"founded" db:"founded"`
	National  bool           `json:"national" db:"national"`
	Logo      string         `json:"logo" db:"logo"`
	CreatedAt string         `json:"createdAt" db:"created_at"`
}
