// Talk to MailChimp API
package mailchimp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"appengine/urlfetch"

	"zanaduu3/src/config"
	"zanaduu3/src/sessions"
)

const (
	listId = "c2d8c6a708"
)

// Holds information about a MailChimp account
type Account struct {
	Email     string          `json:"email_address"`
	Status    string          `json:"status"`
	Interests map[string]bool `json:"interests"`
}

func getUrl() string {
	return config.XC.Mailchimp.Production
}

func getListUrl() string {
	return fmt.Sprintf("%s/lists/%s", getUrl(), listId)
}

// SubscribeUser subscribes the given account to the mailing list.
func SubscribeUser(c sessions.Context, account *Account) error {
	account.Status = "subscribed"
	jsonData, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("Couldn't marshal json: %v", err)
	}
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/members", getListUrl()), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Couldn't create request: %v", err)
	}
	request.SetBasicAuth("", config.XC.Mailchimp.Apikey)
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
			return nil, fmt.Errorf("Mailchimp returned '%s', but couldn't decode json: %v", resp.Status, err)
		}
		return nil, fmt.Errorf("Mailchimp returned '%s': %+v", resp.Status, result)
	}
	return resp, nil
}

func init() {
}
