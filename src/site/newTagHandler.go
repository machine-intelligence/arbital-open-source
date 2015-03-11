// newTagHadler.go creates a new tag.
package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// newTagData contains data given to us in the request.
type newTagData struct {
	ParentId int64 `json:",string"`
	Text     string
}

// newTagHandler handles requests to create a tag.
func newTagHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)
	header, str := newTagProcessor(w, r)
	if header > 0 {
		if header == http.StatusInternalServerError {
			c.Inc(strings.Trim(r.URL.Path, "/") + "Fail")
		}
		c.Errorf("%s", str)
		w.WriteHeader(header)
	}
	if len(str) > 0 {
		fmt.Fprintf(w, "%s", str)
	}
}

func newTagProcessor(w http.ResponseWriter, r *http.Request) (int, string) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var data newTagData
	err := decoder.Decode(&data)
	if err != nil {
		return http.StatusBadRequest, fmt.Sprintf("Couldn't decode newTagData.")
	}
	if len(data.Text) <= 0 {
		return http.StatusBadRequest, fmt.Sprintf("Text has to be specified.")
	}

	// Load user.
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't load user: %v", err)
	}
	if !u.IsLoggedIn {
		return http.StatusForbidden, fmt.Sprintf("You are not logged in.")
	}

	// Check that the parent exists.
	var one int64
	if data.ParentId > 0 {
		query := fmt.Sprintf(`SELECT 1 FROM tags WHERE id=%d`, data.ParentId)
		exists, err := database.QueryRowSql(c, query, &one)
		if err != nil {
			return http.StatusInternalServerError, fmt.Sprintf("Couldn't load parent tag: %v", err)
		} else if !exists {
			return http.StatusInternalServerError, fmt.Sprintf("No such parent id (%d): %v", data.ParentId, err)
		}
	}

	// Compute the full name for this tag. It has to be unique, so we'll keep
	// prepending parent tag names until it becomes unique.
	// TODO: account for super short tags?
	fullName := data.Text
	fullNameIsUnique := false
	for tempParentId := data.ParentId; ; {
		// Check if the full name is unique
		query := fmt.Sprintf(`SELECT 1 FROM tags WHERE fullName="%s"`, fullName)
		exists, err := database.QueryRowSql(c, query, &one)
		if err != nil {
			return http.StatusInternalServerError, fmt.Sprintf("Couldn't query tags by fullName: %v", err)
		} else if !exists {
			// Found a unique fullName
			fullNameIsUnique = true
			// If it's too short and we have a parent, let's not stop here.
			// E.g. "D" will be named "Vitamin.D".
			if len(fullName) > 2 || tempParentId == 0 {
				break
			}
		} else if tempParentId == 0 {
			// No more parents, so this is not a unique
			break
		}

		// Get parent info and prepend parent's tag name to our fullName
		var parentFullName string
		query = fmt.Sprintf(`SELECT parentId,fullName FROM tags WHERE id=%d`, tempParentId)
		exists, err = database.QueryRowSql(c, query, &tempParentId, &parentFullName)
		if err != nil {
			return http.StatusInternalServerError, fmt.Sprintf("Couldn't load parent tag: %v", err)
		} else if !exists {
			return http.StatusInternalServerError, fmt.Sprintf("Couldn't find parent tag (%d): %v", tempParentId, err)
		}
		fullName = fmt.Sprintf("%s.%s", parentFullName, fullName)
	}
	if !fullNameIsUnique {
		return http.StatusInternalServerError, fmt.Sprintf("Tag already exists: %v", err)
	}

	// Insert new tag
	hashmap := make(map[string]interface{})
	hashmap["createdBy"] = u.Id
	hashmap["text"] = data.Text
	hashmap["fullName"] = fullName
	hashmap["parentId"] = data.ParentId
	hashmap["createdAt"] = database.Now()
	query := database.GetInsertSql("tags", hashmap)
	result, err := database.ExecuteSql(c, query)
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't create a tag: %v", err)
	}
	var tagId int64
	tagId, err = result.LastInsertId()
	if err != nil {
		return http.StatusInternalServerError, fmt.Sprintf("Couldn't get last insert id: %v", err)
	}

	return 0, fmt.Sprintf("%s,%d", fullName, tagId)
}
