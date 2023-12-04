package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/db"
)

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

type FixtureHandler struct {
	dao db.Top90DAO
}

func NewFixtureHandler(dao db.Top90DAO) *FixtureHandler {
	return &FixtureHandler{
		dao: dao,
	}
}

func (h *FixtureHandler) GetFixture(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	fixtureId, err := strconv.Atoi(id)
	if err != nil {
		log.Println(err)
	}

	fixture, err := h.dao.GetFixture(fixtureId)
	if err != nil {
		log.Println(err)
	}

	respond(w, http.StatusOK, GetFixtureResponse{
		Fixture: fixture,
	})
}

func (h *FixtureHandler) GetFixtures(w http.ResponseWriter, r *http.Request) {
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

	fixtures, err := h.dao.GetFixtures(filter)
	if err != nil {
		log.Println(err)
	}

	respond(w, http.StatusOK, GetFixturesResponse{
		Fixtures: fixtures,
	})
}
