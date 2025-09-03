package redgifsapi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type RedGifsClient struct {
	authToken   string
	tokenExpiry int64
}
type loginResponse struct {
	token string `json:"token"`
}

func NewClient() *RedGifsClient {
	return &RedGifsClient{}
}

func login(client *RedGifsClient) error {
	client.authToken = "" // reset any existing token
	resp, err := http.Get("https://api.redgifs.com/v2/auth/temporary")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to login: %s", resp.Status)
	}

	var loginResp loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return err
	}

	client.authToken = loginResp.token

	// Decode JWT token to extract exp
	parts := strings.Split(loginResp.token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT token format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("failed to decode JWT payload: %w", err)
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return fmt.Errorf("failed to parse JWT claims: %w", err)
	}
	client.tokenExpiry = claims.Exp

	return nil
}

func (c *RedGifsClient) IsTokenExpired() bool {
	if c.tokenExpiry == 0 {
		return true // treat unset expiry as expired
	}
	return time.Now().Unix() >= c.tokenExpiry
}

func request() {}
