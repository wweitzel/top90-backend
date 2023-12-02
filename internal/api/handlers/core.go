package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func respond(w http.ResponseWriter, statusCode int, response any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json, _ := json.Marshal(response)
	w.Write(json)
}

func unmarshal[T any](jsonStr string) (*T, error) {
	out := new(T)

	decodedJson, err := url.QueryUnescape(jsonStr)
	if err != nil {
		return nil, errors.New("error decoding json")
	}

	err = json.Unmarshal([]byte(decodedJson), &out)
	if err != nil {
		return nil, errors.New("error unmarshalling json")
	}

	return out, nil
}
