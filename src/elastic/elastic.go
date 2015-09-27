// elastic.go contains all the stuff for working with ElasticSearch.
package elastic

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"

	"appengine/urlfetch"

	"zanaduu3/src/config"
	"zanaduu3/src/sessions"
)

const (
	rootPEM = `-----BEGIN CERTIFICATE-----
MIICjTCCAfYCCQCdtV757poJ4TANBgkqhkiG9w0BAQUFADCBijELMAkGA1UEBhMC
VVMxEzARBgNVBAgMCkNhbGlmb3JuaWExETAPBgNVBAcMCEJlcmtlbGV5MREwDwYD
VQQKDAhPbW5pbWVudDEXMBUGA1UEAwwOQWxleGVpIEFuZHJlZXYxJzAlBgkqhkiG
9w0BCQEWGGFsZXhlaS5hbmRyZWV2QGdtYWlsLmNvbTAeFw0xNTA5MjYyMzIzMDda
Fw0yNTA5MjMyMzIzMDdaMIGKMQswCQYDVQQGEwJVUzETMBEGA1UECAwKQ2FsaWZv
cm5pYTERMA8GA1UEBwwIQmVya2VsZXkxETAPBgNVBAoMCE9tbmltZW50MRcwFQYD
VQQDDA5BbGV4ZWkgQW5kcmVldjEnMCUGCSqGSIb3DQEJARYYYWxleGVpLmFuZHJl
ZXZAZ21haWwuY29tMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCz+XzDdko6
Q0tmHTBZ2oVzpZwY3nl11wFNnDZnkmzEa9zAPPN9k/ez5vEc+ZhtvZzne+AHq9YO
PQQTa7ee9Mtj3OvwQIhJHR2qFhqPgdtZlJU5BWf9kuyiQnaTCPomjYz3S8D4M52g
Vt3kNVR5EbbAR1hgw8mIYCC+Fsop57IL1QIDAQABMA0GCSqGSIb3DQEBBQUAA4GB
AGFNjsZBejXq/1wVeBTc00mXCAOhI0XpFRzwFJf4ILgrl+ylK0s9GorJGqled1Gh
ArkOL1SI0oJr/CQugA/6X99hzLoAk7cDnx9kAaRAEmXGsz4w0o2sBySyWXcMzvQZ
EpJfDRwd0JPE7D4Vck5SYMl0zN/GZHlzgUbRmXB26lyd
-----END CERTIFICATE-----`
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
type Result struct {
	Hits []Hits `json:"hits"`
}

type Hits struct {
	Total    int     `json:"total"`
	MaxScore float64 `json:"max_core"`
	Hits     []Hit   `json:"hits"`
}

type Hit struct {
	Score  float64  `json:"_score"`
	Source Document `json:"_source"`
}

// sendRequest sends the given request object to the elastic search server.
func sendRequest(c sessions.Context, request *http.Request) (*http.Response, error) {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM([]byte(rootPEM))
	tlsConfig := &tls.Config{RootCAs: certPool}
	tlsConfig.BuildNameToCertificate()
	transport := &urlfetch.Transport{Context: c, AllowInvalidServerCertificate: true}

	//tr := &urlfetch.Transport{Context: c, Deadline: TimeoutDuration, AllowInvalidServerCertificate: allowInvalidServerCertificate}

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
func SearchPageIndex(c sessions.Context, jsonStr string) (map[string]interface{}, error) {
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
	var results map[string]interface{}
	err = decoder.Decode(&results)
	if err != nil {
		return nil, fmt.Errorf("Couldn't decode json: %v", err)
	}
	return results, nil
}
