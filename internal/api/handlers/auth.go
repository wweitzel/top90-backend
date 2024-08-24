package handlers

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/wweitzel/top90/internal/auth"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
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
	token, err := auth.CreateToken(admin)
	if err != nil {
		internalServerError(w, err)
		return
	}

	useScureCookie, err := strconv.ParseBool(os.Getenv("TOP90_USE_SECURE_COOKIE"))
	if err != nil {
		internalServerError(w, errors.New("invalid value for TOP90_USE_SECURE_COOKIE env variable"))
		return
	}

	domain := os.Getenv("TOP90_COOKIE_DOMAIN")
	if domain == "" {
		internalServerError(w, errors.New("invalid value for TOP90_COOKIE_DOMAIN env variable"))
		return
	}

	authCookie, err := auth.SignCookie(http.Cookie{
		Name:     "top90-auth-token",
		Path:     "/",
		Value:    token,
		MaxAge:   3600,
		Domain:   domain,
		HttpOnly: true,
		Secure:   useScureCookie,
		SameSite: http.SameSiteLaxMode,
	})
	if err != nil {
		internalServerError(w, err)
		return
	}

	loginCookie, err := auth.SignCookie(http.Cookie{
		Name:     "top90-logged-in",
		Path:     "/",
		Value:    "true",
		MaxAge:   3600,
		HttpOnly: false,
		Domain:   domain,
		Secure:   useScureCookie,
		SameSite: http.SameSiteLaxMode,
	})
	if err != nil {
		internalServerError(w, err)
		return
	}

	http.SetCookie(w, &authCookie)
	http.SetCookie(w, &loginCookie)
	ok(w, LoginResponse{Message: "Success"})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	useScureCookie, err := strconv.ParseBool(os.Getenv("TOP90_USE_SECURE_COOKIE"))
	if err != nil {
		internalServerError(w, errors.New("invalid value for TOP90_USE_SECURE_COOKIE env variable"))
		return
	}

	domain := os.Getenv("TOP90_COOKIE_DOMAIN")
	if domain == "" {
		internalServerError(w, errors.New("invalid value for TOP90_COOKIE_DOMAIN env variable"))
		return
	}

	authCookie, err := auth.SignCookie(http.Cookie{
		Name:     "top90-auth-token",
		Path:     "/",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Domain:   domain,
		Secure:   useScureCookie,
		SameSite: http.SameSiteLaxMode,
	})
	if err != nil {
		internalServerError(w, err)
		return
	}

	loginCookie, err := auth.SignCookie(http.Cookie{
		Name:     "top90-logged-in",
		Path:     "/",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: false,
		Domain:   domain,
		Secure:   useScureCookie,
		SameSite: http.SameSiteLaxMode,
	})
	if err != nil {
		internalServerError(w, err)
		return
	}

	http.SetCookie(w, &authCookie)
	http.SetCookie(w, &loginCookie)
	ok(w, LogoutResponse{Message: "Success"})
}
