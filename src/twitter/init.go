// Package twitter is used for all things related to Twitter and Twitter API
package twitter

import (
	"encoding/gob"
)

func init() {
	gob.Register(&referralFlash{})
	gob.Register(&TwitterUser{})
}
