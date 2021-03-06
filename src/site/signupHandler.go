// signupHandler.go serves the signup page.

package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/facebook"
	"zanaduu3/src/okta"
	"zanaduu3/src/pages"
	"zanaduu3/src/sessions"
)

const (
	minUserAliasLength = 4
)

// signupHandlerData is the data received from the request.
type signupHandlerData struct {
	Email     string
	FirstName string
	LastName  string
	Password  string

	// Alternatively, signup with Facebook
	FbAccessToken string
	FbUserID      string
	// Alternatively, signup with FB code token
	FbCodeToken   string
	FbRedirectURL string
}

var signupHandler = siteHandler{
	URI:         "/signup/",
	HandlerFunc: signupHandlerFunc,
	Options: pages.PageOptions{
		AllowAnyone: true,
	},
}

func signupHandlerFunc(params *pages.HandlerParams) *pages.Result {
	c := params.C
	u := params.U
	db := params.DB

	decoder := json.NewDecoder(params.R.Body)
	var data signupHandlerData
	err := decoder.Decode(&data)
	if err != nil {
		return pages.Fail("Couldn't decode json", err).Status(http.StatusBadRequest)
	}
	if len(data.FbCodeToken) > 0 && len(data.FbRedirectURL) > 0 {
		// Convert FB code token to access token + user id
		data.FbAccessToken, err = facebook.ProcessCodeToken(c, data.FbCodeToken, data.FbRedirectURL)
		if err != nil {
			return pages.Fail("Couldn't process FB code token", err)
		}
		data.FbUserID, err = facebook.ProcessAccessToken(c, data.FbAccessToken)
		if err != nil {
			return pages.Fail("Couldn't process FB token", err)
		}
	}

	socialMediaSignup := false
	if len(data.FbAccessToken) > 0 && len(data.FbUserID) >= 0 {
		// Get data from FB
		account, err := okta.CreateNewFbUser(c, data.FbAccessToken)
		if err != nil {
			return pages.Fail("Couldn't create a new user", err)
		}
		data.Email = account.Email
		data.FirstName = account.GivenName
		data.LastName = account.Surname

		// Set the cookie
		_, _, err = core.SaveCookie(params.W, params.R, data.Email)
		if err != nil {
			return pages.Fail("Couldn't save a cookie", err)
		}
		socialMediaSignup = true
	} else if len(data.Email) > 0 && len(data.FirstName) > 0 && len(data.LastName) > 0 && len(data.Password) > 0 {
		// Valid request
	} else {
		return pages.Fail("A required field is not set.", nil).Status(http.StatusBadRequest)
	}

	// Check if this user already exists.
	var existingFbUserID string
	var existingID string
	exists, err := db.NewStatement(`
		SELECT id,fbUserId
		FROM users
		WHERE email=?`).QueryRow(data.Email).Scan(&existingID, &existingFbUserID)
	if err != nil {
		return pages.Fail("Error checking for existing user", err)
	}
	if exists {
		if existingFbUserID != data.FbUserID {
			// Update user's FB id in the DB
			hashmap := make(database.InsertMap)
			hashmap["id"] = existingID
			hashmap["fbUserId"] = data.FbUserID
			statement := db.NewInsertStatement("users", hashmap, "fbUserId")
			if _, err := statement.Exec(); err != nil {
				return pages.Fail("Couldn't update user's record", err)
			}
		}

		// Load the user
		loginData := loginHandlerData{
			Email: data.Email,
		}
		return setUserInternalFunc(params, &loginData)
	}

	// Compute user's page alias and prevent collisions
	cleanupRegexp := regexp.MustCompile(core.ReplaceRegexpStr)
	aliasBase := fmt.Sprintf("%s%s", data.FirstName, data.LastName)
	aliasBase = cleanupRegexp.ReplaceAllLiteralString(aliasBase, "")
	if len(aliasBase) > 1 && '0' <= aliasBase[0] && aliasBase[0] <= '9' {
		// Only ids can start with numbers
		aliasBase = "a" + aliasBase
	}
	alias := aliasBase
	suffix := 2
	for ; len(alias) >= minUserAliasLength; suffix++ {
		var ignore int
		exists, err := database.NewQuery(`
				SELECT 1
				FROM pageInfos AS pi
				WHERE type=?`, core.GroupPageType).Add(`
					AND alias=?`, alias).Add(`
					AND`).AddPart(core.PageInfosFilter(nil)).ToStatement(db).QueryRow().Scan(&ignore)
		if err != nil {
			return pages.Fail("Error checking for existing alias", err)
		}
		if !exists {
			break
		}
		alias = fmt.Sprintf("%s%d", aliasBase, suffix)
	}

	if len(data.Password) > 0 {
		// Sign up the user through Okta
		err = okta.CreateNewUser(c, data.FirstName, data.LastName, data.Email, data.Password)
		if err != nil {
			// It could be that the user already has an account. Let's try to authenticate.
			err2 := okta.AuthenticateUser(c, data.Email, data.Password)
			if err2 != nil {
				return pages.Fail("Couldn't create a new user", err)
			}
		}
	} else {
		// If there is no password, then the user must have signed up through a social network
	}

	// Begin the transaction.
	err2 := db.Transaction(func(tx *database.Tx) sessions.Error {

		userID, err := core.GetNextAvailableID(tx)
		if err != nil {
			return sessions.NewError("Couldn't get last insert id for new user", err)
		}
		if len(alias) < minUserAliasLength {
			alias = userID
		}

		// Create new user
		hashmap := make(database.InsertMap)
		hashmap["id"] = userID
		hashmap["firstName"] = data.FirstName
		hashmap["lastName"] = data.LastName
		hashmap["email"] = data.Email
		hashmap["fbUserId"] = data.FbUserID
		hashmap["createdAt"] = database.Now()
		hashmap["lastWebsiteVisit"] = database.Now()
		hashmap["emailFrequency"] = core.DefaultEmailFrequency
		hashmap["emailThreshold"] = core.DefaultEmailThreshold
		statement := tx.DB.NewInsertStatement("users", hashmap).WithTx(tx)
		_, err = statement.Exec()
		if err != nil {
			return sessions.NewError("Couldn't update user's record", err)
		}

		// Set the user value in params, since some internal handlers we might call
		// will expect it to be set
		u.ID = userID
		u.Email = data.Email

		// Create new domain for the user.
		fullName := fmt.Sprintf("%s %s", data.FirstName, data.LastName)
		err2 := core.NewUserDomainPage(tx, u, fullName, alias)
		if err2 != nil {
			return err2
		}

		// The user might have some data stored under their session id
		if u.SessionID != "" {
			statement = database.NewQuery(`
				UPDATE userMasteryPairs SET userId=? WHERE userId=?`, userID, u.SessionID).ToTxStatement(tx)
			if _, err := statement.Exec(); err != nil {
				return sessions.NewError("Couldn't delete existing page summaries", err)
			}

			statement = database.NewQuery(`
				UPDATE userPageObjectPairs SET userId=? WHERE userId=?`, userID, u.SessionID).ToTxStatement(tx)
			if _, err := statement.Exec(); err != nil {
				return sessions.NewError("Couldn't delete existing page summaries", err)
			}
		}

		// Add an update for each invite that will be claimed
		statement = database.NewQuery(`
			INSERT INTO updates
			(userId,type,createdAt,subscribedToId,goToPageId,byUserId)
			SELECT ?,?,now(),fromUserId,domainId,fromUserId`, u.ID, core.InviteReceivedUpdateType).Add(`
			FROM invites
			WHERE toEmail=?`, u.Email).ToTxStatement(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't insert updates for invites", err)
		}

		// Claim all existing invites for this user
		statement = database.NewQuery(`
			UPDATE invites
			SET toUserId=?,claimedAt=?`, u.ID, database.Now()).Add(`
			WHERE toEmail=?`, u.Email).ToTxStatement(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't delete existing page summaries", err)
		}

		// Redeem invites
		statement = database.NewQuery(`
			INSERT INTO domainMembers
			(domainId,userId,role,createdAt)
			SELECT domainId,toUserId,role,now()
			FROM invites
			WHERE toEmail=?`, u.Email).ToTxStatement(tx)
		if _, err := statement.Exec(); err != nil {
			return sessions.NewError("Couldn't insert updates for invites", err)
		}

		return nil
	})
	if err2 != nil {
		return pages.FailWith(err2)
	}

	loginData := loginHandlerData{
		Email:    data.Email,
		Password: data.Password,
	}
	if socialMediaSignup {
		return setUserInternalFunc(params, &loginData)
	}

	return loginHandlerInternalFunc(params, &loginData)
}
