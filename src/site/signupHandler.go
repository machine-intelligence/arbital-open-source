// signupPage.go serves the signup page.
package site

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/user"
)

// signupHandlerData is the data received from the request.
type signupHandlerData struct {
	Email       string
	FirstName   string
	LastName    string
	InviteCode  string
	ContinueUrl string
}

var signupHandler = siteHandler{
	URI:         "/signup/",
	HandlerFunc: signupHandlerFunc,
	Options:     pages.PageOptions{},
}

func signupHandlerFunc(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	if u.Id <= 0 {
		return pages.HandlerForbiddenFail("Need to login", nil)
	}

	decoder := json.NewDecoder(params.R.Body)
	var data signupHandlerData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if len(data.Email) <= 0 || len(data.FirstName) <= 0 || len(data.LastName) <= 0 {
		return pages.HandlerBadRequestFail("Must specify email, first and last names.", nil)
	}
	nameRegexp := regexp.MustCompile("^[A-Za-z]+$")
	if !nameRegexp.MatchString(data.FirstName) || !nameRegexp.MatchString(data.LastName) {
		return pages.HandlerBadRequestFail("Only letter characters are allowed in the name", nil)
	}

	// Process invite code and assign karma
	inviteCode := strings.ToUpper(data.InviteCode)
	karma := 0
	if inviteCode == "TRUTH" {
		karma = 200
	}
	if u.Karma > karma {
		karma = u.Karma
	}

	// Set default email settings
	emailFrequency := user.DefaultEmailFrequency
	emailThreshold := user.DefaultEmailThreshold

	// Prevent alias collision
	aliasBase := fmt.Sprintf("%s%s", data.FirstName, data.LastName)
	alias := aliasBase
	suffix := 2
	for ; ; suffix++ {
		var ignore int
		exists, err := db.NewStatement(`
				SELECT 1
				FROM pages
				WHERE type="group" AND alias=?`).QueryRow(alias).Scan(&ignore)
		if err != nil {
			return pages.HandlerErrorFail("Error checking for existing alias", err)
		}
		if !exists {
			break
		}
		alias = fmt.Sprintf("%s%d", aliasBase, suffix)
	}

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		hashmap := make(database.InsertMap)
		hashmap["id"] = u.Id
		hashmap["firstName"] = data.FirstName
		hashmap["lastName"] = data.LastName
		hashmap["email"] = data.Email
		hashmap["createdAt"] = database.Now()
		hashmap["lastWebsiteVisit"] = database.Now()
		hashmap["inviteCode"] = inviteCode
		hashmap["karma"] = karma
		// Don't send emails to anyone yet (except Alexei and Eliezer)
		hashmap["updateEmailSentAt"] = time.Now().UTC().Add(30000 * time.Hour).Format(database.TimeLayout)
		hashmap["emailFrequency"] = emailFrequency
		hashmap["emailThreshold"] = emailThreshold
		statement := tx.NewReplaceTxStatement("users", hashmap)
		if _, err := statement.Exec(); err != nil {
			return "Couldn't update user's record", err
		}
		u.FirstName = data.FirstName
		u.LastName = data.LastName
		u.Karma = karma
		u.IsLoggedIn = true
		u.EmailFrequency = emailFrequency
		u.EmailThreshold = emailThreshold
		err := u.Save(params.W, params.R)
		if err != nil {
			return "Couldn't re-save the user after adding the name", err
		}

		// Create new group for the user.
		fullName := fmt.Sprintf("%s %s", data.FirstName, data.LastName)
		return core.NewUserGroup(tx, u.Id, fullName, alias)
	})
	if errMessage != "" {
		return pages.HandlerErrorFail(errMessage, err)
	}

	if data.ContinueUrl == "" {
		data.ContinueUrl = "/"
	}
	return pages.RedirectWith(data.ContinueUrl)
}
