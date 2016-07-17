// Package config knows how to read config.yaml.
package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
)

var (
	configName    = "config.yaml"
	overridesName = "overrides.yaml"
)

// Load returns the config with applied overrides, or panics.
func Load() *Config {
	c, err := tryLoad(configName)
	if err != nil {
		log.Fatalf("FATAL: %v\n", err)
	}
	co, err := c.addOverrides()
	if err != nil {
		log.Fatalf("FATAL: %v\n", err)
	}
	return co
}

// load returns the config.
func load(path string) (Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("couldn't read config: %v", err)
	}

	c := Config{}
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return Config{}, fmt.Errorf("couldn't unmarshal config: %v", err)
	}
	return c, nil
}

// tryLoad returns the config by looking in parent directories,
// stepping up one level at a time a "reasonable number of times",
// until the named config file is found.
func tryLoad(name string) (Config, error) {
	var err error
	var conf Config
	tries := 0
	maxTries := 5 // five levels of directories ought to be enough for anyone </gates>
	path := name
	for tries < maxTries {
		conf, err = load(path)
		if err == nil {
			return conf, nil
		}
		path = filepath.Join("..", path)
		tries += 1
	}
	return Config{}, fmt.Errorf("failed to find a valid %q: %v", name, err)
}

// AddOverrides applies the values in specified overrides file to the config.
func (c *Config) addOverrides() (*Config, error) {
	o, err := tryLoad(overridesName)
	if err != nil {
		// Note: Since it's not required to have an overrides.yaml, we
		// treat a failed load as a non-error. It would be nice to log an
		// INFO message at this point to alert the caller that overrides
		// file is missing (to make the feature more discoverable), but we
		// can't use glog in case we're called from AppEngine.
		return c, nil
	}
	err = mergo.Merge(c, o)
	if err != nil {
		return nil, fmt.Errorf("couldn't merge config with overrides: %v\n", overridesName, err)
	}
	return c, nil
}

// Config describes the structure of config.yaml.
//
// Not all values in config.yaml need to be represented here, just the
// ones needed in Go clients.
type Config struct {
	Twitter struct {
		Token, Secret string
		Daemon        struct {
			Token, Secret string
		}
	}
	MySQL struct {
		User, Password, Database string
		Live                     struct {
			Instance, Address string
		}
		Root struct {
			Password string
		}
	}
	Vm struct {
		Monitoring struct {
			Instance, Zone, Address string
		}
	}
	Site struct {
		Cookie string
		Live   struct {
			Address string
		}
		Dev struct {
			Address, Auth string
		}
		Session struct {
			Auth, Crypt string
		}
	}
	Elastic struct {
		Live struct {
			Address        string
			User, Password string
		}
		Dev struct {
			Address string
		}
	}
	Stormpath struct {
		ID                         string
		Secret                     string
		Tenant                     string
		AccountStoreMappings       string
		Accounts                   string
		ApiKeys                    string
		AuthTokens                 string
		CustomData                 string
		DefaultAccountStoreMapping string
		DefaultGroupStoreMapping   string
		Groups                     string
		Production                 string
		Stage                      string
		LoginAttempts              string
		OAuthPolicy                string
		PasswordResetTokens        string
		VerificationEmails         string
	}
	Facebook struct {
		ID         string
		Secret     string
		TestId     string `yaml:"testId"`
		TestSecret string `yaml:"testSecret"`
	}
	Mailchimp struct {
		ID         string
		Secret     string
		Apikey     string
		Production string
	}
	Monitoring struct {
		Whitelist []string
		Sendgrid  struct {
			User, Password string
		}
		Twitter struct {
			User, Password string
		}
		Service struct {
			ID, Secret string
		}
		Influx struct {
			Database struct {
				Live, Dev string
			}
			Root struct {
				Password string
			}
			Monitoring struct {
				User, Password string
			}
		}
	}
}
