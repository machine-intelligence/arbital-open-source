// Talk to Okta API
package okta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/appengine/urlfetch"

	"zanaduu3/src/sessions"
)

// Holds information about a Okta account
type Account struct {
	GivenName string `json:"givenName"`
	Email     string `json:"email"`
	Surname   string `json:"surname"`
}

func getOktaURL() string {
	if sessions.Live {
		//return config.XC.Okta.Production
		return "https://dev-552052.oktapreview.com"
	}
	return "https://dev-552052.oktapreview.com"
	//return config.XC.Okta.Stage
}

func CreateNewUser(c sessions.Context, givenName, surname, email, password string) error {
	jsonStr := fmt.Sprintf(`{
		"profile": {
			"firstName": "%s",
			"lastName": "%s",
			"email": "%s",
			"login": "%s"
		},
		"credentials": {
			"password": { "value": "%s" }
		}
	}`, givenName, surname, email, email, password)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/users?activate=true", getOktaURL()), bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return fmt.Errorf("Couldn't create request: %v", err)
	}
	//request.SetBasicAuth(config.XC.Okta.ID, config.XC.Okta.Secret)
	request.Header.Set("Authorization", "SSWS "+`00JBBfWG_O--Kg8tZDOF-5PIAjTWJluCKQ1M7WPYLl`)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	// Execute request
	_, err = sendRequest(c, request)
	if err != nil {
		return fmt.Errorf("Couldn't execute request: %v", err)
	}
	return nil
}

func CreateNewFbUser(c sessions.Context, accessToken string) (*Account, error) {
	jsonStr := fmt.Sprintf(`{
		"providerData": {
			"providerId": "facebook",
			"accessToken": "%s"
		}
	}`, accessToken)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/accounts", getOktaURL()), bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return nil, fmt.Errorf("Couldn't create request: %v", err)
	}
	//request.SetBasicAuth(config.XC.Okta.ID, config.XC.Okta.Secret)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := sendRequest(c, request)
	if err != nil {
		return nil, fmt.Errorf("Couldn't execute request: %v", err)
	}

	decoder := json.NewDecoder(resp.Body)
	var result Account
	err = decoder.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("Couldn't decode json: %v", err)
	}

	return &result, nil
}

func AuthenticateUser(c sessions.Context, email, password string) error {
	jsonStr := fmt.Sprintf(`{
		"username": "%s",
		"password": "%s"
	}`, email, password)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/authn", getOktaURL()), bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return fmt.Errorf("Couldn't create request: %v", err)
	}
	//request.SetBasicAuth(config.XC.Okta.ID, config.XC.Okta.Secret)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	// Execute request
	_, err = sendRequest(c, request)
	if err != nil {
		return fmt.Errorf("Couldn't execute request: %v", err)
	}
	return nil
}

func ForgotPassword(c sessions.Context, email string) error {
	jsonStr := fmt.Sprintf(`{
		"email": "%s"
	}`, email)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/passwordResetTokens", getOktaURL()), bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return fmt.Errorf("Couldn't create request: %v", err)
	}
	//request.SetBasicAuth(config.XC.Okta.ID, config.XC.Okta.Secret)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	// Execute request
	_, err = sendRequest(c, request)
	if err != nil {
		return fmt.Errorf("Couldn't execute request: %v", err)
	}
	return nil
}

func VerifyEmail(c sessions.Context, spToken string) error {
	request, err := http.NewRequest("POST", fmt.Sprintf("https://api.stormpath.com/v1/accounts/emailVerificationTokens/%s", spToken), bytes.NewBuffer([]byte("")))
	if err != nil {
		return fmt.Errorf("Couldn't create request: %v", err)
	}
	//request.SetBasicAuth(config.XC.Okta.ID, config.XC.Okta.Secret)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	// Execute request
	_, err = sendRequest(c, request)
	if err != nil {
		return fmt.Errorf("Couldn't execute request: %v", err)
	}
	return nil
}

// sendRequest sends the given request object to the Okta server.
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
			return nil, fmt.Errorf("Okta returned '%s', but couldn't decode json: %v", resp.Status, err)
		}
		return nil, fmt.Errorf("Okta returned '%s': %+v", resp.Status, result)
	}
	return resp, nil
}

func init() {
}
