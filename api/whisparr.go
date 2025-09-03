package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func lookupSceneUrl() string {
	return os.Getenv("WHISPAR_DOMAIN") + "/api/v3/lookup/scene"
}
func addSceneUrl() string {
	return os.Getenv("WHISPAR_DOMAIN") + "/api/v3/movie"
}
func createHeaders() http.Header {
	headers := http.Header{}
	apiKey := os.Getenv("WHISPAR_API_KEY")
	if apiKey == "" {
		return nil
	}
	headers.Add("X-Api-Key", apiKey)
	headers.Add("Accept", "*/*")
	headers.Add("Connection", "keep-alive")
	return headers
}

// response model minimal for the lookup API when the root is an array
type lookupItem struct {
	Movie struct {
		Title       string `json:"title"`
		StudioTitle string `json:"studioTitle"`
		ForeignId   string `json:"foreignId"`
		Year        int    `json:"year"`
		Added       string `json:"added"`
	} `json:"movie"`
}

func LookupScene(stashId string) (bool, error) {
	if stashId == "" {
		return false, nil
	}
	endpoint := lookupSceneUrl()
	apiKey := os.Getenv("WHISPAR_API_KEY")
	if apiKey == "" {
		return false, errors.New("WHISPAR_API_KEY not set")
	}

	headers := createHeaders()
	if headers == nil {
		return false, errors.New("failed to create headers")
	}

	// Build URL with ?term=<escaped>
	reqURL := endpoint + "?term=" + url.QueryEscape(stashId)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return false, err
	}
	req.Header = headers
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// not found -> scene does not exist
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to lookup scene: status %d", resp.StatusCode)
	}

	var body []lookupItem
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&body); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(body) == 0 {
		return false, nil
	}
	added := strings.TrimSpace(body[0].Movie.Added)
	if added == "" {
		return false, nil
	}
	// sentinel zero time in many APIs
	if strings.HasPrefix(added, "0001-01-01") {
		return false, nil
	}

	return true, nil
}

func AddScene(stashId string) (bool, error) {
	if stashId == "" {
		return false, errors.New("empty scene id")
	}

	headers := createHeaders()
	if headers == nil {
		return false, errors.New("failed to create headers")
	}

	// Re-run lookup to fetch movie data for construction of the add payload
	lookupURL := lookupSceneUrl() + "?term=" + url.QueryEscape(stashId)
	req, err := http.NewRequest("GET", lookupURL, nil)
	if err != nil {
		return false, err
	}
	req.Header = headers
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, errors.New("movie not found for add")
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("lookup failed: status %d: %s", resp.StatusCode, string(b))
	}

	var body []lookupItem
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false, fmt.Errorf("failed to decode lookup response: %w", err)
	}
	if len(body) == 0 {
		return false, errors.New("no movie data from lookup")
	}

	movie := body[0].Movie

	// Construct payload for add. Adjust field names to match Whispar's API expectations.
	rootPath := os.Getenv("ROOT_FOLDER")
	qpEnv := os.Getenv("QUALITY")
	var qpVal interface{} = nil
	if qpEnv != "" {
		if id, err := strconv.Atoi(qpEnv); err == nil {
			qpVal = id
		} else {
			// ignore invalid quality profile id, proceed without it
		}
	}

	payload := map[string]interface{}{
		"title":          movie.Title,
		"studioTitle":    movie.StudioTitle,
		"foreignId":      movie.ForeignId,
		"year":           movie.Year,
		"rootFolderPath": rootPath,
		"monitored":      true,
		"addOptions":     map[string]bool{"searchForMovie": true},
	}
	if qpVal != nil {
		payload["qualityProfileId"] = qpVal
	}
	js, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	addReq, err := http.NewRequest("POST", addSceneUrl(), bytes.NewReader(js))
	if err != nil {
		return false, err
	}
	addReq.Header = headers
	addReq.Header.Set("Content-Type", "application/json")

	addResp, err := client.Do(addReq)
	if err != nil {
		return false, err
	}
	defer addResp.Body.Close()

	if addResp.StatusCode < 200 || addResp.StatusCode >= 300 {
		b, _ := io.ReadAll(addResp.Body)
		return false, fmt.Errorf("add failed: status %d: %s", addResp.StatusCode, string(b))
	}

	return true, nil
}
