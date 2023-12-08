package db

type League struct {
	Id            int     `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	Type          string  `json:"type" db:"type"`
	Logo          string  `json:"logo" db:"logo"`
	CreatedAt     string  `json:"createdAt" db:"created_at"`
	CurrentSeason NullInt `json:"currentSeason" db:"current_season"`
}
