package handlers

import (
	"net/http"

	"github.com/wweitzel/top90/internal/clients/apifootball"
	"github.com/wweitzel/top90/internal/db"
)

type GetLeaguesResponse struct {
	Leagues []apifootball.League `json:"leagues"`
}

type LeagueHandler struct {
	dao db.Top90DAO
}

func NewLeagueHandler(dao db.Top90DAO) *LeagueHandler {
	return &LeagueHandler{
		dao: dao,
	}
}

func (h *LeagueHandler) GetLeagues(w http.ResponseWriter, r *http.Request) {
	leagues, err := h.dao.GetLeagues()
	if err != nil {
		internalServerError(w, err)
		return
	}

	ok(w, GetLeaguesResponse{Leagues: leagues})
}
