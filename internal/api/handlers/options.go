package handlers

import "net/http"

type OptionsHandler struct{}

func (s *OptionsHandler) Default(w http.ResponseWriter, r *http.Request) {
	ok(w, nil)
}
