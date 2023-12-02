package handlers

import "net/http"

type GetWelcomeResponse struct {
	Message string `json:"message"`
}

type WelcomeHandler struct{}

func (s *WelcomeHandler) GetWelcome(w http.ResponseWriter, r *http.Request) {
	var apiInfo GetWelcomeResponse
	apiInfo.Message = "Welcome to the top90 API ⚽️ 🥅 "
	respond(w, http.StatusOK, apiInfo)
}
