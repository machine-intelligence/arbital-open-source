// Package config provides a wrapper around config.yaml
package config

import "log"

var (
	// We unfortunately must carry around a copy of the config.yaml
	// (canonically in the top directory of the repo), as AppEngine
	// otherwise won't be smart enough to deploy it to live. (It will
	// work on dev, but break the live site.)
	XC = Load()
)

func init() {
	if XC.Site.Session.Auth == "" {
		log.Fatalf("FATAL: missing site.session.auth value in config.yaml - can't use encrypted cookies!\n")
	}
	if XC.Site.Session.Crypt == "" {
		log.Fatalf("FATAL: missing site.session.crypt value in config.yaml - can't use encrypted cookies!\n")
	}
}
