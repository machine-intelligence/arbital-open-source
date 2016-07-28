// creds.go: handles credentials
package sessions

import (
	"fmt"
	"net/http"

	"github.com/garyburd/go-oauth/oauth"
)

var (
	credsKey   = "credentials" // key for session storage
	emptyCreds = Credentials{}
	FakeCreds  = Credentials{&oauth.Credentials{"FAKE_TOKEN", "FAKE_SECRET"}}
)

type Credentials struct {
	*oauth.Credentials
}

// Save stores the credentials in session.
func (creds *Credentials) Save(w http.ResponseWriter, r *http.Request) error {
	s, err := GetSession(r)
	if err != nil {
		return fmt.Errorf("couldn't get session: %v", err)
	}

	s.Values[credsKey] = *creds
	err = s.Save(r, w)
	if err != nil {
		return fmt.Errorf("failed to save credentials key to session: %v", err)
	}
	return nil
}
