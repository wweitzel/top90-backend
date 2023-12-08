package db

import (
	"time"
)

type GetGoalsFilter struct {
	SearchTerm string
	StartDate  string
	LeagueId   int
	Season     int
	TeamId     int
	FixtureId  int
}

type Goal struct {
	Id                    string     `json:"id" db:"id"`
	RedditFullname        string     `json:"redditFullname" db:"reddit_fullname"`
	RedditLinkUrl         string     `json:"redditLinkUrl" db:"reddit_link_url"`
	RedditPostTitle       string     `json:"redditPostTitle" db:"reddit_post_title"`
	RedditPostCreatedAt   time.Time  `json:"redditPostCreatedAt" db:"reddit_post_created_at"`
	CreatedAt             string     `json:"createdAt" db:"created_at"`
	FixtureId             NullInt    `json:"fixtureId" db:"fixture_id"`
	S3ObjectKey           string     `json:"s3ObjectKey" db:"s3_object_key"`
	PresignedUrl          string     `json:"presignedUrl" db:"presigned_url"`
	ThumbnailS3Key        NullString `json:"thumbnailS3Key" db:"thumbnail_s3_key"`
	ThumbnailPresignedUrl string     `json:"thumbnailPresignedUrl" db:"thumbnail_presigned_url"`
}
