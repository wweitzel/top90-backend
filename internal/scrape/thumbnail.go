package scrape

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
)

func extractThumbnail(video *os.File) (filename string, err error) {
	randomId := uuid.NewString()
	randomId = strings.Replace(randomId, "-", "", -1)
	thumbnailFilename := fmt.Sprintf("tmp/%s.jpg", randomId)

	ffmpegPath := os.Getenv("TOP90_FFMPEG_PATH")
	cmd := exec.Command(ffmpegPath, "-i", video.Name(), "-q:v", "8", "-vframes", "1", thumbnailFilename)
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	return thumbnailFilename, nil
}
