package rule34

import (
	"encoding/json"
	"fmt"
	"kannonfoundry/whutbot3/api"
	"kannonfoundry/whutbot3/config"
	"kannonfoundry/whutbot3/db"
	"net/http"
	"strings"
)

type R34Posts []R34Post
type R34Post struct {
	ID       int64  `json:"id"`
	Tags     string `json:"tags"`
	FileURL  string `json:"file_url"`
	Hash     string `json:"hash"`
	FileName string `json:"image"`
}

var (
	baseUrl = "https://api.rule34.xxx/index.php?json=1&page=dapi&s=post&q=index"
)

func getSearchUrl(tags []string) string {
	cfg := config.Default()
	return fmt.Sprintf("%s&tags=%s&user_id=%s&api_key=%s", baseUrl, strings.Join(tags, "+"), cfg.R34UserID, cfg.R34ApiKey)
}

func NewClient() *R34MediaSearcher {
	return &R34MediaSearcher{}
}
type R34MediaSearcher struct{}

func (s *R34MediaSearcher) Search(tags []string) (file []api.FileToSend, err error) {
	posts, err := GetPosts(tags)
	if err != nil {
		return []api.FileToSend{}, err
	}
	if len(posts) == 0 {
		return []api.FileToSend{}, fmt.Errorf("no posts found")
	}
	var results = []api.FileToSend{}
	for _, post := range posts {
		results = append(results, api.FileToSend{
			Name: post.FileName,
			URL:  post.FileURL,
		})
	}
	return results, nil
}

func (s *R34MediaSearcher) FormatAndModifySearch(tags []string, authorID int64) (searchTerm string, err error) {
	prefs, err := db.GetPreferences(authorID)
	if err != nil {
		return "", err
	}
	searchTerm = strings.Join(tags, " ")
	for _, pref := range prefs {
		searchTerm += " " + pref.Preference
	}
	return searchTerm, nil
}

func GetPosts(tags []string) (R34Posts, error) {
	// Implementation for fetching posts from the Rule34 API

	endpoint := getSearchUrl(tags)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return R34Posts{}, err
	}
	req.Close = true

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return R34Posts{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return R34Posts{}, fmt.Errorf("failed to fetch posts: %s", resp.Status)
	}

	// Parse the response body
	var data R34Posts
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("error decoding response body: %v", err)
		return R34Posts{}, err
	}

	return data, nil
}
