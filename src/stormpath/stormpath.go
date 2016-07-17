// Talk to Stormpath API
package stormpath

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

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

func getStormpathUrl() string {
	if sessions.Live {
		return config.XC.Stormpath.Production
	}
	return config.XC.Stormpath.Stage
}

func CreateNewUser(c sessions.Context, givenName, surname, email, password string) error {
	jsonStr := fmt.Sprintf(`{
		"givenName": "%s",
		"surname": "%s",
		"email": "%s",
		"password":"%s"
	}`, givenName, surname, email, password)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/accounts", getStormpathUrl()), bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return fmt.Errorf("Couldn't create request: %v", err)
	}
	request.SetBasicAuth(config.XC.Stormpath.ID, config.XC.Stormpath.Secret)
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
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/accounts", getStormpathUrl()), bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return nil, fmt.Errorf("Couldn't create request: %v", err)
	}
	request.SetBasicAuth(config.XC.Stormpath.ID, config.XC.Stormpath.Secret)
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

/*func GetAccountInfo(c sessions.Context, accountId string) (*Account, error) {
	request, err := http.Get("GET", fmt.Sprintf("https://api.stormpath.com/v1/accounts/%s", accountId), bytes.NewBuffer([]byte("")))
	if err != nil {
		return fmt.Errorf("Couldn't create request: %v", err)
	}
	request.SetBasicAuth(config.XC.Stormpath.Id, config.XC.Stormpath.Secret)

	// Execute request
	resp, err := sendRequest(c, request)
	if err != nil {
		return fmt.Errorf("Couldn't execute request: %v", err)
	}

	return nil
}*/

func AuthenticateUser(c sessions.Context, email, password string) error {
	value := base64.StdEncoding.EncodeToString([]byte((fmt.Sprintf("%s:%s", email, password))))
	jsonStr := fmt.Sprintf(`{
		"type": "basic",
		"value": "%s"
	}`, value)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/loginAttempts", getStormpathUrl()), bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return fmt.Errorf("Couldn't create request: %v", err)
	}
	request.SetBasicAuth(config.XC.Stormpath.ID, config.XC.Stormpath.Secret)
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
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/passwordResetTokens", getStormpathUrl()), bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return fmt.Errorf("Couldn't create request: %v", err)
	}
	request.SetBasicAuth(config.XC.Stormpath.ID, config.XC.Stormpath.Secret)
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
	request.SetBasicAuth(config.XC.Stormpath.ID, config.XC.Stormpath.Secret)
	request.Header.Set("Content-Type", "application/json")

	// Execute request
	_, err = sendRequest(c, request)
	if err != nil {
		return fmt.Errorf("Couldn't execute request: %v", err)
	}
	return nil
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
			return nil, fmt.Errorf("Stormpath returned '%s', but couldn't decode json: %v", resp.Status, err)
		}
		return nil, fmt.Errorf("Stormpath returned '%s': %+v", resp.Status, result)
	}
	return resp, nil
}

func init() {
}
