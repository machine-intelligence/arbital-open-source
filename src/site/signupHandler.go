// signupHandler.go serves the signup page.
package site

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/facebook"
	"zanaduu3/src/pages"
	"zanaduu3/src/stormpath"
	"zanaduu3/src/user"
)

// signupHandlerData is the data received from the request.
type signupHandlerData struct {
	Email      string
	FirstName  string
	LastName   string
	Password   string
	InviteCode string

	// Alternatively, signup with Facebook
	FbAccessToken string
	FbUserId      string
	// Alternatively, signup with FB code token
	FbCodeToken   string
	FbRedirectUrl string
}

var signupHandler = siteHandler{
	URI:         "/signup/",
	HandlerFunc: signupHandlerFunc,
	Options: pages.PageOptions{
		AllowAnyone: true,
	},
}

func signupHandlerFunc(params *pages.HandlerParams) *pages.Result {
	db := params.DB

	decoder := json.NewDecoder(params.R.Body)
	var data signupHandlerData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.HandlerBadRequestFail("Couldn't decode json", err)
	}
	if len(data.FbCodeToken) > 0 && len(data.FbRedirectUrl) > 0 {
		// Convert FB code token to access token + user id
		data.FbAccessToken, err = facebook.ProcessCodeToken(params.C, data.FbCodeToken, data.FbRedirectUrl)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't process FB code token", err)
		}
		data.FbUserId, err = facebook.ProcessAccessToken(params.C, data.FbAccessToken)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't process FB token", err)
		}
	}
	if len(data.FbAccessToken) > 0 && len(data.FbUserId) >= 0 {
		// Get data from FB
		account, err := stormpath.CreateNewFbUser(params.C, data.FbAccessToken)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't create a new user", err)
		}
		data.Email = account.Email
		data.FirstName = account.GivenName
		data.LastName = account.Surname

		// Set the cookie
		err = user.SaveEmailCookie(params.W, params.R, data.Email)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't save a cookie", err)
		}
	} else if len(data.Email) > 0 && len(data.FirstName) > 0 && len(data.LastName) > 0 && len(data.Password) > 0 {
		// Valid request
	} else {
		return pages.HandlerBadRequestFail("A required field is not set.", nil)
	}

	// Check if this user already exists.
	var existingFbUserId string
	var existingId string
	exists, err := db.NewStatement(`
		SELECT id,fbUserId
		FROM users
		WHERE email=?`).QueryRow(data.Email).Scan(&existingId, &existingFbUserId)
	if err != nil {
		return pages.HandlerErrorFail("Error checking for existing user", err)
	}
	if exists {
		if existingFbUserId != data.FbUserId {
			// Update user's FB id in the DB
			hashmap := make(database.InsertMap)
			hashmap["id"] = existingId
			hashmap["fbUserId"] = data.FbUserId
			statement := db.NewInsertStatement("users", hashmap, "fbUserId")
			if _, err := statement.Exec(); err != nil {
				return pages.HandlerErrorFail("Couldn't update user's record", err)
			}
		}
		return pages.StatusOK(nil)
	}

	// Process invite code and assign karma
	inviteCode := strings.ToUpper(data.InviteCode)
	karma := 0
	if inviteCode == core.CorrectInviteCode {
		karma = core.CorrectInviteKarma
	}

	// Prevent alias collision
	cleanupRegexp := regexp.MustCompile(core.ReplaceRegexpStr)
	aliasBase := fmt.Sprintf("%s%s", data.FirstName, data.LastName)
	aliasBase = cleanupRegexp.ReplaceAllLiteralString(aliasBase, "")
	if len(aliasBase) <= 3 {
		return pages.HandlerBadRequestFail("Not enough good characters for an alias", nil)
	} else if '0' <= aliasBase[0] && aliasBase[0] <= '9' {
		// Only ids can start with numbers
		aliasBase = "a" + aliasBase
	}
	alias := aliasBase
	suffix := 2
	for ; ; suffix++ {
		var ignore int
		exists, err := db.NewStatement(`
				SELECT 1
				FROM pageInfos
				WHERE type="group" AND alias=?`).QueryRow(alias).Scan(&ignore)
		if err != nil {
			return pages.HandlerErrorFail("Error checking for existing alias", err)
		}
		if !exists {
			break
		}
		alias = fmt.Sprintf("%s%d", aliasBase, suffix)
	}

	// If there is no password, then the user must have signed up through a social network
	if len(data.Password) > 0 {
		// Sign up the user through Stormpath
		err = stormpath.CreateNewUser(params.C, data.FirstName, data.LastName, data.Email, data.Password)
		if err != nil {
			return pages.HandlerErrorFail("Couldn't create a new user", err)
		}
	}

	// Begin the transaction.
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {

		userId, err := user.GetNextAvailableId(tx)
		if err != nil {
			return "", fmt.Errorf("Couldn't get last insert id for new user: %v", err)
		}

		hashmap := make(database.InsertMap)
		hashmap["id"] = userId
		hashmap["firstName"] = data.FirstName
		hashmap["lastName"] = data.LastName
		hashmap["email"] = data.Email
		hashmap["fbUserId"] = data.FbUserId
		hashmap["createdAt"] = database.Now()
		hashmap["lastWebsiteVisit"] = database.Now()
		hashmap["inviteCode"] = inviteCode
		hashmap["karma"] = karma
		hashmap["emailFrequency"] = user.DefaultEmailFrequency
		hashmap["emailThreshold"] = user.DefaultEmailThreshold
		statement := tx.DB.NewInsertStatement("users", hashmap).WithTx(tx)
		_, err = statement.Exec()
		if err != nil {
			return "Couldn't update user's record", err
		}

		// Create new group for the user.
		fullName := fmt.Sprintf("%s %s", data.FirstName, data.LastName)
		errorMessage, err := core.NewUserGroup(tx, userId, fullName, alias)
		if errorMessage != "" {
			return errorMessage, err
		}

		// Set the user value in params, since some internal handlers we might call
		// will expect it to be set
		params.U.Id = userId

		// The user might have some data stored under their session id
		if params.U.SessionId != "" {
			statement = database.NewQuery(`
				UPDATE userMasteryPairs SET userId=? WHERE userId=?`, userId, params.U.SessionId).ToTxStatement(tx)
			if _, err := statement.Exec(); err != nil {
				return "Couldn't delete existing page summaries", err
			}

			statement = database.NewQuery(`
				UPDATE userPageObjectPairs SET userId=? WHERE userId=?`, userId, params.U.SessionId).ToTxStatement(tx)
			if _, err := statement.Exec(); err != nil {
				return "Couldn't delete existing page summaries", err
			}
		}

		// Signup for that page
		return addSubscription(tx, userId, userId, 0)
	})
	if errMessage != "" {
		return pages.HandlerErrorFail(errMessage, err)
	}

	return pages.StatusOK(nil)
}
