// signupPage.go serves the signup page.
package site

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/pages"
	"zanaduu3/src/user"
)

// signupData stores the data that we pass to the signup.tmpl to render the page
type signupData struct {
	User        *user.User
	ContinueUrl string
}

// signupPage serves the signup page.
var signupPage = newPageWithOptions(
	"/signup/",
	signupRenderer,
	append(baseTmpls,
		"tmpl/signupPage.tmpl"),
	pages.PageOptions{})

// signupRenderer renders the signup page.
func signupRenderer(params *pages.HandlerParams) *pages.Result {
	u := params.U
	db := params.DB

	var data signupData
	data.User = u

	if data.User.Id <= 0 {
		return pages.RedirectWith(data.User.LoginLink)
	}

	// Check if there are parameters. In this case, this is a form submit
	// request. We can process it and, if successful, redirect the user
	// to the page they came from / were trying to get to.
	q := params.R.URL.Query()
	firstName := q.Get("firstName")
	lastName := q.Get("lastName")
	if firstName != "" || lastName != "" {
		// This is a form submission.
		if len(firstName) <= 0 || len(lastName) <= 0 {
			return pages.Fail("Must specify both first and last names", nil)
		}

		nameRegexp := regexp.MustCompile("^[A-Za-z]+$")
		if !nameRegexp.MatchString(firstName) || !nameRegexp.MatchString(lastName) {
			return pages.Fail("Only letter characters are allowed in the name", nil)
		}

		// Process invite code and assign karma
		inviteCode := strings.ToUpper(q.Get("inviteCode"))
		karma := 0
		if inviteCode == "TRUTH" {
			karma = 200
		} else {
			return pages.Fail("Need invite code", nil)
		}
		if data.User.Karma > karma {
			karma = data.User.Karma
		}

		// Set default email settings
		emailFrequency := user.DefaultEmailFrequency
		emailThreshold := user.DefaultEmailThreshold

		// Prevent alias collision
		aliasBase := fmt.Sprintf("%s%s", firstName, lastName)
		alias := aliasBase
		suffix := 2
		for ; ; suffix++ {
			var ignore int
			exists, err := db.NewStatement(`
				SELECT 1
				FROM pages
				WHERE type="group" AND alias=?`).QueryRow(alias).Scan(&ignore)
			if err != nil {
				return pages.Fail("Error checking for existing alias", err)
			}
			if !exists {
				break
			}
			alias = fmt.Sprintf("%s%d", aliasBase, suffix)
		}

		// Begin the transaction.
		errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
			hashmap := make(database.InsertMap)
			hashmap["id"] = data.User.Id
			hashmap["firstName"] = firstName
			hashmap["lastName"] = lastName
			hashmap["email"] = data.User.Email
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
			data.User.FirstName = firstName
			data.User.LastName = lastName
			data.User.Karma = karma
			data.User.IsLoggedIn = true
			data.User.EmailFrequency = emailFrequency
			data.User.EmailThreshold = emailThreshold
			err := data.User.Save(params.W, params.R)
			if err != nil {
				return "Couldn't re-save the user after adding the name", err
			}

			// Create new group for the user.
			fullName := fmt.Sprintf("%s %s", firstName, lastName)
			hashmap = make(database.InsertMap)
			hashmap["pageId"] = data.User.Id
			hashmap["edit"] = 1
			hashmap["title"] = fullName
			hashmap["alias"] = alias
			hashmap["creatorId"] = data.User.Id
			hashmap["createdAt"] = database.Now()
			hashmap["type"] = core.GroupPageType
			hashmap["editGroupId"] = data.User.Id
			hashmap["clickbait"] = "Automatically generated group for " + fullName
			hashmap["isCurrentEdit"] = true
			statement = tx.NewInsertTxStatement("pages", hashmap)
			if _, err = statement.Exec(); err != nil {
				return "Couldn't create a new page", err
			}

			// Add new group to pageInfos.
			hashmap = make(database.InsertMap)
			hashmap["pageId"] = data.User.Id
			hashmap["currentEdit"] = 1
			hashmap["maxEdit"] = 1
			hashmap["createdAt"] = database.Now()
			statement = tx.NewInsertTxStatement("pageInfos", hashmap)
			if _, err = statement.Exec(); err != nil {
				return "Couldn't create a new page", err
			}

			// Add user to their own group.
			hashmap = make(database.InsertMap)
			hashmap["userId"] = data.User.Id
			hashmap["groupId"] = data.User.Id
			hashmap["createdAt"] = database.Now()
			statement = tx.NewInsertTxStatement("groupMembers", hashmap)
			if _, err = statement.Exec(); err != nil {
				return "Couldn't add user to the group", err
			}
			return "", nil
		})
		if errMessage != "" {
			return pages.Fail(errMessage, err)
		}

		continueUrl := q.Get("continueUrl")
		if continueUrl == "" {
			continueUrl = "/"
		}
		return pages.RedirectWith(continueUrl)
	}
	data.ContinueUrl = q.Get("continueUrl")
	return pages.StatusOK(data)
}
