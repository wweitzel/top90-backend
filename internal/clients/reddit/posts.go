package reddit

import (
	"encoding/json"
	"log"
	"strings"
)

type Post struct {
	Kind string
	Data struct {
		URL                  string
		Title                string
		Created_utc          float64
		Id                   string
		Link_flair_css_class string
	}
}

type PostsResponse struct {
	Kind string
	Data struct {
		Dist     int
		After    string
		Children []Post
		Before   interface{}
	}
}

// Get newest posts
func (c *Client) GetNewPosts() []Post {
	url := `https://api.reddit.com/r/soccer/new?include_over_18=on`
	resp := c.getPosts(url)

	posts := resp.Data.Children
	if len(posts) == 0 {
		return posts
	}

	newMediaPosts := mediaPosts(posts)
	logDataUrls("New links:", newMediaPosts)

	posts = supportedPosts(posts)
	logDataUrls("Supported links:", posts)
	return posts
}

// Get all posts for a search term
func (c *Client) GetAllPosts(searchTerm string) []Post {
	posts := c.getAllPosts(searchTerm)
	logDataUrls("All links:", posts)

	posts = supportedPosts(posts)
	logDataUrls("Supported links:", posts)
	return posts
}

func (c *Client) getPosts(url string) *PostsResponse {
	resp, err := c.doGet(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	r := &PostsResponse{}
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		log.Println(err)
	}
	return r
}

func (c *Client) getAllPosts(searchTerm string) []Post {
	var posts []Post

	var resp *PostsResponse
	resp.Data.After = "start"
	for resp.Data.After != "" {
		url := searchUrl(searchTerm, resp.Data.After)
		resp = c.getPosts(url)
		posts = append(posts, resp.Data.Children...)
	}
	return posts
}

func mediaPosts(posts []Post) []Post {
	var mediaPosts []Post
	for _, post := range posts {
		if post.Data.Link_flair_css_class == "media" {
			mediaPosts = append(mediaPosts, post)
		}
	}
	return mediaPosts
}

func supportedPosts(posts []Post) []Post {
	unsupportedDomains := unsupportedDomains()

	var filteredPosts []Post
	for _, post := range posts {
		for _, unsupportedDomain := range unsupportedDomains {
			if strings.Contains(post.Data.URL, unsupportedDomain) {
				break
			}
		}
		filteredPosts = append(filteredPosts, post)
	}
	return filteredPosts
}

func unsupportedDomains() []string {
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
		"skysports.com",
		"telegraaf.nl",
		"theguardian.com",
	}
}

func searchUrl(searchTerm string, after string) string {
	return `https://api.reddit.com/r/soccer/search/?q=flair%3Amedia+` + searchTerm +
		`&include_over_18=on&restrict_sr=on&sort=new&limit=100&after=` + after
}

func logDataUrls(caption string, posts []Post) {
	log.Println(caption)
	log.Println("[")
	for _, post := range posts {
		log.Println("\t" + post.Data.URL + ",")
	}
	log.Println("]")
}
