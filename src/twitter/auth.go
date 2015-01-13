// twitter_auth.go holds integration with Twitter API
package twitter

import (
	"encoding/json"
	"fmt"
	"net/http"

	"appengine/urlfetch"

	"xelaie/src/go/config"
	"xelaie/src/go/sessions"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/gorilla/mux"
	"github.com/hkjn/pages"
)

var (
	xc          = config.Load()
	oauthClient = oauth.Client{
		TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authenticate",
		TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
		Credentials:                   oauth.Credentials{xc.Twitter.Token, xc.Twitter.Secret},
	}
	daemonCreds = oauth.Credentials{xc.Twitter.Daemon.Token, xc.Twitter.Daemon.Secret}
)

// User holds information about a Twitter user.
// Json names are according to Twitter API specification.
type TwitterUser struct {
	Id             int64  `json:"id"`
	Name           string `json:"name"`
	ScreenName     string `json:"screen_name"`
	FollowersCount int    `json:"followers_count"`
	ProfileURL     string `json:"profile_image_url"`
}

type referralFlash struct {
	Id string
}

// GetAuthUrl initiates 3-legged OAuth by returning a redirect URL.
//
// GetAuthUrl also saves the temporary credentials in the user's session.
func GetAuthUrl(w http.ResponseWriter, r *http.Request) (string, error) {
	callback := sessions.GetDomain() + "/authorize_callback"
	c := sessions.NewContext(r)
	client := urlfetch.Client(c)
	// Request temporary credentials. Real ones will be fetched in the
	// handler for /authorize_callback
	tempCreds, err := oauthClient.RequestTemporaryCredentials(
		client, callback, nil)
	if err != nil {
		c.Inc("temp_credential_request_fail")
		return "", fmt.Errorf("failed to request temporary credentials: %v", err)
	}

	session, err := sessions.GetSession(r)
	if err != nil {
		return "", fmt.Errorf("failed to load session: %v", err)
	}
	c.Debugf("saving temp credentials as flash message: %v", tempCreds)
	session.AddFlash(sessions.Credentials{tempCreds}, "tempcreds")

	// Add referral flash if it exists
	referralId := r.URL.Query().Get("referralId")
	if len(referralId) > 0 {
		c.Debugf("User is being referred via #: %s", referralId)
		session.AddFlash(referralFlash{referralId}, "referral")
	} else {
		referralUserId := r.URL.Query().Get("referralUserId")
		if len(referralUserId) > 0 {
			c.Debugf("User is being referred by user: %s", referralUserId)
			session.AddFlash(referralFlash{referralUserId}, "referralUser")
		}
	}

	err = session.Save(r, w)
	if err != nil {
		c.Inc("temp_credential_save_fail")
		return "", fmt.Errorf("failed to save credentials to session: %v", err)
	}
	return oauthClient.AuthorizationURL(tempCreds, nil), nil
}

// verifyTempCreds returns verified temporary credentials
func verifyTempCreds(w http.ResponseWriter, r *http.Request, token string) (*sessions.Credentials, error) {
	c := sessions.NewContext(r)

	s, err := sessions.GetSession(r)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %v", err)
	}

	credFlashes := s.Flashes("tempcreds")
	// Save session first, to ensure that the flashes we just read are cleared.
	err = s.Save(r, w)
	if err != nil {
		return nil, fmt.Errorf("failed to save session: %v", err)
	}

	var tempCreds *sessions.Credentials
	for _, f := range credFlashes {
		var ok bool
		tempCreds, ok = f.(*sessions.Credentials)
		if ok && token == tempCreds.Token {
			return tempCreds, nil
		}
	}

	if tempCreds == nil {
		c.Inc("bad_temp_creds_missing_creds_fail")
		return nil, fmt.Errorf("no valid temp credentials in flash message (out of %d flashes): %v", len(credFlashes), credFlashes)
	}

	if token != tempCreds.Token {
		c.Inc("bad_temp_creds_token_fail")
		return nil, fmt.Errorf("returned oauth_token and temporary token mismatch; %q != %q\n", token, tempCreds.Token)
	}
	c.Inc("bad_temp_creds_fallthrough_bug_fail")
	return nil, fmt.Errorf("no valid temp credentials, but no specific error - this shouldn't happen")
}

// getReferralId returns referrals that might have been stored in a flash
func getReferralId(w http.ResponseWriter, r *http.Request) (referralId string, referralUserId string) {
	c := sessions.NewContext(r)
	s, err := sessions.GetSession(r)
	if err != nil {
		return
	}

	referralFlashes := s.Flashes("referral")
	referralUserFlashes := s.Flashes("referralUser")
	// Save session first, to ensure that the flashes we just read are cleared.
	err = s.Save(r, w)
	if err != nil {
		return
	}

	for _, f := range referralFlashes {
		referral, ok := f.(*referralFlash)
		if ok {
			referralId = referral.Id
			c.Debugf("Extract referralId from flash: %s", referralId)
			break
		}
	}
	for _, f := range referralUserFlashes {
		referral, ok := f.(*referralFlash)
		if ok {
			referralUserId = referral.Id
			c.Debugf("Extract referralUserId from flash: %s", referralUserId)
			break
		}
	}
	return
}

// AuthHandler handles the OAuth callback after receiving credentials.
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	v := mux.Vars(r)
	token := v["token"]
	verifier := v["verifier"]
	c.Debugf("/authorize_callback got oauth_token=%q, oauth_verifier=%q\n", token, verifier)

	tempCreds, err := verifyTempCreds(w, r, token)
	if err != nil {
		pages.ShowError(w, r, fmt.Errorf("failed to verify temp credentials: %v", err))
		return
	}
	referralId, referralUserId := getReferralId(w, r)
	c.Debugf("we have valid temp credentials, fetching access token..\n")
	creds, _, err := oauthClient.RequestToken(urlfetch.Client(c), tempCreds.Credentials, verifier)
	if err != nil {
		c.Inc("access_token_request_fail")
		pages.ShowError(w, r, fmt.Errorf("error requesting OAuth access token: %v\n", err))
		return
	}
	accessCreds := sessions.Credentials{creds}
	err = accessCreds.Save(w, r)
	if err != nil {
		pages.ShowError(w, r, fmt.Errorf("error saving access token: %v\n", err))
		return
	}
	nextURL := "/"
	if len(referralId) > 0 {
		c.Debugf("Passing on referral id: %s", referralId)
		params := pages.Values{
			"referralId": referralId,
		}
		nextURL = params.AddTo(nextURL)
	} else if len(referralUserId) > 0 {
		c.Debugf("Passing on referral user id: %s", referralUserId)
		params := pages.Values{
			"referralUserId": referralUserId,
		}
		nextURL = params.AddTo(nextURL)
	}
	c.Inc("twitter_signin_succes")
	http.Redirect(w, r, nextURL, http.StatusFound)
}

// NewUser returns a new User instance by hitting the
// verify_credentials.json endpoint in the Twitter API.
func NewUser(w http.ResponseWriter, r *http.Request, creds *sessions.Credentials) (*TwitterUser, error) {
	c := sessions.NewContext(r)
	url := twitterBaseUrl + "/account/verify_credentials.json"
	c.Debugf("hitting %q\n", url)
	resp, err := oauthClient.Get(urlfetch.Client(c), creds.Credentials, url, nil)
	if err != nil {
		c.Inc("verify_credentials_call_fail")
		return nil, fmt.Errorf("%s call failed: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.Inc("verify_credentials_call_bad_status_fail")
		return nil, fmt.Errorf("%s returned %q: %v", url, resp.Status, resp)
	}

	var user TwitterUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		c.Inc("verify_credentials_call_bad_response_fail")
		return nil, fmt.Errorf("failed to create new user info: %v", err)
	}
	return &user, nil
}
