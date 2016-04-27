// Package user manages information about the current user.
package core

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

const (
	DailyEmailFrequency       = "daily"
	WeeklyEmailFrequency      = "weekly"
	NeverEmailFrequency       = "never"
	ImmediatelyEmailFrequency = "immediately"

	PersonalInviteType = "personal"
	GroupInviteType    = "group"

	Base31Chars             = "0123456789bcdfghjklmnpqrstvwxyz"
	Base31CharsForFirstChar = "0123456789"
)

const (
	// Karma requirements to perform various actions
	CommentKarmaReq            = -5
	LikeKarmaReq               = 0
	PrivatePageKarmaReq        = 5
	VoteKarmaReq               = 10
	AddParentKarmaReq          = 20
	CreateAliasKarmaReq        = 50
	EditPageKarmaReq           = 0 //50 // edit wiki page
	DeleteParentKarmaReq       = 100
	KarmaLockKarmaReq          = 100
	ChangeSortChildrenKarmaReq = 100
	ChangeAliasKarmaReq        = 200
	DeletePageKarmaReq         = 500
	DashlessAliasKarmaReq      = 1000

	DefaultInviteKarma = 200
)

var (
	sessionKey = "arbitalSession2" // key for session storage

	// Highest karma lock a user can create is equal to their karma * this constant.
	MaxKarmaLockFraction = 0.8

	DefaultEmailFrequency = DailyEmailFrequency
	DefaultEmailThreshold = 3
)

// User holds information about the current user of the app.
// Note: this structure is also stored in a cookie.
type CurrentUser struct {
	// DB variables
	Id             string `json:"id"`
	FbUserId       string `json:"fbUserId"`
	Email          string `json:"email"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	IsAdmin        bool   `json:"isAdmin"`
	IsTrusted      bool   `json:"isTrusted"`
	Karma          int    `json:"karma"`
	MaxKarmaLock   int    `json:"maxKarmaLock"`
	EmailFrequency string `json:"emailFrequency"`
	EmailThreshold int    `json:"emailThreshold"`
	InviteCode     string `json:"inviteCode"`
	IgnoreMathjax  bool   `json:"ignoreMathjax"`

	// If the user isn't logged in, this is set to their unique session id
	SessionId string `json:"-"`

	// Computed variables
	UpdateCount    int                `json:"updateCount"`
	GroupIds       []string           `json:"groupIds"`
	TrustMap       map[string]*Trust  `json:"trust"`
	InvitesClaimed map[string]*Invite `json:"invitesClaimed"`
	// If set, these are the lists the user is subscribed to via mailchimp
	MailchimpInterests map[string]bool `json:"mailchimpInterests"`
}

// Invite represents a code we can send to invite one or more users.
type Invite struct {
	Code      string     `json:"code"`
	Type      string     `json:"type"`
	DomainId  string     `json:"domainId"`
	Invitees  []*Invitee `json:"invitees"`
	CreatedAt string     `json:"createdAt"`
}

// Invitee is an invited user.
type Invitee struct {
	Email          string `json:"email"`
	ClaimingUserId string `json:"claimingUserId"`
	ClaimedAt      string `json:"claimedAt"`
}

// InviteMatch is the result for whether or not there exists a match for an invite code
type InviteMatch struct {
	Invite     *Invite
	Invitee    *Invitee
	CodeMatch  bool
	EmailMatch bool
}

// Trust has the different scores for how much we trust a user.
type Trust struct {
	GeneralTrust int `json:"generalTrust"`
	EditTrust    int `json:"editTrust"`
}

type CookieSession struct {
	Email     string
	SessionId string

	// Randomly generated string (for security/encryption reasons)
	Random string
}

func NewUser() *CurrentUser {
	var u CurrentUser
	u.GroupIds = make([]string, 0)
	u.TrustMap = make(map[string]*Trust)
	u.InvitesClaimed = make(map[string]*Invite)
	u.MailchimpInterests = make(map[string]bool)
	return &u
}

func (user *CurrentUser) FullName() string {
	return user.FirstName + " " + user.LastName
}

// GetSomeId returns user's id or, if not available, session id, which could still be ""
func (user *CurrentUser) GetSomeId() string {
	if user.Id != "" {
		return user.Id
	}
	return user.SessionId
}

// GetMaxKarmaLock returns the highest possible karma lock a user with the
// given amount of karma can create.
func GetMaxKarmaLock(karma int) int {
	return int(float64(karma) * MaxKarmaLockFraction)
}

// IsMemberOfGroup returns true iff the user is member of the given group.
// NOTE: we are assuming GroupIds have been loaded.
func (user *CurrentUser) IsMemberOfGroup(groupId string) bool {
	isMember := false
	oldGroupIdStr := groupId
	for _, groupIdStr := range user.GroupIds {
		if groupIdStr == oldGroupIdStr {
			isMember = true
			break
		}
	}
	return isMember
}

// Store user's email in a cookie
func SaveEmailCookie(w http.ResponseWriter, r *http.Request, email string) error {
	s, err := sessions.GetSession(r)
	if err != nil {
		return fmt.Errorf("Couldn't get session: %v", err)
	}

	rand.Seed(time.Now().UnixNano())
	s.Values[sessionKey] = CookieSession{Email: email, Random: fmt.Sprintf("%d", rand.Int63())}
	err = s.Save(r, w)
	if err != nil {
		return fmt.Errorf("Failed to save user to session: %v", err)
	}
	return nil
}

// Store a unique id in a cookie so we can track session
func saveSessionCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	s, err := sessions.GetSession(r)
	if err != nil {
		return "", fmt.Errorf("Couldn't get session: %v", err)
	}

	rand.Seed(time.Now().UnixNano())
	sessionId := fmt.Sprintf("sid:%d", rand.Int63())
	s.Values[sessionKey] = CookieSession{
		SessionId: sessionId,
		Random:    fmt.Sprintf("%d", rand.Int63()),
	}
	err = s.Save(r, w)
	if err != nil {
		return "", fmt.Errorf("Failed to save user to session: %v", err)
	}
	return sessionId, nil
}

// loadUserFromDb tries to load the current user's info from the database. If
// there is no data in the DB, but the user is logged in through AppEngine,
// a new record is created.
func loadUserFromDb(w http.ResponseWriter, r *http.Request, db *database.DB) (*CurrentUser, error) {
	// Load email from the cookie
	s, err := sessions.GetSession(r)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get session: %v", err)
	}
	u := NewUser()

	var cookie *CookieSession
	if cookieStruct, ok := s.Values[sessionKey]; !ok {
		sessionId, err := saveSessionCookie(w, r)
		u.SessionId = sessionId
		return u, err
	} else {
		cookie = cookieStruct.(*CookieSession)
	}
	if cookie.Email == "" {
		u.SessionId = cookie.SessionId
		return u, err
	}

	row := db.NewStatement(`
		SELECT id,fbUserId,email,firstName,lastName,isAdmin,isTrusted,karma,
			emailFrequency,emailThreshold,inviteCode,ignoreMathjax
		FROM users
		WHERE email=?`).QueryRow(cookie.Email)
	exists, err := row.Scan(&u.Id, &u.FbUserId, &u.Email, &u.FirstName, &u.LastName,
		&u.IsAdmin, &u.IsTrusted, &u.Karma, &u.EmailFrequency, &u.EmailThreshold,
		&u.InviteCode, &u.IgnoreMathjax)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a user: %v", err)
	} else if !exists {
		return nil, fmt.Errorf("Couldn't find that email in DB")
	}
	u.MaxKarmaLock = GetMaxKarmaLock(u.Karma)
	return u, nil
}

// LoadUser returns user object corresponding to logged in user. First, we check
// if the user is logged in via App Engine. If they are, we make sure they are
// in the database. If the user is not logged in, we return a partially filled
// User object.
// A user object is returned iff there is no error.
func LoadCurrentUser(w http.ResponseWriter, r *http.Request, db *database.DB) (userPtr *CurrentUser, err error) {
	userPtr, err = loadUserFromDb(w, r, db)
	if err != nil {
		return
	} else if userPtr == nil {
		userPtr = NewUser()
	}
	return
}

// Replace a rune at a specific index in a string
func replaceAtIndex(in string, r rune, i int) string {
	out := []rune(in)
	out[i] = r
	return string(out)
}

// Get the next highest base36 character, without vowels
// Returns the character, and true if it wrapped around to 0
// Since we decided that ids must begin with a digit, only allow characters 0-9 for the first character index
func GetNextBase31Char(c sessions.Context, char rune, isFirstChar bool) (rune, bool, error) {
	validChars := Base31Chars
	if isFirstChar {
		validChars = Base31CharsForFirstChar
	}
	index := strings.Index(validChars, strings.ToLower(string(char)))
	if index < 0 {
		return '0', false, fmt.Errorf("invalid character")
	}
	if index < len(validChars)-1 {
		nextChar := rune(validChars[index+1])
		return nextChar, false, nil
	} else {
		nextChar := rune(validChars[0])
		return nextChar, true, nil
	}
}

// Increment a base31 Id string
func IncrementBase31Id(c sessions.Context, previousId string) (string, error) {
	// Add 1 to the base36 value, skipping vowels
	// Start at the last character in the Id string, carrying the 1 as many times as necessary
	nextAvailableId := previousId
	index := len(nextAvailableId) - 1
	var newChar rune
	var err error
	processNextChar := true
	for processNextChar {
		// If we need to carry the 1 all the way to the beginning, then add a 1 at the beginning of the string
		if index < 0 {
			nextAvailableId = "1" + nextAvailableId
			processNextChar = false
		} else {
			// Increment the character at the current index in the Id string
			newChar, processNextChar, err = GetNextBase31Char(c, rune(nextAvailableId[index]), index == 0)
			if err != nil {
				return "", fmt.Errorf("Error processing id: %v", err)
			}
			nextAvailableId = replaceAtIndex(nextAvailableId, newChar, index)
			index = index - 1
		}
	}

	return nextAvailableId, nil
}

// Call GetNextAvailableId in a new transaction
func GetNextAvailableIdInNewTransaction(db *database.DB) (string, error) {
	return db.Transaction(func(tx *database.Tx) (string, error) {
		return GetNextAvailableId(tx)
	})
}

// Get the next available base36 Id string that doesn't contain vowels
func GetNextAvailableId(tx *database.Tx) (string, error) {
	// Query for the highest used pageId or userId
	var highestUsedId string
	row := database.NewQuery(`
		SELECT MAX(pageId)
		FROM (
			SELECT pageId
			FROM pageInfos
			UNION
			SELECT id
			FROM users
		) AS combined
		WHERE char_length(pageId) = (
			SELECT MAX(char_length(pageId))
			FROM (
				SELECT pageId
				FROM pageInfos
				UNION
				SELECT id
				FROM users
			) AS combined2
    )
		`).ToTxStatement(tx).QueryRow()
	_, err := row.Scan(&highestUsedId)
	if err != nil {
		return "", fmt.Errorf("Couldn't load id: %v", err)
	}
	return IncrementBase31Id(tx.DB.C, highestUsedId)
}

// LoadUpdateCount returns the number of unseen updates the given user has.
func LoadUpdateCount(db *database.DB, userId string) (int, error) {
	editTypes := []string{PageEditUpdateType, CommentEditUpdateType}

	var editUpdateCount int
	row := database.NewQuery(`
		SELECT COUNT(DISTINCT type, subscribedToId, byUserId)
		FROM updates
		WHERE unseen AND userId=?`, userId).Add(`
			AND type IN`).AddArgsGroupStr(editTypes).ToStatement(db).QueryRow()
	_, err := row.Scan(&editUpdateCount)
	if err != nil {
		return -1, err
	}

	var nonEditUpdateCount int
	row = database.NewQuery(`
		SELECT COUNT(*)
		FROM updates
		WHERE unseen AND userId=?`, userId).Add(`
			AND type NOT IN`).AddArgsGroupStr(editTypes).ToStatement(db).QueryRow()
	_, err = row.Scan(&nonEditUpdateCount)
	if err != nil {
		return -1, err
	}

	return editUpdateCount + nonEditUpdateCount, err
}

// LoadUserTrust returns the trust that the user has in all domains.
func LoadUserTrust(db *database.DB, u *CurrentUser) (map[string]*Trust, error) {
	trustMap := make(map[string]*Trust)

	domainIds, err := LoadAllDomainIds(db, nil)
	if err != nil {
		return nil, err
	}

	for _, domainId := range domainIds {
		trustMap[domainId] = &Trust{}
	}

	// Compute bonus trust
	rows := db.NewStatement(`
		SELECT domainId,bonusEditTrust
		FROM userDomainBonusTrust
		WHERE userId=?`).Query(u.Id)
	err = rows.Process(func(db *database.DB, rows *database.Rows) error {
		var domainId string
		var bonusEditTrust int
		err := rows.Scan(&domainId, &bonusEditTrust)
		if err != nil {
			return fmt.Errorf("failed to scan an invite: %v", err)
		}
		trustMap[domainId].EditTrust += bonusEditTrust
		return nil
	})
	if err != nil {
		return nil, err
	}

	return trustMap, nil
}

// Get invites where a certain column matches a certain query string
func GetInvitesWhere(db *database.DB, whereColumn string, query string) (map[string]*Invite, error) {
	invites := make(map[string]*Invite)
	rows := database.NewQuery(`
		SELECT i.code,i.type,i.domainId,i.createdAt,ie.email,ie.claimingUserId,ie.claimedAt
		FROM inviteEmailPairs AS ie
		JOIN invites AS i
		ON (i.code=ie.code)
		WHERE`).Add(whereColumn).Add(`=?`).ToStatement(db).Query(query)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		invite := &Invite{}
		invitee := &Invitee{}
		err := rows.Scan(&invite.Code, &invite.Type, &invite.DomainId, &invite.CreatedAt,
			&invitee.Email, &invitee.ClaimingUserId, &invitee.ClaimedAt)
		if err != nil {
			return fmt.Errorf("failed to scan an invite: %v", err)
		}
		if existingInvite, ok := invites[invite.Code]; !ok {
			invite.Invitees = make([]*Invitee, 0)
			invites[invite.Code] = invite
		} else {
			invite = existingInvite
		}
		invite.Invitees = append(invite.Invitees, invitee)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error while loading domain ids WHERE %v='query': %v", whereColumn, err)
	}
	return invites, nil
}

// See if an invite code matches one in the db, and if the email matches
func MatchInvite(db *database.DB, code string, userEmail string) (*InviteMatch, error) {
	var invite Invite
	var invitee Invitee
	invite.Invitees = append(invite.Invitees, &invitee)
	codeMatch := false
	emailMatch := false
	rows := db.NewStatement(`
		SELECT ie.code, ie.email, ie.claimingUserId, i.domainId, i.type
		FROM inviteEmailPairs as ie
		JOIN invites as i
		ON (i.code=ie.code)
		WHERE ie.code=?`).Query(code)
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		err := rows.Scan(&invite.Code, &invitee.Email, &invitee.ClaimingUserId, &invite.DomainId, &invite.Type)
		if err != nil {
			return fmt.Errorf("Failed to scan row for matching invites: %v", err)
		}
		if invite.Code != "" {
			codeMatch = true
		}
		if userEmail == invitee.Email {
			emailMatch = true
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Can't process invites to match: %v", err)
	}
	inviteMatch := InviteMatch{&invite, &invitee, codeMatch, emailMatch}
	return &inviteMatch, err
}

// ClaimCode claims the given code for the current user, and gives them the appropriate amount
// of bonus trust.
func ClaimCode(db *database.DB, inviteCode string, domainId string, u *CurrentUser) (int, error) {
	// Update or insert row into inviteEmailPairs and userDomainBonusTrust
	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		hashmap := make(database.InsertMap)
		hashmap["code"] = inviteCode
		hashmap["email"] = u.Email
		hashmap["claimingUserId"] = u.Id
		hashmap["claimedAt"] = database.Now()
		statement := tx.DB.NewInsertStatement("inviteEmailPairs", hashmap, "claimingUserId", "claimedAt").WithTx(tx)
		if _, err := statement.Exec(); err != nil {
			return "Couldn't create invite email pair", err
		}

		hashmap = make(database.InsertMap)
		hashmap["domainId"] = domainId
		hashmap["bonusEditTrust"] = DefaultInviteKarma
		hashmap["userId"] = u.Id
		statement = db.NewInsertStatement("userDomainBonusTrust", hashmap, "bonusEditTrust").WithTx(tx)
		if _, err := statement.Exec(); err != nil {
			return "Couldn't set bonus trust in domain", err
		}

		return "", nil
	})
	if errMessage != "" {
		return 0, fmt.Errorf("Transaction failed: %s %v", errMessage, err)
	}
	return DefaultInviteKarma, nil
}

func init() {
	// Need to register this so it can be stored inside a cookie
	gob.Register(&CookieSession{})
}
