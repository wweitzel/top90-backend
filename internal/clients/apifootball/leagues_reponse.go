package apifootball

import db "github.com/wweitzel/top90/internal/db/models"

type GetLeaguesResponse struct {
	Get        string `json:"get"`
	Parameters struct {
		Country string `json:"country"`
		Season  string `json:"season"`
	} `json:"parameters"`
	Errors  []interface{} `json:"errors"`
	Results int           `json:"results"`
	Paging  struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"paging"`
	Data []struct {
		League struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			Logo string `json:"logo"`
		} `json:"league"`
		Country struct {
			Name string      `json:"name"`
			Code interface{} `json:"code"`
			Flag interface{} `json:"flag"`
		} `json:"country"`
		Seasons []struct {
			Year     int    `json:"year"`
			Start    string `json:"start"`
			End      string `json:"end"`
			Current  bool   `json:"current"`
			Coverage struct {
				Fixtures struct {
					Events             bool `json:"events"`
					Lineups            bool `json:"lineups"`
					StatisticsFixtures bool `json:"statistics_fixtures"`
					StatisticsPlayers  bool `json:"statistics_players"`
				} `json:"fixtures"`
				Standings   bool `json:"standings"`
				Players     bool `json:"players"`
				TopScorers  bool `json:"top_scorers"`
				TopAssists  bool `json:"top_assists"`
				TopCards    bool `json:"top_cards"`
				Injuries    bool `json:"injuries"`
				Predictions bool `json:"predictions"`
				Odds        bool `json:"odds"`
			} `json:"coverage"`
		} `json:"seasons"`
	} `json:"response"`
}

func (resp *GetLeaguesResponse) toLeagues() []db.League {
	var leagues []db.League
	for _, l := range resp.Data {
		league := db.League{}
		league.Id = l.League.ID
		league.Logo = l.League.Logo
		league.Name = l.League.Name
		league.Type = l.League.Type
		for _, season := range l.Seasons {
			if season.Current {
				league.CurrentSeason = db.NullInt(season.Year)
			}
		}
		leagues = append(leagues, league)
	}
	return leagues
}
