// metaText.go contains all the mastery stuff
package core

import (
	"gopkg.in/yaml.v2"
)

// MetaText is structure containing the parsed data from the page's meta text.
type MetaText map[string]interface{}

// ParseMetaText parses the give meta text and returns a new MetaText object.
func ParseMetaText(metaText string) (MetaText, error) {
	var metaData MetaText
	err := yaml.Unmarshal([]byte(metaText), &metaData)
	if err != nil {
		return nil, err
	}
	return metaData, nil
}
