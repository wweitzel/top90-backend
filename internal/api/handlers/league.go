package handlers

import (
	"net/http"

	"github.com/wweitzel/top90/internal/db/dao"
	db "github.com/wweitzel/top90/internal/db/models"
)

type GetLeaguesResponse struct {
	Leagues []db.League `json:"leagues"`
}

type LeagueHandler struct {
	dao dao.Top90DAO
}

func NewLeagueHandler(dao dao.Top90DAO) *LeagueHandler {
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
