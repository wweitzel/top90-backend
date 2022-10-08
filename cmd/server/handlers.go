package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/chromedp/chromedp"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/apifootball"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/scrape"
)

type GetApiInfoResponse struct {
	Message string `json:"message"`
}

type GetGoalsResponse struct {
	Goals []top90.Goal `json:"goals"`
	Total int          `json:"total"`
}

type GetGoalsCrawlResponse struct {
	Goals []top90.Goal `json:"goals"`
}

type GetLeaguesResponse struct {
	Leagues []apifootball.League `json:"leagues"`
}

type GetTeamsResponse struct {
	Teams []apifootball.Team `json:"teams"`
}

func GetApiInfoHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	start := time.Now()

	var apiInfo GetApiInfoResponse
	apiInfo.Message = "Welcome to the top90 API ⚽️ 🥅 "

	w.Header().Add("Content-Type", "application/json")
	json, _ := json.Marshal(apiInfo)
	w.Write(json)
	log.Printf("%s %s %v", r.Method, r.URL, time.Since(start))
}

func GetGoalsHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	start := time.Now()

	queryParams := r.URL.Query()

	skip, _ := strconv.Atoi(queryParams.Get("skip"))
	limit, err := strconv.Atoi(queryParams.Get("limit"))
	if err != nil {
		limit = 10
	}

	search := queryParams.Get("search")
	filter := db.GetGoalsFilter{SearchTerm: search}
	count, err := dao.CountGoals(filter)
	if err != nil {
		log.Println(err)
	}

	goals, err := dao.GetGoals(db.Pagination{Skip: skip, Limit: limit}, filter)
	if err != nil {
		log.Println(err)
	}

	for i := range goals {
		url, err := s3Client.NewSignedGetURL(goals[i].S3ObjectKey, config.AwsBucketName, time.Minute*10)
		if err != nil {
			log.Println(err)
		}
		goals[i].PresignedUrl = url
	}

	getGoalsResponse := GetGoalsResponse{
		Goals: goals,
		Total: count,
	}

	w.Header().Add("Content-Type", "application/json")
	json, _ := json.Marshal(getGoalsResponse)
	w.Write(json)
	log.Printf("%s %s %v", r.Method, r.URL, time.Since(start))
}

func GetGoalsCrawlHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	start := time.Now()

	browserContext, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	scraper := scrape.Scraper{
		BrowserContext: browserContext,
	}

	skip := 0
	limit := 10
	goals, err := dao.GetGoals(db.Pagination{Skip: skip, Limit: limit}, db.GetGoalsFilter{})
	if err != nil {
		log.Println(err)
	}

	for i := range goals {
		log.Println(goals[i].RedditPostTitle)
		goals[i].PresignedUrl = scraper.GetVideoSourceUrl(goals[i].RedditLinkUrl)
	}

	w.Header().Add("Content-Type", "application/json")
	json, _ := json.Marshal(GetGoalsCrawlResponse{
		Goals: goals,
	})
	w.Write(json)
	log.Printf("%s %s %v", r.Method, r.URL, time.Since(start))
}

func GetLeaguesHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	start := time.Now()

	leagues, err := dao.GetLeagues()
	if err != nil {
		log.Println(err)
	}

	getLeaguesResponse := GetLeaguesResponse{
		Leagues: leagues,
	}

	w.Header().Add("Content-Type", "application/json")
	json, _ := json.Marshal(getLeaguesResponse)
	w.Write(json)
	log.Printf("%s %s %v", r.Method, r.URL, time.Since(start))
}

func GetTeamsHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	start := time.Now()

	teams, err := dao.GetTeams()
	if err != nil {
		log.Println(err)
	}

	getTeamsResponse := GetTeamsResponse{
		Teams: teams,
	}

	w.Header().Add("Content-Type", "application/json")
	json, _ := json.Marshal(getTeamsResponse)
	w.Write(json)
	log.Printf("%s %s %v", r.Method, r.URL, time.Since(start))
}
