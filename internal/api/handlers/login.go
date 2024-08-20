package handlers

import (
	"net/http"
	"os"

	"github.com/wweitzel/top90/internal/jwt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type LoginHandler struct{}

func NewLoginHandler() *LoginHandler {
	return &LoginHandler{}
}

func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	json := queryParams.Get("json")

	request, err := unmarshal[LoginRequest](json)
	if err != nil {
		internalServerError(w, err)
		return
	}

	username := os.Getenv("TOP90_ADMIN_USERNAME")
	password := os.Getenv("TOP90_ADMIN_PASSWORD")
	if request.Username != username || request.Password != password {
		unauthorized(w, "Incorrect username and password combination")
		return
	}

	admin := true
	token, err := jwt.CreateToken(admin)
	if err != nil {
		internalServerError(w, err)
		return
	}
	ok(w, LoginResponse{Token: token})
}
