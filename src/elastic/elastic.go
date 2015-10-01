// elastic.go contains all the stuff for working with ElasticSearch.
package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"appengine/urlfetch"

	"zanaduu3/src/config"
	"zanaduu3/src/sessions"
)

var (
	ElasticDomain = fmt.Sprintf("%s/arbital", sessions.GetElasticDomain())
)

// Document describes the document which goes into the pages search index.
type Document struct {
	PageId    int64  `json:"pageId,string"`
	Alias     string `json:"alias"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Clickbait string `json:"clickbait"`
	Text      string `json:"text"`
	GroupId   int64  `json:"groupId,string"`
	CreatorId int64  `json:"creatorId,string"`
}

// All the elasticsearch result structs
type SearchResult struct {
	Hits *Hits `json:"hits"`
}

type Hits struct {
	Total    int     `json:"total"`
	MaxScore float32 `json:"max_score"`
	Hits     []*Hit  `json:"hits"`
}

type Hit struct {
	Id     int64     `json:"_id,string"`
	Score  float32   `json:"_score"`
	Source *Document `json:"_source"`
}

func EscapeMatchTerm(text string) string {
	escapeRx := regexp.MustCompile(`(["\\])`)
	return escapeRx.ReplaceAllStringFunc(text, func(term string) string {
		return `\` + term
	})
}

// sendRequest sends the given request object to the elastic search server.
func sendRequest(c sessions.Context, request *http.Request) (*http.Response, error) {
	transport := &urlfetch.Transport{Context: c, AllowInvalidServerCertificate: true}
	resp, err := transport.RoundTrip(request)
	if err != nil {
		return nil, fmt.Errorf("Round trip failed: %v", err)
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		// Process an error
		decoder := json.NewDecoder(resp.Body)
		var result map[string]interface{}
		err = decoder.Decode(&result)
		if err != nil {
			return nil, fmt.Errorf("Elastic returned '%s', but couldn't decode json: %v", resp.Status, err)
		}
		return nil, fmt.Errorf("Elastic returned '%s': %+v", resp.Status, result)
	}
	return resp, nil
}

// AddPageToIndex adds a page document to the pages index.
func AddPageToIndex(c sessions.Context, doc *Document) error {
	// Construct request body
	jsonData, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("Error marshalling data into json:", err)
	}
	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/page/%d", ElasticDomain, doc.PageId), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Couldn't create request: %v", err)
	}
	if sessions.Live {
		request.SetBasicAuth(config.XC.Elastic.Live.User, config.XC.Elastic.Live.Password)
	}
	request.Header.Set("Content-Type", "application/json")

	// Execute request
	_, err = sendRequest(c, request)
	if err != nil {
		return fmt.Errorf("Couldn't execute request: %v", err)
	}

	return nil
}

// DeletePageFromIndex dletes a page from the pages index.
func DeletePageFromIndex(c sessions.Context, pageId int64) error {
	request, err := http.NewRequest("DELETE", fmt.Sprintf("%s/page/%d", ElasticDomain, pageId), nil)
	if err != nil {
		return fmt.Errorf("Couldn't create request: %v", err)
	}
	if sessions.Live {
		request.SetBasicAuth(config.XC.Elastic.Live.User, config.XC.Elastic.Live.Password)
	}
	request.Header.Set("Content-Type", "application/json")

	// Execute request
	_, err = sendRequest(c, request)
	if err != nil {
		return fmt.Errorf("Couldn't execute request: %v", err)
	}

	return nil
}

type IndexSchema struct {
	Mappings map[string]*Mapping `json:"mappings"`
}

type Mapping struct {
	Properties   map[string]*Property `json:"properties"`
	IncludeInAll bool                 `json:"include_in_all"`
	Dynamic      string               `json:"dynamic"`
}

type Property struct {
	Type     string `json:"type,omitempty"`
	Index    string `json:"index,omitempty"`
	Analyzer string `json:"analyzer,omitempty"`
}

// CreatePageIndex creates the pages index.
func CreatePageIndex(c sessions.Context) error {
	var mapping Mapping
	mapping.IncludeInAll = false
	mapping.Dynamic = "strict"
	mapping.Properties = make(map[string]*Property)
	mapping.Properties["pageId"] = &Property{Type: "string", Index: "not_analyzed"}
	mapping.Properties["type"] = &Property{Type: "string", Index: "not_analyzed"}
	mapping.Properties["title"] = &Property{Type: "string", Analyzer: "english"}
	mapping.Properties["clickbait"] = &Property{Type: "string", Analyzer: "english"}
	mapping.Properties["text"] = &Property{Type: "string", Analyzer: "english"}
	mapping.Properties["alias"] = &Property{Type: "string"}
	mapping.Properties["groupId"] = &Property{Type: "string", Index: "not_analyzed"}
	mapping.Properties["creatorId"] = &Property{Type: "string", Index: "not_analyzed"}

	var schema IndexSchema
	schema.Mappings = make(map[string]*Mapping)
	schema.Mappings["page"] = &mapping

	// Construct request body
	jsonData, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("Error marshalling data into json:", err)
	}
	request, err := http.NewRequest("PUT", ElasticDomain, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Couldn't create request: %v", err)
	}
	if sessions.Live {
		request.SetBasicAuth(config.XC.Elastic.Live.User, config.XC.Elastic.Live.Password)
	}
	request.Header.Set("Content-Type", "application/json")

	// Execute request
	_, err = sendRequest(c, request)
	if err != nil {
		return fmt.Errorf("Couldn't execute request: %v", err)
	}

	return nil
}

// DeletePageIndex deletes the pages index.
func DeletePageIndex(c sessions.Context) error {
	request, err := http.NewRequest("DELETE", ElasticDomain, nil)
	if err != nil {
		return fmt.Errorf("Couldn't create request: %v", err)
	}
	if sessions.Live {
		request.SetBasicAuth(config.XC.Elastic.Live.User, config.XC.Elastic.Live.Password)
	}
	request.Header.Set("Content-Type", "application/json")

	// Execute request
	_, err = sendRequest(c, request)
	if err != nil {
		return fmt.Errorf("Couldn't execute request: %v", err)
	}

	return nil
}

// SearchPageIndex searches the pages index.
func SearchPageIndex(c sessions.Context, jsonStr string) (*SearchResult, error) {
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/page/_search", ElasticDomain), bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return nil, fmt.Errorf("Couldn't create request: %v", err)
	}
	if sessions.Live {
		request.SetBasicAuth(config.XC.Elastic.Live.User, config.XC.Elastic.Live.Password)
	}
	request.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := sendRequest(c, request)
	if err != nil {
		return nil, fmt.Errorf("Couldn't execute request: %v", err)
	}

	// Process results
	decoder := json.NewDecoder(resp.Body)
	var searchResult SearchResult
	err = decoder.Decode(&searchResult)
	if err != nil {
		return nil, fmt.Errorf("Couldn't decode json: %v", err)
	}
	return &searchResult, nil
}
