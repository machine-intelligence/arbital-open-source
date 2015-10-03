// signupPage.go serves the signup page.
package site

import (
	"fmt"
	"strings"

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
		"tmpl/signupPage.tmpl", "tmpl/navbar.tmpl"),
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

		// Process invite code and assign karma
		inviteCode := strings.ToUpper(q.Get("inviteCode"))
		karma := 0
		if inviteCode == "BAYES" || inviteCode == "LESSWRONG" {
			karma = 200
		} else {
			return pages.Fail("Need invite code", nil)
		}
		if data.User.Karma > karma {
			karma = data.User.Karma
		}

		// Begin the transaction.
		err := db.Transaction(func(tx *database.Tx) error {
			hashmap := make(database.InsertMap)
			hashmap["id"] = data.User.Id
			hashmap["firstName"] = firstName
			hashmap["lastName"] = lastName
			hashmap["inviteCode"] = inviteCode
			hashmap["karma"] = karma
			hashmap["createdAt"] = database.Now()
			statement := tx.NewInsertTxStatement("users", hashmap, "firstName", "lastName", "inviteCode", "karma")
			if _, err := statement.Exec(); err != nil {
				return fmt.Errorf("Couldn't update user's record: %v", err)
			}
			data.User.FirstName = firstName
			data.User.LastName = lastName
			data.User.Karma = karma
			data.User.IsLoggedIn = true
			err := data.User.Save(params.W, params.R)
			if err != nil {
				return fmt.Errorf("Couldn't re-save the user after adding the name: %v", err)
			}

			// Add new group for the user.
			hashmap = make(database.InsertMap)
			hashmap["id"] = data.User.Id
			hashmap["name"] = fmt.Sprintf("%s_%s", firstName, lastName)
			hashmap["createdAt"] = database.Now()
			hashmap["isVisible"] = true
			statement = tx.NewInsertTxStatement("groups", hashmap)
			if _, err = statement.Exec(); err != nil {
				return fmt.Errorf("Couldn't create a new group: %v", err)
			}

			// Add user to their own group.
			hashmap = make(database.InsertMap)
			hashmap["userId"] = data.User.Id
			hashmap["groupId"] = data.User.Id
			hashmap["createdAt"] = database.Now()
			statement = tx.NewInsertTxStatement("groupMembers", hashmap)
			if _, err = statement.Exec(); err != nil {
				return fmt.Errorf("Couldn't add user to the group: %v", err)
			}
			return nil
		})
		if err != nil {
			return pages.Fail("Couldn't exec transaction", err)
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
