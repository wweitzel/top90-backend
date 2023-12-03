package scrape

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func download(srcUrl string) *os.File {
	var videoFile *os.File
	if strings.HasSuffix(srcUrl, ".m3u8") {
		videoFile = downloadBlob("https://juststream.live/", srcUrl)
	} else {
		videoFile = downloadVideo(srcUrl)
	}
	return videoFile
}

func downloadBlob(referrer string, url string) *os.File {
	file, err := ioutil.TempFile("tmp", "*.mp4")
	if err != nil {
		log.Fatalln(err)
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
		log.Println(err)
	}

	mp4, err := os.Open(file.Name())
	if err != nil {
		log.Println(err)
	}

	return mp4
}

func downloadVideo(url string) *os.File {
	file, err := ioutil.TempFile("tmp", "*.mp4")
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}

	req.Header.Add("User-Agent", "jawnt")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed GET url: %s due to: %v", url, err)
		return file
	}

	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Println(err)
	}

	return file
}
