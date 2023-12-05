package scrape

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func download(srcUrl string) (*os.File, error) {
	var videoFile *os.File
	var err error

	if strings.HasSuffix(srcUrl, ".m3u8") {
		videoFile, err = downloadBlob("https://juststream.live/", srcUrl)
	} else {
		videoFile, err = downloadVideo(srcUrl)
	}

	return videoFile, err
}

func downloadBlob(referrer string, url string) (*os.File, error) {
	file, err := os.CreateTemp("tmp", "*.mp4")
	if err != nil {
		return nil, err
	}

	ffmpegPath := os.Getenv("TOP90_FFMPEG_PATH")

	cmd := exec.Command(
		ffmpegPath,
		"-headers", "Referer: "+referrer,
		"-i", url,
		"-y",
		"-c", "copy", file.Name())

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	mp4, err := os.Open(file.Name())
	if err != nil {
		return nil, err
	}

	return mp4, nil
}

func downloadVideo(url string) (*os.File, error) {
	file, err := os.CreateTemp("tmp", "*.mp4")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "jawnt")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		os.Remove(file.Name())
		return file, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, err
	}

	return file, nil
}
