package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wweitzel/top90/internal/db/dao"
	db "github.com/wweitzel/top90/internal/db/models"
)

type GetFixturesResponse struct {
	Fixtures []db.Fixture `json:"fixtures"`
}

type GetFixturesRequest struct {
	LeagueId  int  `json:"leagueId"`
	TeamId    int  `json:"teamId"`
	TodayOnly bool `json:"todayOnly"`
}

type GetFixtureResponse struct {
	Fixture db.Fixture `json:"fixture"`
}

type FixtureHandler struct {
	dao dao.Top90DAO
}

func NewFixtureHandler(dao dao.Top90DAO) *FixtureHandler {
	return &FixtureHandler{
		dao: dao,
	}
}

func (h *FixtureHandler) GetFixture(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fixtureId, err := strconv.Atoi(id)
	if err != nil {
		internalServerError(w, err)
		return
	}
	fixture, err := h.dao.GetFixture(fixtureId)
	if err != nil {
		internalServerError(w, err)
		return
	}

	ok(w, GetFixtureResponse{Fixture: fixture})
}

func (h *FixtureHandler) GetFixtures(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	json := queryParams.Get("json")

	request, err := unmarshal[GetFixturesRequest](json)
	if err != nil {
		internalServerError(w, err)
		return
	}
	if request.LeagueId == 0 && request.TeamId == 0 && !request.TodayOnly {
		badRequest(w, "LeagueId or TeamId query param must be set if todayOnly is not true.")
		return
	}

	var filter db.GetFixturesFilter
	filter.LeagueId = request.LeagueId
	filter.TeamId = request.TeamId
	if request.TodayOnly {
		filter.Date = time.Now()
		supportedLeagueIds := []int{1, 2, 3, 39, 45, 48}
		filter.LeagueIds = supportedLeagueIds
	}
	fixtures, err := h.dao.GetFixtures(filter)
	if err != nil {
		internalServerError(w, err)
		return
	}

	ok(w, GetFixturesResponse{Fixtures: fixtures})
}
