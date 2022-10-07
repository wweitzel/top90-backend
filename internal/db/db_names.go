package db

type TableNames struct {
	Goals   string
	Leagues string
}

type GoalColumns struct {
	Id                  string
	RedditFullname      string
	RedditLinkUrl       string
	RedditPostTitle     string
	RedditPostCreatedAt string
	S3ObjectKey         string
	CreatedAt           string
}

type LeagueColumns struct {
	Id   string
	Name string
	Type string
	Logo string
}

var tableNames = TableNames{
	Goals:   "goals",
	Leagues: "leagues",
}

var goalColumns = GoalColumns{
	Id:                  "id",
	RedditFullname:      "reddit_fullname",
	RedditLinkUrl:       "reddit_link_url",
	RedditPostTitle:     "reddit_post_title",
	RedditPostCreatedAt: "reddit_post_created_at",
	S3ObjectKey:         "s3_object_key",
	CreatedAt:           "created_at",
}

var leagueColumns = LeagueColumns{
	Id:   "id",
	Name: "name",
	Type: "type",
	Logo: "logo",
}
