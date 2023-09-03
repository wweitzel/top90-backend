package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gorilla/mux"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/apifootball"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/scrape"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type GetApiInfoResponse struct {
	Message string `json:"message"`
}

type GetFixturesResponse struct {
	Fixtures []apifootball.Fixture `json:"fixtures"`
}

type GetGoalsResponse struct {
	Goals []top90.Goal `json:"goals"`
	Total int          `json:"total"`
}

type GetGoalResponse struct {
	Goal top90.Goal `json:"goal"`
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

func GetFixturesHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	start := time.Now()

	queryParams := r.URL.Query()

	leagueParam := queryParams.Get("leagueId")
	if leagueParam == "" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(400)
		json, _ := json.Marshal(ErrorResponse{Message: "Bad request. leagueId query param must be set."})
		w.Write(json)
		log.Printf("%s %s %v", r.Method, r.URL, time.Since(start))
		return
	}

	leagueId, err := strconv.Atoi(leagueParam)
	if err != nil {
		log.Panicf("Failed converting leagueId query param to int %v", err)
	}

	filter := db.GetFixuresFilter{LeagueId: leagueId}

	fixtures, err := dao.GetFixtures(filter)
	if err != nil {
		log.Println(err)
	}

	getFixturesResponse := GetFixturesResponse{
		Fixtures: fixtures,
	}

	w.Header().Add("Content-Type", "application/json")
	json, _ := json.Marshal(getFixturesResponse)
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
	leagueId, _ := strconv.Atoi(queryParams.Get("leagueId"))
	season, _ := strconv.Atoi(queryParams.Get("season"))
	teamId, _ := strconv.Atoi(queryParams.Get("teamId"))

	filter := db.GetGoalsFilter{SearchTerm: search, LeagueId: leagueId, Season: season, TeamId: teamId}
	count, err := dao.CountGoals(filter)
	if err != nil {
		log.Println(err)
	}

	goals, err := dao.GetGoals(db.Pagination{Skip: skip, Limit: limit}, filter)
	if err != nil {
		log.Println(err)
	}

	for i := range goals {
		videoUrl, err := s3Client.NewSignedGetURL(goals[i].S3ObjectKey, config.AwsBucketName, time.Minute*10)
		if err != nil {
			log.Println(err)
		}
		thumbnailUrl, err := s3Client.NewSignedGetURL(goals[i].ThumbnailS3Key, config.AwsBucketName, time.Minute*10)
		if err != nil {
			log.Println(err)
		}

		goals[i].PresignedUrl = videoUrl
		goals[i].ThumbnailPresignedUrl = thumbnailUrl
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

func GetGoalHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	start := time.Now()

	id := mux.Vars(r)["id"]

	goal, err := dao.GetGoal(id)
	if err != nil {
		log.Println(err)
	}

	// TODO: This is nearly duplicated with the code in getGoals
	videoUrl, err := s3Client.NewSignedGetURL(goal.S3ObjectKey, config.AwsBucketName, time.Minute*1)
	if err != nil {
		log.Println(err)
	}
	thumbnailUrl, err := s3Client.NewSignedGetURL(goal.ThumbnailS3Key, config.AwsBucketName, time.Minute*10)
	if err != nil {
		log.Println(err)
	}

	goal.PresignedUrl = videoUrl
	goal.ThumbnailPresignedUrl = thumbnailUrl

	getGoalResponse := GetGoalResponse{
		Goal: goal,
	}

	w.Header().Add("Content-Type", "application/json")
	json, _ := json.Marshal(getGoalResponse)
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

	queryParams := r.URL.Query()
	leagueId, _ := strconv.Atoi(queryParams.Get("leagueId"))
	season, _ := strconv.Atoi(queryParams.Get("season"))
	searchTerm := queryParams.Get("searchTerm")

	var teams []apifootball.Team
	var err error

	if searchTerm != "" {
		teams, err = dao.GetTeams(db.GetTeamsFilter{SearchTerm: searchTerm})
	} else {
		teams, err = dao.GetTeamsForLeagueAndSeason(leagueId, season)
	}

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

func MessagePreviewHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	start := time.Now()

	id := mux.Vars(r)["id"]
	goal, _ := dao.GetGoal(id)

	// TODO: This is nearly duplicated with the code in getGoals and getGoal
	videoUrl, err := s3Client.NewSignedGetURL(goal.S3ObjectKey, config.AwsBucketName, time.Hour*24*7)
	if err != nil {
		log.Println(err)
	}
	thumbnailUrl, err := s3Client.NewSignedGetURL(goal.ThumbnailS3Key, config.AwsBucketName, time.Hour*24*7)
	if err != nil {
		log.Println(err)
	}

	goal.PresignedUrl = videoUrl
	goal.ThumbnailPresignedUrl = thumbnailUrl

	html :=
		`<!doctype html>` +
			`<html lang="en">` +
			`<head>` +
			` <meta charset="utf-8">` +
			// TODO: If possible, better to return the video and image data directly (in base64)
			// rather than returning the presigned url since we don't really want these to expire.
			// Max expiry date for presigned url is apparently one week.
			`	<meta property="og:title" content="` + goal.RedditPostTitle + `"` + `>` +
			`	<meta property="og:image" content="` + goal.ThumbnailPresignedUrl + `"` + `>` +
			`	<meta property="og:video" content="` + goal.PresignedUrl + `"` + `>` +
			` <meta http-equiv="refresh" content="0; url='https://top90.io/goals/` + goal.Id + `'" />` +
			` <title>top90.io</title>` +
			`</head>` +
			`<body>` +
			`</body>` +
			`</html>`

	w.Header().Add("Content-Type", "text/html")
	w.Write([]byte(html))
	log.Printf("%s %s %v", r.Method, r.URL, time.Since(start))
}
