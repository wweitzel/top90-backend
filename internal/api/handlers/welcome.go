package handlers

import "net/http"

type GetWelcomeResponse struct {
	Message string `json:"message"`
}

type WelcomeHandler struct{}

func (s *WelcomeHandler) GetWelcome(w http.ResponseWriter, r *http.Request) {
	var resp GetWelcomeResponse
	resp.Message = "Welcome to the top90 API ⚽️ 🥅 "
	ok(w, resp)
}
