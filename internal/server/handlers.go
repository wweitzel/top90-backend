package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
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

type GetFixturesRequest struct {
	LeagueId  int  `json:"leagueId"`
	TodayOnly bool `json:"todayOnly"`
}

type GetFixtureResponse struct {
	Fixture apifootball.Fixture `json:"fixture"`
}

type GetGoalsResponse struct {
	Goals []top90.Goal `json:"goals"`
	Total int          `json:"total"`
}

type GetGoalsRequest struct {
	Pagination db.Pagination     `json:"pagination"`
	Filter     db.GetGoalsFilter `json:"filter"`
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

type GetTeamsRequest struct {
	LeagueId   int    `json:"leagueId"`
	Season     int    `json:"season"`
	SearchTerm string `json:"searchTerm"`
}

type GetTeamsResponse struct {
	Teams []apifootball.Team `json:"teams"`
}

func (s *Server) GetApiInfoHandler(w http.ResponseWriter, r *http.Request) {
	var apiInfo GetApiInfoResponse
	apiInfo.Message = "Welcome to the top90 API ⚽️ 🥅 "
	respond(w, http.StatusOK, apiInfo)
}

func (s *Server) GetFixturesHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	json := queryParams.Get("json")

	request, err := unmarshal[GetFixturesRequest](json)
	if err != nil {
		respond(w, http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	if request.LeagueId == 0 && !request.TodayOnly {
		respond(w, http.StatusBadRequest, ErrorResponse{Message: "Bad request. leagueId query param must be set if todayOnly is not true."})
		return
	}

	var filter db.GetFixuresFilter
	filter.LeagueId = request.LeagueId

	if request.TodayOnly {
		filter.Date = time.Now()
	}

	fixtures, err := s.dao.GetFixtures(filter)
	if err != nil {
		log.Println(err)
	}

	respond(w, http.StatusOK, GetFixturesResponse{
		Fixtures: fixtures,
	})
}

func (s *Server) GetFixtureHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	fixtureId, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
	}

	fixture, err := s.dao.GetFixture(fixtureId)
	if err != nil {
		log.Println(err)
	}

	respond(w, http.StatusOK, GetFixtureResponse{
		Fixture: fixture,
	})
}

func (s *Server) GetGoalsHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	json := queryParams.Get("json")

	request, err := unmarshal[GetGoalsRequest](json)
	if err != nil {
		respond(w, http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	if request.Pagination.Limit == 0 {
		request.Pagination.Limit = 10
	}

	count, err := s.dao.CountGoals(request.Filter)
	if err != nil {
		log.Println(err)
	}

	goals, err := s.dao.GetGoals(request.Pagination, request.Filter)
	if err != nil {
		log.Println(err)
	}

	for i := range goals {
		goals[i].PresignedUrl = s.presignedUrl(goals[i].S3ObjectKey)
		goals[i].ThumbnailPresignedUrl = s.presignedUrl(goals[i].ThumbnailS3Key)
	}

	respond(w, http.StatusOK, GetGoalsResponse{
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

	respond(w, http.StatusOK, GetGoalResponse{
		Goal: goal,
	})
}

func (s *Server) GetLeaguesHandler(w http.ResponseWriter, r *http.Request) {
	leagues, err := s.dao.GetLeagues()
	if err != nil {
		log.Println(err)
	}

	respond(w, http.StatusOK, GetLeaguesResponse{
		Leagues: leagues,
	})
}

func (s *Server) GetTeamsHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	json := queryParams.Get("json")

	request, err := unmarshal[GetTeamsRequest](json)
	if err != nil {
		respond(w, http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	var teams []apifootball.Team
	if request.SearchTerm != "" {
		teams, err = s.dao.GetTeams(db.GetTeamsFilter{SearchTerm: request.SearchTerm})
	} else {
		teams, err = s.dao.GetTeamsForLeagueAndSeason(request.LeagueId, request.Season)
	}

	if err != nil {
		log.Println(err)
	}

	respond(w, http.StatusOK, GetTeamsResponse{
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

func respond(w http.ResponseWriter, statusCode int, response any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json, _ := json.Marshal(response)
	w.Write(json)
}

func unmarshal[T any](jsonStr string) (*T, error) {
	out := new(T)

	decodedJson, err := url.QueryUnescape(jsonStr)
	if err != nil {
		return nil, errors.New("error decoding json")
	}

	err = json.Unmarshal([]byte(decodedJson), &out)
	if err != nil {
		return nil, errors.New("error unmarshalling json")
	}

	return out, nil
}
