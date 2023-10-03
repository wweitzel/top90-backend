package server

import "github.com/gorilla/mux"

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", GetApiInfoHandler)
	r.HandleFunc("/fixtures", GetFixturesHandler)
	r.HandleFunc("/goals", GetGoalsHandler)
	r.HandleFunc("/goals/{id}", GetGoalHandler)
	r.HandleFunc("/goals_crawl", GetGoalsCrawlHandler)
	r.HandleFunc("/leagues", GetLeaguesHandler)
	r.HandleFunc("/teams", GetTeamsHandler)
	r.HandleFunc("/message_preview/{id}", MessagePreviewHandler)
	return r
}
