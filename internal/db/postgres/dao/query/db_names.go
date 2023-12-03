package query

type TableNames struct {
	Goals    string
	Leagues  string
	Teams    string
	Fixtures string
}

type GoalColumns struct {
	Id                  string
	RedditFullname      string
	RedditLinkUrl       string
	RedditPostTitle     string
	RedditPostCreatedAt string
	S3ObjectKey         string
	CreatedAt           string
	FixtureId           string
	ThumbnailS3Key      string
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

type FixtureColumns struct {
	Id         string
	Referee    string
	Date       string
	HomeTeamId string
	AwayTeamId string
	LeagueId   string
	Season     string
	CreatedAt  string
}

var tableNames = TableNames{
	Goals:    "goals",
	Leagues:  "leagues",
	Teams:    "teams",
	Fixtures: "fixtures",
}

var goalColumns = GoalColumns{
	Id:                  "id",
	RedditFullname:      "reddit_fullname",
	RedditLinkUrl:       "reddit_link_url",
	RedditPostTitle:     "reddit_post_title",
	RedditPostCreatedAt: "reddit_post_created_at",
	S3ObjectKey:         "s3_object_key",
	CreatedAt:           "created_at",
	FixtureId:           "fixture_id",
	ThumbnailS3Key:      "thumbnail_s3_key",
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

var fixtureColumns = FixtureColumns{
	Id:         "id",
	Referee:    "referee",
	Date:       "date",
	HomeTeamId: "home_team_id",
	AwayTeamId: "away_team_id",
	LeagueId:   "league_id",
	Season:     "season",
	CreatedAt:  "created_at",
}
