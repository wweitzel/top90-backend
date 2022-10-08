package db

type TableNames struct {
	Goals   string
	Leagues string
	Teams   string
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

type TeamColumns struct {
	Id       string
	Name     string
	Code     string
	Country  string
	Founded  string
	National string
	Logo     string
}

var tableNames = TableNames{
	Goals:   "goals",
	Leagues: "leagues",
	Teams:   "teams",
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

var teamColumns = TeamColumns{
	Id:       "id",
	Name:     "name",
	Code:     "code",
	Country:  "country",
	Founded:  "founded",
	National: "national",
	Logo:     "logo",
}
