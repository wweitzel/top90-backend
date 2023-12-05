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
func (c *Client) GetNewPosts() ([]Post, error) {
	url := `https://api.reddit.com/r/soccer/new?include_over_18=on`
	resp, err := c.getPosts(url)
	if err != nil {
		return nil, err
	}

	posts := resp.Data.Children
	if len(posts) == 0 {
		return posts, nil
	}

	newMediaPosts := mediaPosts(posts)
	printUrls("New links:", newMediaPosts)

	posts = supportedPosts(newMediaPosts)
	printUrls("Supported links:", posts)
	return posts, nil
}

// Get all posts for a search term
func (c *Client) GetAllPosts(searchTerm string) ([]Post, error) {
	posts, err := c.getAllPosts(searchTerm)
	if err != nil {
		return nil, err
	}
	printUrls("All links:", posts)

	posts = supportedPosts(posts)
	printUrls("Supported links:", posts)
	return posts, nil
}

func (c *Client) getPosts(url string) (*PostsResponse, error) {
	resp, err := c.doGet(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := &PostsResponse{}
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (c *Client) getAllPosts(searchTerm string) ([]Post, error) {
	var posts []Post

	var resp *PostsResponse
	resp.Data.After = "start"
	for resp.Data.After != "" {
		url := searchUrl(searchTerm, resp.Data.After)
		resp, err := c.getPosts(url)
		if err != nil {
			return nil, err
		}
		posts = append(posts, resp.Data.Children...)
	}
	return posts, nil
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
		supported := true
		for _, unsupportedDomain := range unsupportedDomains {
			if strings.Contains(post.Data.URL, unsupportedDomain) {
				supported = false
				break
			}
		}
		if supported {
			filteredPosts = append(filteredPosts, post)
		}
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

func printUrls(caption string, posts []Post) {
	log.Print(caption)
	log.Print("[")
	for _, post := range posts {
		log.Print("\t" + post.Data.URL + ",")
	}
	log.Print("]")
}
