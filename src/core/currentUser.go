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
	// NOTE: all the numbers are made up right now. The only real number is 200
	CommentKarmaReq            = -5
	LikeKarmaReq               = 0
	PrivatePageKarmaReq        = 5
	VoteKarmaReq               = 10
	AddParentKarmaReq          = 20
	CreateAliasKarmaReq        = 50
	EditPageKarmaReq           = 200
	DeleteParentKarmaReq       = 100
	KarmaLockKarmaReq          = 100
	ChangeSortChildrenKarmaReq = 100
	ChangeAliasKarmaReq        = 200
	DeletePageKarmaReq         = 200
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
	EmailFrequency string `json:"emailFrequency"`
	EmailThreshold int    `json:"emailThreshold"`
	IgnoreMathjax  bool   `json:"ignoreMathjax"`

	// If the user isn't logged in, this is set to their unique session id
	SessionId string `json:"-"`

	// Computed variables
	UpdateCount    int                `json:"updateCount"`
	GroupIds       []string           `json:"groupIds"`
	TrustMap       map[string]*Trust  `json:"trustMap"`
	InvitesClaimed map[string]*Invite `json:"invitesClaimed"`
	// If set, these are the lists the user is subscribed to via mailchimp
	MailchimpInterests map[string]bool `json:"mailchimpInterests"`
}

// Invite represents a code we can send to invite one or more users.
type Invite struct {
	Code      string     `json:"code"`
	Type      string     `json:"type"`
	SenderId  string     `json:"senderId"`
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

	CanDeletePage bool `json:"canDeletePage"`
	CanEditPage   bool `json:"canEditPage"`
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
		SELECT id,fbUserId,email,firstName,lastName,isAdmin,isTrusted,
			emailFrequency,emailThreshold,ignoreMathjax
		FROM users
		WHERE email=?`).QueryRow(cookie.Email)
	exists, err := row.Scan(&u.Id, &u.FbUserId, &u.Email, &u.FirstName, &u.LastName,
		&u.IsAdmin, &u.IsTrusted, &u.EmailFrequency, &u.EmailThreshold, &u.IgnoreMathjax)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a user: %v", err)
	} else if !exists {
		return nil, fmt.Errorf("Couldn't find that email in DB")
	}

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
func LoadUserTrust(db *database.DB, u *CurrentUser) error {
	domainIds, err := LoadAllDomainIds(db, nil)
	if err != nil {
		return err
	}

	for _, domainId := range domainIds {
		u.TrustMap[domainId] = &Trust{}
	}

	// NOTE: this should come last in computing trust, so that the bonus trust from
	// an invite slowly goes away as the user accumulates real trust.
	// Compute trust from invites
	wherePart := database.NewQuery(`WHERE ie.claimingUserId=?`, u.Id)
	inviteMap, err := LoadInvitesWhere(db, wherePart)
	if err != nil {
		return fmt.Errorf("Couldn't process existing invites: %v", err)
	}
	for _, invite := range inviteMap {
		if u.TrustMap[invite.DomainId].EditTrust < DefaultInviteKarma {
			u.TrustMap[invite.DomainId].EditTrust = DefaultInviteKarma
		}
	}

	// Now comupte permissions
	for _, trust := range u.TrustMap {
		trust.CanEditPage = trust.EditTrust >= EditPageKarmaReq || u.IsAdmin
		trust.CanDeletePage = trust.EditTrust >= DeletePageKarmaReq || u.IsAdmin
	}

	return nil
}

// Get invites where a certain column matches a certain query string
// Returns map: inviteCode -> invite
func LoadInvitesWhere(db *database.DB, wherePart *database.QueryPart) (map[string]*Invite, error) {
	invites := make(map[string]*Invite)
	rows := database.NewQuery(`
		SELECT i.code,i.type,i.domainId,i.senderId,i.createdAt,ie.email,ie.claimingUserId,ie.claimedAt
		FROM inviteEmailPairs AS ie
		JOIN invites AS i
		ON (i.code=ie.code)`).AddPart(wherePart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		invite := &Invite{}
		invitee := &Invitee{}
		err := rows.Scan(&invite.Code, &invite.Type, &invite.DomainId, &invite.SenderId,
			&invite.CreatedAt, &invitee.Email, &invitee.ClaimingUserId, &invitee.ClaimedAt)
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
		return nil, fmt.Errorf("Error while loading invites WHERE %v: %v", wherePart, err)
	}
	return invites, nil
}

// ClaimCode claims the given code for the given user, and gives them the appropriate amount
// of bonus trust.
// If the code was claimed successfully, the Invite object is returned.
// If the code was already claimed / couldn't be found, the Invite object is nil.
func ClaimCode(tx *database.Tx, inviteCode string, claimingEmail string, claimingUserId string) (*Invite, error) {
	inviteCode = strings.ToUpper(inviteCode)

	wherePart := database.NewQuery(`WHERE ie.code=?`, inviteCode).Add(`AND ie.email=?`, claimingEmail)
	inviteMap, err := LoadInvitesWhere(tx.DB, wherePart)
	if err != nil {
		return nil, fmt.Errorf("Failed to scan row for matching invites: %v", err)
	}
	invite, ok := inviteMap[inviteCode]
	if !ok {
		// No valid invite found
		return nil, nil
	}
	invitee := invite.Invitees[0]
	if invitee.ClaimingUserId != "" {
		// Invite is already claimed
		return nil, nil
	}

	hashmap := make(database.InsertMap)
	hashmap["code"] = inviteCode
	hashmap["email"] = claimingEmail
	hashmap["claimingUserId"] = claimingUserId
	hashmap["claimedAt"] = database.Now()
	statement := tx.DB.NewInsertStatement("inviteEmailPairs", hashmap, "claimingUserId", "claimedAt").WithTx(tx)
	if _, err := statement.Exec(); err != nil {
		return nil, fmt.Errorf("Couldn't create invite email pair", err)
	}

	return invite, nil
}

func init() {
	// Need to register this so it can be stored inside a cookie
	gob.Register(&CookieSession{})
}
