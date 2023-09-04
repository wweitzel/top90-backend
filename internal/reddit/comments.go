package reddit

import (
	"encoding/json"
	"fmt"
	"log"
)

type RedditComment struct {
	Data struct {
		Body    string
		Id      string
		Replies RedditRepliesResponse
	}
}

type RedditCommentsResponse struct {
	Kind string
	Data struct {
		Children []RedditComment
	}
}

type RedditReply struct {
	Data struct {
		Children []string
	}
}

type RedditRepliesResponse struct {
	Data struct {
		Children []RedditReply
	}
}

func (client *RedditClient) GetComments(postId string) []RedditComment {
	url := "https://api.reddit.com/r/soccer/comments/" + postId + "/?sort=old"

	body, err := client.doGet(url)
	if err != nil {
		log.Println(err)
	}

	// Note: arr[0] is actually a RedditPostResponse and arr[1] is a RedditCommentsResponse.
	var commentsResponse []RedditCommentsResponse
	err = json.Unmarshal(body, &commentsResponse)
	if err != nil {
		log.Println(postId, err)
		return []RedditComment{}
	}
	return commentsResponse[1].Data.Children
}

func (client *RedditClient) GetComment(commentId string) RedditComment {
	url := "https://api.reddit.com/r/soccer/api/info?id=" + "t1_" + commentId

	body, err := client.doGet(url)
	if err != nil {
		log.Println(err)
	}

	// Note: arr[0] is actually a RedditPostResponse and arr[1] is a RedditCommentsResponse.
	var commentResponse RedditCommentsResponse
	err = json.Unmarshal(body, &commentResponse)
	if err != nil {
		log.Println(commentId, "HAPPENING IN GET COMMENT!!!!!!!!", err)
		return RedditComment{}
	}
	return commentResponse.Data.Children[0]
}

// TODO: Probably a better way to handle this
func handleUnmarshallError(err error) {
	expectedJsonError := "json: cannot unmarshal string into Go struct field .Data.Children.Data.Replies of type reddit.RedditRepliesResponse"
	if err != nil && fmt.Sprintf("%vw", err) != expectedJsonError {
		log.Println(err)
	}
}
