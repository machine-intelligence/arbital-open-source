// pageInfoHandler.go contains the handler for returning information for a given page.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"

	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// pageInfoData contains parameters passed in to create a page.
type pageInfoData struct {
	PageId     int64 `json:",string"`
	PrivacyKey string
}

// pageInfoHandler handles requests to create a new page.
func pageInfoHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	// Decode data
	decoder := json.NewDecoder(r.Body)
	var data pageInfoData
	err := decoder.Decode(&data)
	if err != nil {
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Load user object
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("page_handler_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load the page.
	var pagePtr *page
	pagePtr, err = loadPage(c, data.PageId)
	if err != nil {
		c.Errorf("Couldn't load a page: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	p := &richPage{page: *pagePtr}
	pageIdStr := fmt.Sprintf("%d", p.PageId)
	pageMap := make(map[int64]*richPage)
	pageMap[p.PageId] = p

	// Check privacy setting
	if p.PrivacyKey.Valid && fmt.Sprintf("%d", p.PrivacyKey.Int64) != data.PrivacyKey {
		c.Warningf("Didn't specify correct privacy key: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Load answers
	err = p.loadAnswers(c)
	if err != nil {
		c.Errorf("Couldn't load answers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load likes
	err = loadLikes(c, u.Id, pageIdStr, pageMap)
	if err != nil {
		c.Errorf("Couldn't load likes: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Load probability votes
	err = loadVotes(c, u.Id, pageIdStr, pageMap)
	if err != nil {
		c.Errorf("Couldn't load probability votes: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return the page in JSON format.
	var jsonData []byte
	jsonData, err = json.Marshal(p)
	if err != nil {
		fmt.Println("Error marshalling page into json:", err)
	}
	fmt.Fprintf(w, "%s", jsonData)
}
