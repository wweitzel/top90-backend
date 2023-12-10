package apifootball

import db "github.com/wweitzel/top90/internal/db/models"

type GetPlayersResponse struct {
	Get        string `json:"get"`
	Parameters struct {
		League string `json:"league"`
		Season string `json:"season"`
	} `json:"parameters"`
	Errors  []interface{} `json:"errors"`
	Results int           `json:"results"`
	Paging  struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"paging"`
	Data []struct {
		Player struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Firstname string `json:"firstname"`
			Lastname  string `json:"lastname"`
			Age       int    `json:"age"`
			Birth     struct {
				Date    string `json:"date"`
				Place   string `json:"place"`
				Country string `json:"country"`
			} `json:"birth"`
			Nationality string `json:"nationality"`
			Height      string `json:"height"`
			Weight      string `json:"weight"`
			Injured     bool   `json:"injured"`
			Photo       string `json:"photo"`
		} `json:"player"`
		Statistics []struct {
			Team struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
				Logo string `json:"logo"`
			} `json:"team"`
			League struct {
				ID      int    `json:"id"`
				Name    string `json:"name"`
				Country string `json:"country"`
				Logo    string `json:"logo"`
				Flag    string `json:"flag"`
			} `json:"league"`
			Games struct {
				Appearences int         `json:"appearences"`
				Lineups     int         `json:"lineups"`
				Minutes     int         `json:"minutes"`
				Number      interface{} `json:"number"`
				Position    string      `json:"position"`
				Rating      interface{} `json:"rating"`
				Captain     bool        `json:"captain"`
			} `json:"games"`
			Substitutes struct {
				In    int `json:"in"`
				Out   int `json:"out"`
				Bench int `json:"bench"`
			} `json:"substitutes"`
			Shots struct {
				Total interface{} `json:"total"`
				On    interface{} `json:"on"`
			} `json:"shots"`
			Goals struct {
				Total    int         `json:"total"`
				Conceded int         `json:"conceded"`
				Assists  interface{} `json:"assists"`
				Saves    interface{} `json:"saves"`
			} `json:"goals"`
			Passes struct {
				Total    interface{} `json:"total"`
				Key      interface{} `json:"key"`
				Accuracy interface{} `json:"accuracy"`
			} `json:"passes"`
			Tackles struct {
				Total         interface{} `json:"total"`
				Blocks        interface{} `json:"blocks"`
				Interceptions interface{} `json:"interceptions"`
			} `json:"tackles"`
			Duels struct {
				Total interface{} `json:"total"`
				Won   interface{} `json:"won"`
			} `json:"duels"`
			Dribbles struct {
				Attempts interface{} `json:"attempts"`
				Success  interface{} `json:"success"`
				Past     interface{} `json:"past"`
			} `json:"dribbles"`
			Fouls struct {
				Drawn     interface{} `json:"drawn"`
				Committed interface{} `json:"committed"`
			} `json:"fouls"`
			Cards struct {
				Yellow    int `json:"yellow"`
				Yellowred int `json:"yellowred"`
				Red       int `json:"red"`
			} `json:"cards"`
			Penalty struct {
				Won      interface{} `json:"won"`
				Commited interface{} `json:"commited"`
				Scored   int         `json:"scored"`
				Missed   int         `json:"missed"`
				Saved    interface{} `json:"saved"`
			} `json:"penalty"`
		} `json:"statistics"`
	} `json:"response"`
}

func (resp *GetPlayersResponse) toPlayers() []db.Player {
	var players []db.Player
	for _, p := range resp.Data {
		player := db.Player{}
		player.Id = p.Player.ID
		player.Name = p.Player.Name
		player.FirstName = p.Player.Firstname
		player.LastName = p.Player.Lastname
		player.Age = p.Player.Age
		player.Nationality = p.Player.Nationality
		player.Height = p.Player.Height
		player.Weight = p.Player.Weight
		player.Photo = p.Player.Photo
		players = append(players, player)
	}
	return players
}
