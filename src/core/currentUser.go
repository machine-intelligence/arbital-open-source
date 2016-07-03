// Package user manages information about the current user.
package core

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

const (
	DailyEmailFrequency       = "daily"
	WeeklyEmailFrequency      = "weekly"
	NeverEmailFrequency       = "never"
	ImmediatelyEmailFrequency = "immediately"

	Base31Chars             = "0123456789bcdfghjklmnpqrstvwxyz"
	Base31CharsForFirstChar = "0123456789"
)

const (
	// Karma requirements to perform various actions
	// NOTE: all the numbers are made up right now. The only real number is 200
	CommentKarmaReq    = 200
	EditPageKarmaReq   = 200
	DeletePageKarmaReq = 200

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
	Id                     string `json:"id"`
	FbUserId               string `json:"fbUserId"`
	Email                  string `json:"email"`
	FirstName              string `json:"firstName"`
	LastName               string `json:"lastName"`
	IsAdmin                bool   `json:"isAdmin"`
	IsTrusted              bool   `json:"isTrusted"`
	EmailFrequency         string `json:"emailFrequency"`
	EmailThreshold         int    `json:"emailThreshold"`
	IgnoreMathjax          bool   `json:"ignoreMathjax"`
	ShowAdvancedEditorMode bool   `json:"showAdvancedEditorMode"`
	IsSlackMember          bool   `json:"isSlackMember"`

	// If the user isn't logged in, this is set to their unique session id
	SessionId string `json:"-"`

	// Computed variables
	// Set to true if the user is a member of at least one domain
	IsDomainMember                bool              `json:"isDomainMember"`
	HasReceivedMaintenanceUpdates bool              `json:"hasReceivedMaintenanceUpdates"`
	HasReceivedNotifications      bool              `json:"hasReceivedNotifications"`
	UpdateCount                   int               `json:"updateCount"`
	NewNotificationCount          int               `json:"newNotificationCount"`
	NewAchievementCount           int               `json:"newAchievementCount"`
	MaintenanceUpdateCount        int               `json:"maintenanceUpdateCount"`
	GroupIds                      []string          `json:"groupIds"`
	TrustMap                      map[string]*Trust `json:"trustMap"`
	InvitesClaimed                []*Invite         `json:"invitesClaimed"`
	// If set, these are the lists the user is subscribed to via mailchimp
	MailchimpInterests map[string]bool `json:"mailchimpInterests"`
}

// Invite is an invitation from a trusted user to another to participate in a domain.
type Invite struct {
	FromUserId  string `json:"fromUserId"`
	DomainId    string `json:"domainId"`
	ToEmail     string `json:"toEmail"`
	CreatedAt   string `json:"createdAt"`
	ToUserId    string `json:"toUserId"`
	ClaimedAt   string `json:"claimedAt"`
	EmailSentAt string `json:"emailSentAt"`
}

// Trust has the different scores for how much we trust a user.
type Trust struct {
	Permissions Permissions `json:"permissions"`

	// Note that we don't want to send the trust numbers to the FE.
	GeneralTrust int `json:"-"`
	EditTrust    int `json:"-"`
}

type CookieSession struct {
	Email     string
	SessionId string

	// Randomly generated string (for security/encryption reasons)
	Random string
}

func NewCurrentUser() *CurrentUser {
	var u CurrentUser
	u.GroupIds = make([]string, 0)
	u.TrustMap = make(map[string]*Trust)
	u.InvitesClaimed = make([]*Invite, 0)
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

// Store a unique session id (and email if they're logged in) in a cookie so we can track the user's session
func SaveCookie(w http.ResponseWriter, r *http.Request, email string) (string, error) {
	s, err := sessions.GetSession(r)
	if err != nil {
		return "", fmt.Errorf("Couldn't get session: %v", err)
	}

	rand.Seed(time.Now().UnixNano())
	sessionId := fmt.Sprintf("sid:%d", rand.Int63())
	s.Values[sessionKey] = CookieSession{
		Email:     email,
		SessionId: sessionId,
		Random:    fmt.Sprintf("%d", rand.Int63()),
	}
	err = s.Save(r, w)
	if err != nil {
		return "", fmt.Errorf("Failed to save user to session: %v", err)
	}
	return sessionId, nil
}

// Load user by id. If u object is given, load data into it. Otherwise create a new user object.
func LoadCurrentUserFromDb(db *database.DB, userId string, u *CurrentUser) (*CurrentUser, error) {
	if u == nil {
		u = NewCurrentUser()
	}
	row := db.NewStatement(`
		SELECT id,fbUserId,email,firstName,lastName,isAdmin,isTrusted,isSlackMember,
			emailFrequency,emailThreshold,ignoreMathjax,showAdvancedEditorMode
		FROM users
		WHERE id=?`).QueryRow(userId)
	exists, err := row.Scan(&u.Id, &u.FbUserId, &u.Email, &u.FirstName, &u.LastName,
		&u.IsAdmin, &u.IsTrusted, &u.IsSlackMember, &u.EmailFrequency, &u.EmailThreshold, &u.IgnoreMathjax,
		&u.ShowAdvancedEditorMode)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load user: %v", err)
	} else if !exists {
		return nil, fmt.Errorf("Couldn't find the user")
	}
	return u, nil
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
	u := NewCurrentUser()

	var cookie *CookieSession
	if cookieStruct, ok := s.Values[sessionKey]; !ok {
		sessionId, err := SaveCookie(w, r, "")
		u.SessionId = sessionId
		return u, err
	} else {
		cookie = cookieStruct.(*CookieSession)
	}
	u.SessionId = cookie.SessionId
	if cookie.Email == "" {
		return u, err
	}

	var pretendToBeUserId string
	row := db.NewStatement(`
		SELECT id,pretendToBeUserId,fbUserId,email,firstName,lastName,isAdmin,isTrusted,
			isSlackMember,emailFrequency,emailThreshold,ignoreMathjax,showAdvancedEditorMode
		FROM users
		WHERE email=?`).QueryRow(cookie.Email)
	exists, err := row.Scan(&u.Id, &pretendToBeUserId, &u.FbUserId, &u.Email, &u.FirstName, &u.LastName,
		&u.IsAdmin, &u.IsTrusted, &u.IsSlackMember, &u.EmailFrequency, &u.EmailThreshold, &u.IgnoreMathjax,
		&u.ShowAdvancedEditorMode)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a user: %v", err)
	} else if !exists {
		return nil, fmt.Errorf("Couldn't find that email in DB")
	}

	// Admins can pretened to be certain users
	if u.IsAdmin && pretendToBeUserId != "" {
		u, err = LoadCurrentUserFromDb(db, pretendToBeUserId, u)
		if err != nil {
			return nil, fmt.Errorf("Couldn't pretend to be a user: %v", err)
		} else if u == nil {
			return nil, fmt.Errorf("Couldn't find user we are pretending to be")
		}
	}

	if err := LoadUserGroupIds(db, u); err != nil {
		return nil, fmt.Errorf("Couldn't load group membership: %v", err)
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
		userPtr = NewCurrentUser()
	}
	return
}

// LoadUpdateCount returns the number of not seen updates the given user has.
func LoadUpdateCount(db *database.DB, userId string) (int, error) {
	editTypes := []string{PageEditUpdateType}

	var editUpdateCount int
	row := database.NewQuery(`
		SELECT COUNT(DISTINCT type, subscribedToId, byUserId)
		FROM updates
		WHERE NOT seen AND userId=?`, userId).Add(`
			AND type IN`).AddArgsGroupStr(editTypes).ToStatement(db).QueryRow()
	_, err := row.Scan(&editUpdateCount)
	if err != nil {
		return -1, err
	}

	var nonEditUpdateCount int
	row = database.NewQuery(`
		SELECT COUNT(*)
		FROM updates
		WHERE NOT seen AND userId=?`, userId).Add(`
			AND type NOT IN`).AddArgsGroupStr(editTypes).ToStatement(db).QueryRow()
	_, err = row.Scan(&nonEditUpdateCount)
	if err != nil {
		return -1, err
	}

	return editUpdateCount + nonEditUpdateCount, err
}

// Load the number of new achievements for this user
func LoadNewAchievementCount(db *database.DB, user *CurrentUser) (int, error) {
	lastAchievementsView, err := LoadLastView(db, user, LastAchievementsModeView)
	if err != nil {
		return -1, err
	}

	var newLikeCount int
	row := database.NewQuery(`
		SELECT COUNT(*)
		FROM `).AddPart(PageInfosTable(user)).Add(` AS pi
		JOIN likes AS l
		ON pi.likeableId=l.likeableId
		JOIN users AS u
		ON l.userId=u.id
		WHERE pi.createdBy=?`, user.Id).Add(`
			AND l.userId!=?`, user.Id).Add(`
			AND l.value=1
			AND l.updatedAt>?`, lastAchievementsView).ToStatement(db).QueryRow()
	_, err = row.Scan(&newLikeCount)
	if err != nil {
		return -1, err
	}

	var newChangeLogLikeCount int
	row = database.NewQuery(`
		SELECT COUNT(*)
		FROM likes as l
		JOIN changeLogs as cl
		ON cl.likeableId=l.likeableId
		WHERE cl.userId=?`, user.Id).Add(`
			AND l.value=1 AND l.userId!=?`, user.Id).Add(`
			AND cl.type=?`, NewEditChangeLog).Add(`
			AND l.updatedAt>?`, lastAchievementsView).ToStatement(db).QueryRow()
	_, err = row.Scan(&newChangeLogLikeCount)
	if err != nil {
		return -1, err
	}

	var newTaughtCount int
	row = database.NewQuery(`
		SELECT COUNT(*)
		FROM userMasteryPairs AS ump
		JOIN `).AddPart(PageInfosTable(user)).Add(` AS pi
		ON ump.taughtBy=pi.pageId
		JOIN users AS u
		ON ump.userId=u.id
		WHERE pi.createdBy=?`, user.Id).Add(`
			AND ump.has=1 AND ump.userId!=?`, user.Id).Add(`
			AND ump.updatedAt>?`, lastAchievementsView).ToStatement(db).QueryRow()
	_, err = row.Scan(&newTaughtCount)
	if err != nil {
		return -1, err
	}

	newAchievementUpdateCount, err := LoadAchievementUpdateCount(db, user.Id, false)
	if err != nil {
		return -1, err
	}

	return newLikeCount + newTaughtCount + newChangeLogLikeCount + newAchievementUpdateCount, nil
}

func LoadNotificationCount(db *database.DB, userId string, includeOldAndDismissed bool) (int, error) {
	return loadUpdateCountInternal(db, userId, GetNotificationUpdateTypes(), includeOldAndDismissed)
}

func LoadMaintenanceUpdateCount(db *database.DB, userId string, includeOldAndDismissed bool) (int, error) {
	return loadUpdateCountInternal(db, userId, GetMaintenanceUpdateTypes(), includeOldAndDismissed)
}

func LoadAchievementUpdateCount(db *database.DB, userId string, includeOldAndDismissed bool) (int, error) {
	return loadUpdateCountInternal(db, userId, GetAchievementUpdateTypes(), includeOldAndDismissed)
}

func loadUpdateCountInternal(db *database.DB, userId string, updateTypes []string, includeOldAndDismissed bool) (int, error) {
	var filterCondition string
	if includeOldAndDismissed {
		filterCondition = "true"
	} else {
		filterCondition = "NOT seen AND NOT dismissed"
	}

	var updateCount int
	row := database.NewQuery(`
		SELECT COUNT(*)
		FROM updates
		WHERE userId=?`, userId).Add(`
			AND type IN`).AddArgsGroupStr(updateTypes).Add(`
			AND`).Add(filterCondition).ToStatement(db).QueryRow()
	_, err := row.Scan(&updateCount)
	if err != nil {
		return -1, err
	}

	return updateCount, err
}

// LoadUserTrust returns the trust that the user has in all domains.
func LoadUserTrust(db *database.DB, u *CurrentUser, domainIds []string) error {
	for _, domainId := range domainIds {
		u.TrustMap[domainId] = &Trust{}
	}

	if u.Id != "" {
		// TODO: load all sources of trust

		// NOTE: this should come last in computing trust, so that the bonus trust from
		// an invite slowly goes away as the user accumulates real trust.
		// Compute trust from invites
		wherePart := database.NewQuery(`WHERE toUserId=?`, u.Id)
		invites, err := LoadInvitesWhere(db, wherePart)
		if err != nil {
			return fmt.Errorf("Couldn't process existing invites: %v", err)
		}
		for _, invite := range invites {
			trust := u.TrustMap[invite.DomainId]
			if trust.EditTrust < DefaultInviteKarma {
				trust.EditTrust = DefaultInviteKarma
			}
			trust.Permissions.DomainAccess.Has = true
			u.IsDomainMember = true
		}

		// Load whether the user has ever had any maintenance updates
		hasReceivedMaintenanceUpdates, err := LoadHasReceivedMaintenanceUpdates(db, u)
		if err != nil {
			return fmt.Errorf("Couldn't process maintenance updates: %v", err)
		}
		u.HasReceivedMaintenanceUpdates = hasReceivedMaintenanceUpdates

		// Load whether the user has ever had any notifications
		hasReceivedNotifications, err := LoadHasReceivedNotifications(db, u)
		if err != nil {
			return fmt.Errorf("Couldn't process notifications: %v", err)
		}
		u.HasReceivedNotifications = hasReceivedNotifications
	}

	// Now compute permissions
	for _, trust := range u.TrustMap {
		if !trust.Permissions.DomainAccess.Has {
			trust.Permissions.DomainAccess.Reason = "You don't have access to this domain"
		}
		trust.Permissions.DomainTrust.Has = u.IsTrusted
		if !trust.Permissions.DomainTrust.Has {
			trust.Permissions.DomainTrust.Reason = "You don't have full trust for this domain"
		}
		trust.Permissions.Edit.Has = trust.EditTrust >= EditPageKarmaReq
		if !trust.Permissions.Edit.Has {
			trust.Permissions.Edit.Reason = "Not enough reputation"
		}
		trust.Permissions.Delete.Has = trust.EditTrust >= DeletePageKarmaReq
		if !trust.Permissions.Delete.Has {
			trust.Permissions.Delete.Reason = "Not enough reputation"
		}
		trust.Permissions.Comment.Has = trust.EditTrust >= CommentKarmaReq
		if !trust.Permissions.Comment.Has {
			trust.Permissions.Comment.Reason = "Not enough reputation"
		}
	}

	return nil
}

// Get invites filtered by the given condition.
func LoadInvitesWhere(db *database.DB, wherePart *database.QueryPart) ([]*Invite, error) {
	invites := make([]*Invite, 0)
	rows := database.NewQuery(`
		SELECT fromUserId,domainId,toEmail,createdAt,toUserId,claimedAt,emailSentAt
		FROM invites`).AddPart(wherePart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		invite := &Invite{}
		err := rows.Scan(&invite.FromUserId, &invite.DomainId, &invite.ToEmail,
			&invite.CreatedAt, &invite.ToUserId, &invite.ClaimedAt, &invite.EmailSentAt)
		if err != nil {
			return fmt.Errorf("failed to scan an invite: %v", err)
		}
		invites = append(invites, invite)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error while loading invites WHERE %v: %v", wherePart, err)
	}
	return invites, nil
}

func LoadHasReceivedMaintenanceUpdates(db *database.DB, u *CurrentUser) (bool, error) {
	lifetimeMaintenanceUpdateCount, err := LoadMaintenanceUpdateCount(db, u.Id, true)
	if err != nil {
		return false, fmt.Errorf("Error while retrieving maintenance update count: %v", err)
	}

	return lifetimeMaintenanceUpdateCount > 0, nil
}

func LoadHasReceivedNotifications(db *database.DB, u *CurrentUser) (bool, error) {
	lifetimeNotificationCount, err := LoadNotificationCount(db, u.Id, true)
	if err != nil {
		return false, fmt.Errorf("Error while retrieving notification count: %v", err)
	}

	return lifetimeNotificationCount > 0, nil
}

func init() {
	// Need to register this so it can be stored inside a cookie
	gob.Register(&CookieSession{})
}
