// Package googleauth handles OAuth sign-in using Google+.
//
// The package panics at initialization time if service id / secret
// are not present in the config.
package googleauth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"xelaie/src/go/config"

	"code.google.com/p/goauth2/oauth"
	"github.com/golang/glog"
	"github.com/gorilla/sessions"
)

// config is the configuration specification supplied to the OAuth package.
var (
	xc           = config.Load()
	clientId     = xc.Monitoring.Service.Id
	clientSecret = xc.Monitoring.Service.Secret
	oauthConfig  = &oauth.Config{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		// Scope determines which API calls we are authorized to make.
		Scope:    "profile",
		AuthURL:  "https://accounts.google.com/o/oauth2/auth",
		TokenURL: "https://accounts.google.com/o/oauth2/token",
		// Use "postmessage" for the code-flow for server side apps.
		RedirectURL: "postmessage",
	}
)

// store initializes the Gorilla session store.
var store = sessions.NewCookieStore([]byte(randomString(32)))

// Token represents an OAuth token response.
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IdToken     string `json:"id_token"`
}

// ClaimSet represents an IdToken response.
type ClaimSet struct {
	Sub string
}

// IsLoggedIn returns true if the user is signed in.
func IsLoggedIn(r *http.Request) (bool, error) {
	session, err := getSession(r)
	if err != nil {
		return false, err
	}
	t := session.Values["accessToken"]
	if t == nil {
		return false, nil
	}
	storedToken, ok := t.(string)
	if !ok {
		return false, fmt.Errorf("bad type of %q value in session: %v", "accessToken", err)
	}
	return storedToken != "", nil
}

// exchange takes an authentication code and exchanges it with the OAuth
// endpoint for a Google API bearer token and a Google+ ID
func exchange(code string) (accessToken string, idToken string, err error) {
	// Exchange the authorization code for a credentials object via a POST request
	addr := "https://accounts.google.com/o/oauth2/token"
	values := url.Values{
		"Content-Type":  {"application/x-www-form-urlencoded"},
		"code":          {code},
		"client_id":     {clientId},
		"client_secret": {clientSecret},
		"redirect_uri":  {oauthConfig.RedirectURL},
		"grant_type":    {"authorization_code"},
	}
	resp, err := http.PostForm(addr, values)
	if err != nil {
		return "", "", fmt.Errorf("error exchanging code: %v", err)
	}
	defer resp.Body.Close()

	// Decode the response body into a token object
	var token Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return "", "", fmt.Errorf("error decoding access token: %v", err)
	}

	return token.AccessToken, token.IdToken, nil
}

// decodeIdToken takes an ID Token and decodes it to fetch the Google+ ID within
func decodeIdToken(idToken string) (gplusID string, err error) {
	// An ID token is a cryptographically-signed JSON object encoded in
	// base 64.  Normally, it is critical to validate an ID token before
	// you use it, but since we are communicating directly with Google
	// over an intermediary-free HTTPS channel and using the Client
	// Secret to authenticate ourselves, we can be confident that the
	// token you receive really comes from Google and is valid. If this
	// is ever passed outside the monitoring app, it is extremely
	// important that the other components validate the token before
	// using it.
	var set ClaimSet
	if idToken != "" {
		// Check that the padding is correct for a base64decode
		parts := strings.Split(idToken, ".")
		if len(parts) < 2 {
			return "", fmt.Errorf("bad ID token")
		}
		// Decode the ID token
		b, err := base64Decode(parts[1])
		if err != nil {
			return "", fmt.Errorf("bad ID token: %v", err)
		}
		err = json.Unmarshal(b, &set)
		if err != nil {
			return "", fmt.Errorf("bad ID token: %v", err)
		}
	}
	return set.Sub, nil
}

// getSession returns the user's session.
func getSession(r *http.Request) (*sessions.Session, error) {
	session, err := store.Get(r, "sessionName")
	if err != nil {
		if session.IsNew {
			glog.V(1).Infof("ignoring initial session fetch error since session IsNew\n")
		} else {
			return nil, fmt.Errorf("error fetching session: %v", err)
		}
	}
	return session, nil
}

// isAllowed returns whether specified G+ id is allowed to access the app.
func isAllowed(gplusId string) bool {
	for _, id := range xc.Monitoring.Whitelist {
		if gplusId == id {
			return true
		}
	}
	return false
}

type LoginInfo struct {
	ClientId string // id of client
	StateURI string // URL with one-time authorization code
}

// CheckLogIn returns the user's login info.
func CheckLogIn(w http.ResponseWriter, r *http.Request) (*LoginInfo, error) {
	// Create a state token to prevent request forgery and store it in the session
	// for later validation
	session, err := getSession(r)
	if err != nil {
		return nil, err
	}

	state := randomString(64)
	session.Values["state"] = state
	err = session.Save(r, w)
	if err != nil {
		return nil, fmt.Errorf("failed to save state in session: %v", err)
	}
	glog.V(1).Infof("CheckLogin set %q=%v in user's session\n", "state", state)
	return &LoginInfo{clientId, url.QueryEscape(state)}, nil
}

// Connect exchanges the one-time authorization code for a token and stores the
// token in the session
func Connect(w http.ResponseWriter, r *http.Request) error {
	// Ensure that the request is not a forgery and that the user sending this
	// connect request is the expected user.
	session, err := getSession(r)
	if err != nil {
		return err
	}
	q := r.URL.Query()
	if session.Values["state"] == nil {
		return fmt.Errorf("missing %q variable in session for user trying to log in? bug, or user is trying to spoof log in", "state")
	}
	sessionState := session.Values["state"].(string)
	if q.Get("state") != sessionState {
		// Note: This can happen if CheckLogIn is called multiple times
		// for the same session, e.g. when several tabs are loading
		// protected resources.
		return fmt.Errorf("state mismatch, got %q from form, but had %q in session\n", r.FormValue("state"), sessionState)
	}
	session.Values["state"] = nil

	code := q.Get("code")
	if code == "" {
		return fmt.Errorf("missing %q value in request body", "code")
	}
	glog.V(1).Infof("code=%v\n", code)
	// We got back matching state from user as well as auth code from
	// login button, exchange the one-time auth code for access token +
	// user id.
	accessToken, idToken, err := exchange(code)
	if err != nil {
		return fmt.Errorf("couldn't exchange code for access token: %v", err)
	}
	gplusId, err := decodeIdToken(idToken)
	glog.V(1).Infof("decoded G+ token: %v\n", gplusId)
	if err != nil {
		return fmt.Errorf("couldn't decode ID token: %v", err)
	}

	if !isAllowed(gplusId) {
		return fmt.Errorf("user with G+ %v is not allowed access\n", gplusId)
	}
	glog.V(1).Infof("User %v is allowed to log in\n", gplusId)

	// Check if the user is already connected
	storedToken := session.Values["accessToken"]
	storedGPlusId := session.Values["gplusID"]
	if storedToken != nil && storedGPlusId == gplusId {
		return nil
	}

	// Store the access token in the session for later use
	session.Values["accessToken"] = accessToken
	session.Values["gplusID"] = gplusId
	err = session.Save(r, w)
	if err != nil {
		return fmt.Errorf("failed to save state in session: %v", err)
	}
	return nil
}

// Disconnect revokes the current user's token and resets their session
func Disconnect(w http.ResponseWriter, r *http.Request) error {
	// Only disconnect a connected user.
	session, err := store.Get(r, "sessionName")
	if err != nil {
		return fmt.Errorf("error fetching session: %v", err)
	}
	token := session.Values["accessToken"]
	if token == nil {
		glog.Infof("Current user not connected\n")
		return nil
	}

	// Execute HTTP GET request to revoke current token.
	url := "https://accounts.google.com/o/oauth2/revoke?token=" + token.(string)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to revoke token for given user")
	}
	defer resp.Body.Close()

	// Reset the user's session.
	session.Values["accessToken"] = nil
	err = session.Save(r, w)
	if err != nil {
		return fmt.Errorf("failed to save state in session: %v", err)
	}
	return nil
}

// randomString returns a random string with the specified length
func randomString(length int) (str string) {
	b := make([]byte, length)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func base64Decode(s string) ([]byte, error) {
	// add back missing padding
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

func init() {
	if clientId == "" {
		glog.Fatalf("FATAL: missing monitoring.service.id in config\n")
	}
	if clientSecret == "" {
		glog.Fatalf("FATAL: missing monitoring.service.secret in config\n")
	}
}
