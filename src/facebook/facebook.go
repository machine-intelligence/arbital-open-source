// Talk to Facebook API
package facebook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"appengine/urlfetch"

	"zanaduu3/src/config"
	"zanaduu3/src/sessions"
)

// Holds information about a Stormpath account
type Account struct {
	GivenName string `json:"givenName"`
	Email     string `json:"email"`
	Surname   string `json:"surname"`
}

// ProcessCodeToken takes a code token returned from OAuth, and returns auth_token
func ProcessCodeToken(c sessions.Context, token, redirectUrl string) (string, error) {
	fbId := config.XC.Facebook.ID
	fbSecret := config.XC.Facebook.Secret
	if !sessions.Live {
		fbId = config.XC.Facebook.TestId
		fbSecret = config.XC.Facebook.TestSecret
	}
	tokenUrl := fmt.Sprintf("https://graph.facebook.com/v2.5/oauth/access_token?client_id=%s&client_secret=%s&redirect_uri=%s&code=%s", fbId, fbSecret, url.QueryEscape(redirectUrl), token)
	request, err := http.NewRequest("GET", tokenUrl, bytes.NewBuffer([]byte("")))
	if err != nil {
		return "", fmt.Errorf("Couldn't create request: %v", err)
	}

	// Execute request
	resp, err := sendRequest(c, request)
	if err != nil {
		return "", fmt.Errorf("Couldn't execute request: %v", err)
	}

	decoder := json.NewDecoder(resp.Body)
	var result map[string]interface{}
	err = decoder.Decode(&result)
	if err != nil {
		return "", fmt.Errorf("Couldn't decode json: %v", err)
	}

	accessToken := result["access_token"].(string)
	return accessToken, nil
}

// ProcessAccessToken takes an access token and returns the corresponding account info
func ProcessAccessToken(c sessions.Context, token string) (string, error) {
	fbId := config.XC.Facebook.ID
	fbSecret := config.XC.Facebook.Secret
	if !sessions.Live {
		fbId = config.XC.Facebook.TestId
		fbSecret = config.XC.Facebook.TestSecret
	}
	tokenUrl := fmt.Sprintf("https://graph.facebook.com/debug_token?input_token=%s&access_token=%s|%s", token, fbId, fbSecret)
	request, err := http.NewRequest("GET", tokenUrl, bytes.NewBuffer([]byte("")))
	if err != nil {
		return "", fmt.Errorf("Couldn't create request: %v", err)
	}

	// Execute request
	resp, err := sendRequest(c, request)
	if err != nil {
		return "", fmt.Errorf("Couldn't execute request: %v", err)
	}

	decoder := json.NewDecoder(resp.Body)
	var result map[string]map[string]interface{}
	err = decoder.Decode(&result)
	if err != nil {
		return "", fmt.Errorf("Couldn't decode json: %v", err)
	}

	userId := result["data"]["user_id"].(string)
	return userId, nil
}

// sendRequest sends the given request object to the Stormpath server.
func sendRequest(c sessions.Context, request *http.Request) (*http.Response, error) {
	transport := &urlfetch.Transport{Context: c, AllowInvalidServerCertificate: true}
	resp, err := transport.RoundTrip(request)
	if err != nil {
		return nil, fmt.Errorf("Round trip failed: %v", err)
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		// Process an error
		decoder := json.NewDecoder(resp.Body)
		var result map[string]interface{}
		err = decoder.Decode(&result)
		if err != nil {
			return nil, fmt.Errorf("Facebook returned '%s', but couldn't decode json: %v", resp.Status, err)
		}
		return nil, fmt.Errorf("Facebook returned '%s': %+v", resp.Status, result)
	}
	return resp, nil
}

func init() {
}
