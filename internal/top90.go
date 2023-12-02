package top90

import (
	"time"
)

type Goal struct {
	Id                    string    `json:"id"`
	RedditFullname        string    `json:"redditFullname"`
	RedditLinkUrl         string    `json:"redditLinkUrl"`
	RedditPostTitle       string    `json:"redditPostTitle"`
	RedditPostCreatedAt   time.Time `json:"redditPostCreatedAt"`
	CreatedAt             string    `json:"createdAt"`
	FixtureId             int       `json:"fixtureId"`
	S3ObjectKey           string    `json:"s3ObjectKey"`
	PresignedUrl          string    `json:"presignedUrl"`
	ThumbnailS3Key        string    `json:"thumbnailS3Key"`
	ThumbnailPresignedUrl string    `json:"thumbnailPresignedUrl"`
}
