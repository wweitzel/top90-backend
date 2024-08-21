package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/wweitzel/top90/internal/jwt"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func respond(w http.ResponseWriter, statusCode int, resp any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json, _ := json.Marshal(resp)
	w.Write(json)
}

func ok(w http.ResponseWriter, resp any) {
	respond(w, http.StatusOK, resp)
}

func internalServerError(w http.ResponseWriter, err error) {
	respond(w, http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
}

func badRequest(w http.ResponseWriter, msg string) {
	respond(w, http.StatusBadRequest, ErrorResponse{Message: msg})
}

func unauthorized(w http.ResponseWriter, msg string) {
	respond(w, http.StatusUnauthorized, ErrorResponse{Message: msg})
}

func unmarshal[T any](jsonStr string) (*T, error) {
	decodedJson, err := url.QueryUnescape(jsonStr)
	if err != nil {
		return nil, errors.New("error decoding json")
	}

	out := new(T)
	err = json.Unmarshal([]byte(decodedJson), &out)
	if err != nil {
		return nil, errors.New("error unmarshalling json")
	}
	return out, nil
}

func authorize(r *http.Request) error {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return errors.New("missing authorization header")
	}
	tokenString = tokenString[len("Bearer "):]

	token, err := jwt.ReadToken(tokenString)
	if err != nil {
		return errors.New("invalid token")
	}

	if !token.Admin {
		return errors.New("must be admin")
	}

	return nil
}
