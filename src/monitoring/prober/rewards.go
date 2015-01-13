// rewards.go: Rewards page prober.
package prober

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"zanaduu3/src/config"

	"github.com/golang/glog"
	"github.com/hkjn/pages"
)

var (
	xc             = config.Load()
	twitterUser    = xc.Monitoring.Twitter.User
	twitterPw      = xc.Monitoring.Twitter.Password
	sendgridUser   = xc.Monitoring.Sendgrid.User
	sendgridPass   = xc.Monitoring.Sendgrid.Password
	xelaieCookie   = xc.Site.Cookie // name of our cookie
	twitterAuthURL = "https://api.twitter.com/oauth/authorize"
)

// RewardsProbe probes the Xelaie rewards page.
type RewardsProbe struct {
	targetURL string
}

// NewRewardsProbe returns a new instance of the probe.
func NewRewardsProbe(targetURL string) *Probe {
	return new(&RewardsProbe{targetURL}, "RewardsProbe",
		fmt.Sprintf("Follows sign-in process on %s", targetURL))
}

// Probe requests the Xelaie rewards page, then follows the sign-in
// flow like a new user would.
//
// The probe passes if the rewards page allows sign-in, fails with
// specific error otherwise.
func (p RewardsProbe) Probe() error {
	state, err := p.probeUnauthed()
	if err != nil {
		return err
	}

	nextUrl, err := p.probeTwitterAuth(state.vals)
	if err != nil {
		return err
	}

	err = p.probeAuthed(nextUrl, state.cookieJar)
	if err != nil {
		return err
	}
	return nil
}

// probeUnauthed probes the target URL, expecting the rewards page for
// a new visitor to redirect to Twitter and set a cookie.
//
// probeUnauthed returns the state needed to sign in to Twitter.
func (p RewardsProbe) probeUnauthed() (PageState, error) {
	glog.V(1).Infof("hitting %s..\n", p.targetURL)
	token := ""
	jar := &cookieJar{
		domain: p.targetURL,
		path:   "/",
		cookies: map[string]*http.Cookie{
			xelaieCookie: nil,
		},
	}
	checkAuthRedirect := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 1 {
				urls := ""
				for _, r := range via {
					urls += "%v, " + r.URL.String()
				}
				glog.Warningf("Too many redirects? Got %d previous ones: %s\n", len(via), urls)
			}
			if !strings.HasPrefix(req.URL.String(), "https://api.twitter.com/oauth/authenticate?oauth_token=") {
				return fmt.Errorf("bad URL, want prefix %q but got %v", "https://api.twitter.com/oauth/authenticate?oauth_token=", req.URL)
			}
			token = req.URL.Query().Get("oauth_token")
			return nil
		},
		Jar: jar,
	}
	resp, err := checkAuthRedirect.Get(p.targetURL)
	if err != nil {
		return PageState{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return PageState{}, fmt.Errorf("non-success status from GET %q: %v", p.targetURL, resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("error reading body: %v\n", err)
		return PageState{}, err
	}
	sb := string(body)
	want := "Xelaie LLC"
	if !strings.Contains(sb, want) {
		return PageState{}, fmt.Errorf("response doesn't contain %q: %v\n", want, sb)
	}
	// HACK: Parse out authenticity_token from response by expecting the
	// format to be character-by-character the same. (If / when this
	// breaks, add proper HTML parsing.)
	i := strings.Index(sb, "authenticity_token")
	auth_token := sb[i+40 : i+82]
	if auth_token[0] != '"' || auth_token[41] != '"' {
		return PageState{}, fmt.Errorf("failed to parse %q: %q", "authenticity_token", auth_token)
	}
	u, err := url.Parse(p.targetURL)
	if err != nil {
		return PageState{}, fmt.Errorf("failed to parse %q: %v", p.targetURL, err)
	}
	cookies := jar.Cookies(u)
	if len(cookies) != 1 {
		return PageState{}, fmt.Errorf("want 1 cookie, got %d: %+v", len(cookies), cookies)
	}
	glog.V(1).Infof("got valid response from %s, forwarding to Twitter as user %q\n", p.targetURL, twitterUser)
	return PageState{
		pages.Values{
			"oauth_token":                token,
			"session[username_or_email]": twitterUser,
			"session[password]":          twitterPw,
			"repost_after_login":         "https://api.twitter.com/oauth/authorize",
			"authenticity_token":         auth_token[1:42],
		}, jar,
	}, nil
}

// probeTwitterAuth probes the Twitter OAuth endpoint using the URL
// values received from the sign-in.
func (p RewardsProbe) probeTwitterAuth(vals pages.Values) (string, error) {
	target := twitterAuthURL
	origin := p.targetURL
	glog.V(1).Infof("hitting twitter auth %s with %v\n", target, vals)
	resp, err := http.PostForm(target, vals.UrlValues())
	if err != nil {
		return "", fmt.Errorf("POST failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-success status from POST %q: %v", target, resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading body: %v", err)
	}
	sb := string(body)
	wantUrl := origin + "/authorize_callback?oauth_token="
	if !strings.Contains(sb, wantUrl) {
		return "", fmt.Errorf("response doesn't contain %q: %v\n", wantUrl, sb)
	}
	i := strings.Index(sb, wantUrl)
	end := strings.Index(sb[i:], "\"")
	callbackUrl := sb[i : i+end]
	glog.V(2).Infof("callbackUrl=%q\n", callbackUrl)
	glog.V(1).Infof("got valid response from twitter\n")
	return callbackUrl, nil
}

// probeAuthed probes the target URL, using specified cookie jar and
// expecting the rewards page for a signed in visitor.
func (p RewardsProbe) probeAuthed(target string, jar http.CookieJar) error {
	glog.V(1).Infof("hitting %s with cookie jar %v..\n", target, jar)
	expectedNext := p.targetURL + "/"
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 1 {
				urls := ""
				for _, r := range via {
					urls += "%v, " + r.URL.String()
				}
				glog.Warningf("Too many redirects? Got %d previous ones: %s\n", len(via), urls)
			}
			if req.URL.String() != expectedNext {
				return fmt.Errorf("want redirect to %q, got %v", expectedNext, req.URL)
			}
			return nil
		},
		Jar: jar,
	}
	resp, err := client.Get(target)
	if err != nil {
		return fmt.Errorf("error from GET %s: %v", target, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-success status from GET %q: %v", target, resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %v", err)
	}
	sb := string(body)
	want := "https://twitter.com/" + twitterUser
	if !strings.Contains(sb, want) {
		return fmt.Errorf("want %q in response, didn't get it: \n%s", want, sb)
	}
	glog.V(1).Infof("got valid response from xelaie - now logged in.\n")
	return nil
}
