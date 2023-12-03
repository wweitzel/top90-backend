package scrape

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
)

func extractThumbnail(video *os.File) (filename string) {
	randomId := uuid.NewString()
	randomId = strings.Replace(randomId, "-", "", -1)
	thumbnailFilename := fmt.Sprintf("tmp/%s.avif", randomId)

	ffmpegPath := os.Getenv("TOP90_FFMPEG_PATH")
	cmd := exec.Command(ffmpegPath, "-i", video.Name(), "-q:v", "2", "-vframes", "1", thumbnailFilename)
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		log.Println("warning: error generating thumbnail with ffpmeg", err)
	}

	return thumbnailFilename
}
