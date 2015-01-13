// cookies.go: Cookie support for probes.
package prober

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/golang/glog"
)

// cookieJar is a http.CookieJar implementation that only allows specified
// cookies.
type cookieJar struct {
	cookies map[string]*http.Cookie // map of names to accept to stored cookies
	domain  string                  // domain to match for
	path    string                  // expected path on domain
}

func (cj *cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	glog.V(2).Infof("SetCookies(%v, %v)\n", u, cookies)
	if !strings.HasPrefix(u.String(), cj.domain) {
		return
	}
	for _, c := range cookies {
		if c.Path != cj.path {
			continue
		}
		_, ok := cj.cookies[c.Name]
		if !ok {
			continue
		}
		glog.V(2).Infof("cookie match, storing %v\n", c)
		cj.cookies[c.Name] = c
	}
}
func (cj *cookieJar) Cookies(u *url.URL) []*http.Cookie {
	glog.V(2).Infof("Cookies(%v) (has %v)\n", u, cj.cookies)
	if !strings.HasPrefix(u.String(), cj.domain) {
		return []*http.Cookie{}
	}
	r := []*http.Cookie{}
	for _, c := range cj.cookies {
		if c != nil {
			r = append(r, c)
		}
	}
	return r
}
