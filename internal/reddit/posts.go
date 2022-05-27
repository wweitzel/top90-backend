package reddit

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type RedditPost struct {
	Kind string
	Data struct {
		URL         string
		Title       string
		Created_utc float64
		Id          string
	}
}

type RedditPostsResponse struct {
	Kind string
	Data struct {
		Dist     int
		After    string
		Children []RedditPost
		Before   interface{}
	}
}

const newPostsBaseUrl = `https://api.reddit.com/r/soccer/search/?q=flair%3Amedia&include_over_18=on&restrict_sr=on&sort=new&limit=100`

// after is a unix epoch to get posts after
func (client *RedditClient) GetNewPosts(after int64) []RedditPost {
	postsResponse := client.getPosts(newPostsBaseUrl)

	newPosts := getUnprocessedPosts(postsResponse, after)
	if len(newPosts) == 0 {
		log.Println("No unprocessed posts found.")
		return newPosts
	}

	printDataUrls("New links:", newPosts)

	var supportedPosts = getSupportedPosts(newPosts)
	printDataUrls("Supported links:", supportedPosts)

	return supportedPosts
}

// get all posts for a search term
func (client *RedditClient) GetPosts(searchTerm string) []RedditPost {
	posts := client.getAllPosts(searchTerm)
	log.Println(posts)

	var supportedPosts = getSupportedPosts(posts)
	printDataUrls("Supported links:", posts)
	return supportedPosts
}

func (client *RedditClient) getPosts(url string) RedditPostsResponse {
	fmt.Println(url)
	body, err := client.doGet(url)
	if err != nil {
		log.Println(err)
	}

	var postsResponse RedditPostsResponse
	err = json.Unmarshal(body, &postsResponse)
	if err != nil {
		log.Println(err)
	}

	return postsResponse
}

func (client *RedditClient) getAllPosts(searchTerm string) []RedditPost {
	var posts []RedditPost

	var redditPostsResponse RedditPostsResponse
	redditPostsResponse.Data.After = "start"
	for redditPostsResponse.Data.After != "" {
		apiUrl := buildSearchUrl(searchTerm, redditPostsResponse.Data.After)
		redditPostsResponse = client.getPosts(apiUrl)
		posts = append(posts, redditPostsResponse.Data.Children...)
	}

	return posts
}

func buildSearchUrl(searchTerm string, after string) string {
	return `https://api.reddit.com/r/soccer/search/?q=flair%3Amedia+` + searchTerm +
		`&include_over_18=on&restrict_sr=on&sort=new&limit=100&after=` + after
}

func getUnprocessedPosts(postsResponse RedditPostsResponse, lastProcessedEpoch int64) []RedditPost {
	currentRedditVideoEpoch := int64(postsResponse.Data.Children[0].Data.Created_utc)

	i := 0
	for currentRedditVideoEpoch > lastProcessedEpoch && i < len(postsResponse.Data.Children) {
		currentRedditVideoEpoch = int64(postsResponse.Data.Children[i].Data.Created_utc)
		i++
	}

	if i != 0 {
		i = i - 1
	}

	return postsResponse.Data.Children[:i]
}

// Returns array with posts excluding unsupported domains
func getSupportedPosts(posts []RedditPost) []RedditPost {
	var filteredPosts []RedditPost
	for _, post := range posts {
		shouldAdd := true
		unsupportedDomains := getUnsupportedDomains()
		for _, unsupportedDomain := range unsupportedDomains {
			if strings.Contains(post.Data.URL, unsupportedDomain) {
				shouldAdd = false
				break
			}
		}
		if shouldAdd {
			filteredPosts = append(filteredPosts, post)
		}
	}
	return filteredPosts
}

func getUnsupportedDomains() []string {
	return []string{
		"v.redd.it",
		"i.redd.it",
		"youtu.be",
		"youtube.com",
		"twitter.com",
		"goalstube.online",
		"i.imgur.com",
		"reddit.com",
		"stattosoftware.com",
		"espn.com",
		"imgur.com",
		"skysports.com",   // https://www.skysports.com/watch/video/sports/12579696/ralf-rangnick-harry-maguire-wont-be-booed-at-old-traffor
		"telegraaf.nl",    // https://www.telegraaf.nl/sport/1454551702/manchester-united-kan-ten-hag-voor-ruim-twee-miljoen-euro-oppikken-bij-ajax
		"theguardian.com", // https://www.theguardian.com/football/ng-interactive/2022/apr/05/david-squires-on-christian-eriksen-and-the-comeback-story-of-the-season?CMP=Share_iOSApp_Other
	}
}

func printDataUrls(caption string, posts []RedditPost) {
	log.Println(caption)
	log.Println("[")
	for _, post := range posts {
		log.Println("\t" + post.Data.URL + ",")
	}
	log.Println("]")
}
