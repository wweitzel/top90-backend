package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/db"
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

	thumbnailUrl := "https://s3-redditsoccergoals.top90.io/" + goal.ThumbnailS3Key
	videoUrl := "https://s3-redditsoccergoals.top90.io/" + goal.S3ObjectKey

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
			`	<meta property="og:image" content="` + thumbnailUrl + `"` + `>` +
			`	<meta property="og:video" content="` + videoUrl + `"` + `>` +
			` <meta http-equiv="refresh" content="0; url='https://top90.io/goals/` + goal.Id + `'" />` +
			` <title>top90.io</title>` +
			`</head>` +
			`<body>` +
			`</body>` +
			`</html>`

	w.Header().Add("Content-Type", "text/html")
	w.Write([]byte(html))
}
