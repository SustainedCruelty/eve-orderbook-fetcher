package esi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ESITokens struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    uint   `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

// refreshes our access token
// using the refresh token from the configuration
func (f *Fetcher) RefreshToken() (*ESITokens, error) {
	form := url.Values{
		"grant_type":    []string{"refresh_token"},
		"refresh_token": []string{f.config.RefreshToken},
		"client_id":     []string{f.config.ClientID},
	}

	request, err := http.NewRequest(http.MethodPost, "https://login.eveonline.com/v2/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Host", "login.eveonline.com")

	response, err := f.client.Do(request)
	if err != nil {
		return nil, err
	} else if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned status: %s", response.Status)
	}
	defer response.Body.Close()

	var tokens *ESITokens
	if err = json.NewDecoder(response.Body).Decode(&tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}
