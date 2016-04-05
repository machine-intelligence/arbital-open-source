// Talk to MailChimp API
package mailchimp

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

// Convert an email to a subscriber hash, according to:
// http://developer.mailchimp.com/documentation/mailchimp/guides/manage-subscribers-with-the-mailchimp-api/
func getSubscriberHash(email string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(strings.ToLower(email))))
}

// SubscribeUser subscribes the given account to the mailing list.
func SubscribeUser(c sessions.Context, account *Account) error {
	account.Status = "subscribed"
	jsonData, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("Couldn't marshal json: %v", err)
	}
	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/members/%s", getListUrl(), getSubscriberHash(account.Email)), bytes.NewBuffer(jsonData))
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

// Check which interests the given email is subscribed to.
func GetInterests(c sessions.Context, email string) (map[string]bool, error) {
	//'https://usX.api.mailchimp.com/3.0/lists/57afe96172/members/852aaa9532cb36adfb5e9fef7a4206a9'
	interestMap := make(map[string]bool)
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/members/%s", getListUrl(), getSubscriberHash(email)), bytes.NewBuffer([]byte("")))
	if err != nil {
		return interestMap, fmt.Errorf("Couldn't create request: %v", err)
	}
	request.SetBasicAuth("", config.XC.Mailchimp.Apikey)
	request.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := sendRequest(c, request)
	if err != nil {
		// TODO: technically we need to look at the error to determine if something bad happened,
		// or if the user is simply not in MC database. For now we just assume the latter.
		return interestMap, nil
	}

	// Decode body
	var result map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	if err != nil {
		return interestMap, fmt.Errorf("Mailchimp returned '%s', but couldn't decode json: %v", resp.Status, err)
	}

	interestMapData := result["interests"].(map[string]interface{})
	for k, v := range interestMapData {
		interestMap[k] = v.(bool)
	}
	return interestMap, nil
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
