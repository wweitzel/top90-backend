package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	top90 "github.com/wweitzel/top90/internal"
	"github.com/wweitzel/top90/internal/apifootball"
	"github.com/wweitzel/top90/internal/db"
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

func (s *Server) GetApiInfoHandler(w http.ResponseWriter, r *http.Request) {
	var apiInfo GetApiInfoResponse
	apiInfo.Message = "Welcome to the top90 API ⚽️ 🥅 "
	respond(w, apiInfo)
}

func (s *Server) GetFixturesHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	leagueParam := queryParams.Get("leagueId")
	if leagueParam == "" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(400)
		json, _ := json.Marshal(ErrorResponse{Message: "Bad request. leagueId query param must be set."})
		w.Write(json)
		return
	}

	leagueId, err := strconv.Atoi(leagueParam)
	if err != nil {
		log.Panicf("Failed converting leagueId query param to int %v", err)
	}

	filter := db.GetFixuresFilter{LeagueId: leagueId}

	fixtures, err := s.dao.GetFixtures(filter)
	if err != nil {
		log.Println(err)
	}

	respond(w, GetFixturesResponse{
		Fixtures: fixtures,
	})
}

func (s *Server) GetGoalsHandler(w http.ResponseWriter, r *http.Request) {
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
	count, err := s.dao.CountGoals(filter)
	if err != nil {
		log.Println(err)
	}

	goals, err := s.dao.GetGoals(db.Pagination{Skip: skip, Limit: limit}, filter)
	if err != nil {
		log.Println(err)
	}

	for i := range goals {
		goals[i].PresignedUrl = s.presignedUrl(goals[i].S3ObjectKey)
		goals[i].ThumbnailPresignedUrl = s.presignedUrl(goals[i].ThumbnailS3Key)
	}

	respond(w, GetGoalsResponse{
		Goals: goals,
		Total: count,
	})
}

func (s *Server) GetGoalHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	goal, err := s.dao.GetGoal(id)
	if err != nil {
		log.Println(err)
	}

	goal.PresignedUrl = s.presignedUrl(goal.S3ObjectKey)
	goal.ThumbnailPresignedUrl = s.presignedUrl(goal.ThumbnailS3Key)

	respond(w, GetGoalResponse{
		Goal: goal,
	})
}

func (s *Server) GetLeaguesHandler(w http.ResponseWriter, r *http.Request) {
	leagues, err := s.dao.GetLeagues()
	if err != nil {
		log.Println(err)
	}

	respond(w, GetLeaguesResponse{
		Leagues: leagues,
	})
}

func (s *Server) GetTeamsHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	leagueId, _ := strconv.Atoi(queryParams.Get("leagueId"))
	season, _ := strconv.Atoi(queryParams.Get("season"))
	searchTerm := queryParams.Get("searchTerm")

	var teams []apifootball.Team
	var err error

	if searchTerm != "" {
		teams, err = s.dao.GetTeams(db.GetTeamsFilter{SearchTerm: searchTerm})
	} else {
		teams, err = s.dao.GetTeamsForLeagueAndSeason(leagueId, season)
	}

	if err != nil {
		log.Println(err)
	}

	respond(w, GetTeamsResponse{
		Teams: teams,
	})
}

func (s *Server) MessagePreviewHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	goal, _ := s.dao.GetGoal(id)

	// TODO: This is nearly duplicated with the code in getGoals and getGoal
	videoUrl, err := s.s3Client.NewSignedGetURL(goal.S3ObjectKey, s.config.AwsBucketName, time.Hour*24*7)
	if err != nil {
		log.Println(err)
	}
	thumbnailUrl, err := s.s3Client.NewSignedGetURL(goal.ThumbnailS3Key, s.config.AwsBucketName, time.Hour*24*7)
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
}

func (s *Server) presignedUrl(objectKey string) string {
	url, err := s.s3Client.NewSignedGetURL(objectKey, s.config.AwsBucketName, time.Minute*10)
	if err != nil {
		log.Println(err)
	}
	return url
}

func respond(w http.ResponseWriter, response any) {
	w.Header().Add("Content-Type", "application/json")
	json, _ := json.Marshal(response)
	w.Write(json)
}
