// externalUrlHandler.go gets info about an external url

package site

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/pages"

	"zanaduu3/vendor/github.com/dyatlov/go-opengraph/opengraph"
	"zanaduu3/vendor/golang.org/x/net/html"
	"zanaduu3/vendor/google.golang.org/appengine"
	"zanaduu3/vendor/google.golang.org/appengine/urlfetch"
)

var externalUrlHandler = siteHandler{
	URI:         "/getExternalUrlData/",
	HandlerFunc: externalUrlHandlerFunc,
	Options: pages.PageOptions{
		RequireLogin: true,
	},
}

// externalUrlData contains parameters passed in via the request.
type externalUrlData struct {
	RawExternalUrlString string
}

// externalUrlHandlerFunc handles the request.
func externalUrlHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB
	u := params.U
	returnData := core.NewHandlerData(u)

	// Decode data
	var data externalUrlData
	err := json.NewDecoder(params.R.Body).Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode request", err).Status(http.StatusBadRequest)
	}

	externalUrl, err := url.Parse(data.RawExternalUrlString)
	if err != nil {
		// If not a valid url, just return.
		return pages.Success(returnData)
	}
	externalUrlString := externalUrl.String()

	isDupe, originalPageID, err := core.IsDuplicateExternalUrl(db, u, externalUrlString)
	if err != nil {
		return pages.Fail("Couldn't check if external url is already in use.", err)
	}

	if isDupe {
		returnData.ResultMap["isDupe"] = isDupe
		returnData.ResultMap["originalPageID"] = originalPageID

		// Load data
		core.AddPageToMap(originalPageID, returnData.PageMap, core.TitlePlusLoadOptions)
		err = core.ExecuteLoadPipeline(db, returnData)
		if err != nil {
			return pages.Fail("Pipeline error", err)
		}
	} else {
		client := &http.Client{
			Transport: &urlfetch.Transport{
				Context: db.C,
				AllowInvalidServerCertificate: appengine.IsDevAppServer(),
			},
		}
		resp, err := client.Get(externalUrlString)
		if err != nil {
			// If can't find the page, just return.
			return pages.Success(returnData)
		}

		defer resp.Body.Close()
		htmlBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return pages.Fail("Couldn't read request response.", err)
		}

		title, err := getTitle(externalUrlString, string(htmlBytes))
		if err != nil {
			return pages.Fail("Couldn't get title from html.", err)
		}
		returnData.ResultMap["title"] = title
	}

	return pages.Success(returnData)
}

func getTitle(url string, htmlString string) (string, error) {
	og := opengraph.NewOpenGraph()
	err := og.ProcessHTML(strings.NewReader(htmlString))
	if err != nil {
		return "", err
	}

	title := og.Title
	if len(title) == 0 {
		title, err = getTitleFromMetaTag(htmlString)
		if err != nil {
			return "", err
		}
	}
	title = strings.TrimSpace(title)

	// special cases to strip endings from the titles of links to various sites
	lowercaseUrl := strings.ToLower(url)
	if strings.HasPrefix(lowercaseUrl, "https://lesswrong.com") ||
		strings.HasPrefix(lowercaseUrl, "http://lesswrong.com") {
		title = strings.TrimSuffix(title, " - Less Wrong")
	}
	if strings.HasPrefix(lowercaseUrl, "https://effective-altruism.com/ea") ||
		strings.HasPrefix(lowercaseUrl, "http://effective-altruism.com/ea") {
		title = strings.TrimSuffix(title, " - Effective Altruism Forum")
	}
	if strings.HasPrefix(lowercaseUrl, "https://medium.com/ai-control") ||
		strings.HasPrefix(lowercaseUrl, "http://medium.com/ai-control") {
		title = strings.TrimSuffix(title, " – AI Control")
	}

	return title, nil
}

func getTitleFromMetaTag(htmlString string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return "", err
	}

	for queue := []*html.Node{doc}; len(queue) > 0; queue = queue[1:] {
		n := queue[0]
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content string
			for _, a := range n.Attr {
				switch a.Key {
				case "name":
					name = a.Val
				case "content":
					content = a.Val
				}
			}
			if name == "title" {
				return content, nil
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			queue = append(queue, c)
		}
	}

	return "", nil
}
