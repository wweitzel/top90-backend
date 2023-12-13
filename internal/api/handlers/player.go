package handlers

import (
	"net/http"

	"github.com/wweitzel/top90/internal/db/dao"
	db "github.com/wweitzel/top90/internal/db/models"
)

type SearchPlayersRequest struct {
	SearchTerm string `json:"searchTerm"`
}

type SearchPlayersResponse struct {
	Players []db.Player `json:"players"`
}

type PlayerHandler struct {
	dao dao.Top90DAO
}

func NewPlayerHandler(dao dao.Top90DAO) PlayerHandler {
	return PlayerHandler{
		dao: dao,
	}
}

func (h PlayerHandler) SearchPlayers(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	json := queryParams.Get("json")

	request, err := unmarshal[SearchPlayersRequest](json)
	if err != nil {
		internalServerError(w, err)
		return
	}

	var players []db.Player
	if request.SearchTerm == "" {
		players, err = h.dao.GetTopScorers()
		if err != nil {
			internalServerError(w, err)
			return
		}
		ok(w, SearchPlayersResponse{Players: players})
		return
	}

	players, err = h.dao.SearchPlayers(request.SearchTerm)
	if err != nil {
		internalServerError(w, err)
		return
	}
	ok(w, SearchPlayersResponse{Players: players})
}
