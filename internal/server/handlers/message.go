package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/s3"
)

type MessageHandler struct {
	dao      db.Top90DAO
	s3Client s3.S3Client
	s3Bucket string
}

func NewMessageHandler(dao db.Top90DAO, s3Client s3.S3Client, s3Bucket string) *MessageHandler {
	return &MessageHandler{
		dao:      dao,
		s3Client: s3Client,
		s3Bucket: s3Bucket,
	}
}
func (h *MessageHandler) GetPreview(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	goal, _ := h.dao.GetGoal(id)

	// TODO: This is nearly duplicated with the code in getGoals and getGoal
	videoUrl, err := h.s3Client.NewSignedGetURL(goal.S3ObjectKey, h.s3Bucket, time.Hour*24*7)
	if err != nil {
		log.Println(err)
	}
	thumbnailUrl, err := h.s3Client.NewSignedGetURL(goal.ThumbnailS3Key, h.s3Bucket, time.Hour*24*7)
	if err != nil {
		log.Println(err)
	}

	goal.PresignedUrl = videoUrl
	goal.ThumbnailPresignedUrl = thumbnailUrl

	html :=
		`<!doctype html>` +
			`<html lang="en">` +
			`<head>` +
			` <meta charset="utf-8">` +
			// TODO: If possible, better to return the video and image data directly (in base64)
			// rather than returning the presigned url since we don't really want these to expire.
			// Max expiry date for presigned url is apparently one week.
			`	<meta property="og:title" content="` + goal.RedditPostTitle + `"` + `>` +
			`	<meta property="og:image" content="` + goal.ThumbnailPresignedUrl + `"` + `>` +
			`	<meta property="og:video" content="` + goal.PresignedUrl + `"` + `>` +
			` <meta http-equiv="refresh" content="0; url='https://top90.io/goals/` + goal.Id + `'" />` +
			` <title>top90.io</title>` +
			`</head>` +
			`<body>` +
			`</body>` +
			`</html>`

	w.Header().Add("Content-Type", "text/html")
	w.Write([]byte(html))
}
