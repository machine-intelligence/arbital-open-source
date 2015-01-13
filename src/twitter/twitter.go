// twitter.go holds integration with Twitter API
package twitter

import (
	"encoding/json"
	"fmt"
	"net/url"

	"appengine/urlfetch"

	"xelaie/src/go/sessions"

	"github.com/garyburd/go-oauth/oauth"
)

var (
	twitterBaseUrl       = "https://api.twitter.com/1.1"
	twitterStreamBaseUrl = "https://stream.twitter.com/1.1"
	TimeLayout           = "Mon Jan 2 15:04:05 -0700 2006"
)

type RetweetedStatus struct {
	Text string `json:"text"`
}

// StatusUpdate holds information about a new Twitter status.
type StatusUpdate struct {
	Id              int64           `json:"id"`
	User            TwitterUser     `json:"user"`
	Text            string          `json:"text"`
	CreatedAt       string          `json:"created_at"`
	RetweetedStatus RetweetedStatus `json:"retweeted_status"`
}

type Followers struct {
	Users []TwitterUser `json:"users"`
}

type StatusUpdateFunc func(sessions.Context, *StatusUpdate)

func StreamStatuses(c sessions.Context, f StatusUpdateFunc) error {
	c.Debugf("Opening statuses/filter stream..")
	params := url.Values{}
	params.Add("track", "#rt2win")
	ts, err := sessions.OpenHttpStream(
		c,
		&oauthClient,
		&daemonCreds,
		twitterStreamBaseUrl+"/statuses/filter.json",
		params)
	if err != nil {
		return fmt.Errorf("Couldn't open twitter stream: %v", err)
	}
	defer ts.Close()

	// Loop until stream has a permanent error.
	for ts.Err() == nil {
		// TODO: handle when Twitter returns an "errors" json
		var update StatusUpdate
		if err := ts.UnmarshalNext(&update); err != nil {
			return fmt.Errorf("Error unmashalling StatusUpdate: %v", err)
		}
		f(c, &update)
	}
	return ts.Err()
}

// GetFollowers returns an array of users following the given user id.
// Get followers. Note: current API returns most recent first, but this
// can change without notice.
// TODO: add paging if recent-first ever changes or there is a possibility
// of more than 200 users / minute.
func GetFollowers(c sessions.Context, userId int64) (*Followers, error) {
	params := url.Values{}
	params.Add("user_id", fmt.Sprintf("%d", userId))
	params.Add("count", "200")
	url := twitterBaseUrl + "/followers/list.json"
	c.Debugf("Getting followers..")
	resp, err := oauthClient.Get(urlfetch.Client(c), &daemonCreds, url, params)
	if err != nil {
		return nil, fmt.Errorf("Get request failed for followers: %v", err)
	}
	defer resp.Body.Close()
	var followers Followers
	// TODO: handle when Twitter returns an "errors" json
	err = json.NewDecoder(resp.Body).Decode(&followers)
	if err != nil {
		return nil, fmt.Errorf("Error decoding the followers: %v", err)
	}
	return &followers, nil
}

// PLEASE BE CAREFUL NOT TO CALL THIS TOO OFTEN. We don't want to get banned.
// twitterUpdateStatus posts a tweet with the given text. It returns the text
// as it was actually tweeted, since Twitter can change URLs.
func UpdateStatus(c sessions.Context, text string) (string, error) {
	return UpdateStatusWithCreds(c, &daemonCreds, text)
}

// UpdateStatusWithCreds allows us to pass in user's credentials to update their status.
func UpdateStatusWithCreds(c sessions.Context, creds *oauth.Credentials, text string) (string, error) {
	if !sessions.Live {
		c.Warningf("Not running LIVE, so not sending the tweet: %s", text)
		return text, nil
	}

	params := url.Values{}
	params.Add("status", text)
	updateUrl := twitterBaseUrl + "/statuses/update.json"
	resp, err := oauthClient.Post(urlfetch.Client(c), creds, updateUrl, params)
	if err != nil {
		return "", fmt.Errorf("POST to update the status failed: %v", err)
	}

	defer resp.Body.Close()
	var dat map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&dat)
	if err != nil {
		return "", fmt.Errorf("Couldn't parse JSON from the response: %v", err)
	}
	if _, ok := dat["errors"]; ok {
		errors := dat["errors"].([]interface{})
		return "", fmt.Errorf("Error returned by Twitter: %v", errors[0])
	}
	return dat["text"].(string), nil
}
