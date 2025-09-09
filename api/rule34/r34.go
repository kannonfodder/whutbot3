package rule34

import (
	"encoding/json"
	"fmt"
	"kannonfoundry/whutbot3/config"
	"net/http"
	"strings"
)

type R34Posts []R34Post
type R34Post struct {
	ID      int64  `json:"id"`
	Tags    string `json:"tags"`
	FileURL string `json:"file_url"`
}

var (
	baseUrl = "https://api.rule34.xxx/index.php?json=1&page=dapi&s=post&q=index"
)

func getSearchUrl(tags []string) string {
	cfg := config.Default()
	return fmt.Sprintf("%s&tags=%s&user_id=%s&api_key=%s", baseUrl, strings.Join(tags, ","), cfg.R34UserID, cfg.R34ApiKey)
}

func GetPosts(tags []string) (R34Posts, error) {
	// Implementation for fetching posts from the Rule34 API

	endpoint := getSearchUrl(tags)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return R34Posts{}, err
	}

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
