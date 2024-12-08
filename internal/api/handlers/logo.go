package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type LogoHandler struct{}

const (
	maxRetries        = 5
	initialBackoff    = 1 * time.Second
	backoffMultiplier = 2
)

func NewLogoHandler() *LogoHandler {
	return &LogoHandler{}
}

func (h *LogoHandler) GetLogo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	url := "https://media.api-sports.io/football/teams/" + id + ".png"
	img, contentType, err := fetchImageWithBackoff(url)
	if err != nil {
		internalServerError(w, fmt.Errorf("failed to fetch image: %v", err))
		return
	}

	okImage(w, contentType, img)
}

func fetchImageWithBackoff(url string) ([]byte, string, error) {
	var backoff = initialBackoff

	for i := 0; i <= maxRetries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			return nil, "", fmt.Errorf("error making request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			if i == maxRetries {
				return nil, "", fmt.Errorf("reached maximum retries after HTTP 429")
			}
			time.Sleep(backoff)
			backoff *= backoffMultiplier
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("unexpected response status: %d", resp.StatusCode)
		}

		imageData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", fmt.Errorf("error reading response body: %v", err)
		}

		return imageData, resp.Header.Get("Content-Type"), nil
	}

	return nil, "", fmt.Errorf("exceeded retry limit without success")
}
