package handlers

import (
	"net/http"

	"github.com/wweitzel/top90/internal/db/dao"
	db "github.com/wweitzel/top90/internal/db/models"
)

type GetTeamsRequest struct {
	LeagueId   int    `json:"leagueId"`
	Season     int    `json:"season"`
	SearchTerm string `json:"searchTerm"`
}

type GetTeamsResponse struct {
	Teams []db.Team `json:"teams"`
}

type TeamHandler struct {
	dao dao.Top90DAO
}

func NewTeamHandler(dao dao.Top90DAO) *TeamHandler {
	return &TeamHandler{
		dao: dao,
	}
}

func (h *TeamHandler) GetTeams(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	json := queryParams.Get("json")

	request, err := unmarshal[GetTeamsRequest](json)
	if err != nil {
		internalServerError(w, err)
		return
	}

	var teams []db.Team
	if request.SearchTerm != "" {
		teams, err = h.dao.GetTeams(db.GetTeamsFilter{SearchTerm: request.SearchTerm})
	} else if request.LeagueId != 0 && request.Season != 0 {
		teams, err = h.dao.GetTeamsForLeagueAndSeason(request.LeagueId, request.Season)
	} else {
		teams, err = h.dao.GetTeams(db.GetTeamsFilter{})
	}

	if err != nil {
		internalServerError(w, err)
		return
	}

	ok(w, GetTeamsResponse{Teams: teams})
}
