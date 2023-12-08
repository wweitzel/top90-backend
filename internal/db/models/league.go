package db

type League struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Logo          string `json:"logo"`
	CreatedAt     string `json:"createdAt"`
	CurrentSeason int    `json:"currentSeason"`
}
