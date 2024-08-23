package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (t *Token) getToken(URL, clientID, clientSecret string) (string, error) {

	// Grant Type for OAuth 2.0 is in body of request
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequest("POST", URL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.Body == http.NoBody {
		return "", fmt.Errorf("empty body in response: %w", err)
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return "", err
	}

	// Parse the response
	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return "", fmt.Errorf("parsing response failed: %w", err)
	}

	return token.AccessToken, nil
}
