// Package user manages information about the current user.
package core

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

const (
	DailyEmailFrequency       = "daily"
	WeeklyEmailFrequency      = "weekly"
	NeverEmailFrequency       = "never"
	ImmediatelyEmailFrequency = "immediately"

	DefaultEmailFrequency = DailyEmailFrequency
	DefaultEmailThreshold = 3
)

const (
	sessionKey = "arbitalSession2" // key for session storage
)

// User holds information about the current user of the app.
// Note: this structure is also stored in a cookie.
type CurrentUser struct {
	User

	// DB variables
	FbUserID               string `json:"fbUserId"`
	Email                  string `json:"email"`
	IsAdmin                bool   `json:"isAdmin"`
	EmailFrequency         string `json:"emailFrequency"`
	EmailThreshold         int    `json:"emailThreshold"`
	IgnoreMathjax          bool   `json:"ignoreMathjax"`
	ShowAdvancedEditorMode bool   `json:"showAdvancedEditorMode"`
	IsSlackMember          bool   `json:"isSlackMember"`

	// If the user isn't logged in, this is set to their unique session id
	SessionID string `json:"-"`
	// The analytics id is a hash of the session id
	AnalyticsID string `json:"analyticsId"`

	// Computed variables
	HasReceivedMaintenanceUpdates bool      `json:"hasReceivedMaintenanceUpdates"`
	HasReceivedNotifications      bool      `json:"hasReceivedNotifications"`
	NewNotificationCount          int       `json:"newNotificationCount"`
	NewAchievementCount           int       `json:"newAchievementCount"`
	MaintenanceUpdateCount        int       `json:"maintenanceUpdateCount"`
	InvitesClaimed                []*Invite `json:"invitesClaimed"`
	// If set, these are the lists the user is subscribed to via mailchimp
	MailchimpInterests map[string]bool `json:"mailchimpInterests"`

	// Hacky way to see if the user is in the process of reading any of the arcs
	// These are only loaded when the user is on the home page
	ContinueBayesPath *PathInstance `json:"continueBayesPath"`
	ContinueLogPath   *PathInstance `json:"continueLogPath"`
}

// Invite is an invitation from a trusted user to another to participate in a domain.
type Invite struct {
	FromUserID  string `json:"fromUserId"`
	DomainID    string `json:"domainId"`
	Role        string `json:"role"`
	ToEmail     string `json:"toEmail"`
	CreatedAt   string `json:"createdAt"`
	ToUserID    string `json:"toUserId"`
	ClaimedAt   string `json:"claimedAt"`
	EmailSentAt string `json:"emailSentAt"`
}

type CookieSession struct {
	Email       string
	SessionID   string
	AnalyticsID string

	// Randomly generated salt (for security/encryption reasons)
	Random string
}

func NewCurrentUser() *CurrentUser {
	var u CurrentUser
	u.User = *NewUser()
	u.InvitesClaimed = make([]*Invite, 0)
	u.MailchimpInterests = make(map[string]bool)
	return &u
}

func (user *CurrentUser) FullName() string {
	return user.FirstName + " " + user.LastName
}

// GetSomeId returns user's id or, if not available, session id, which could still be ""
func (user *CurrentUser) GetSomeID() string {
	if user.ID != "" {
		return user.ID
	}
	return user.SessionID
}

// Store a unique session id (and email if they're logged in) in a cookie so we can track the user's session
func SaveCookie(w http.ResponseWriter, r *http.Request, email string) (string, string, error) {
	s, err := sessions.GetSession(r)
	if err != nil {
		return "", "", fmt.Errorf("Couldn't get session: %v", err)
	}

	randString := func() (string, error) {
		b := make([]byte, 30)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(b), nil
	}
	r1, err1 := randString()
	r2, err2 := randString()
	if err1 != nil || err2 != nil {
		return "", "", errors.New("Failed to read random device")
	}
	sessionID := "sid:" + r1
	hashedSessionId := sha256.Sum256([]byte(r1))
	analyticsID := "aid:" + base64.RawStdEncoding.EncodeToString(hashedSessionId[:])

	s.Values[sessionKey] = &CookieSession{
		Email:       email,
		SessionID:   sessionID,
		AnalyticsID: analyticsID,
		Random:      r2,
	}
	err = s.Save(r, w)
	if err != nil {
		return "", "", fmt.Errorf("Failed to save user to session: %v", err)
	}

	return sessionID, analyticsID, nil
}

// Load user by id. If u object is given, load data into it. Otherwise create a new user object.
func LoadCurrentUserFromDb(db *database.DB, userID string, u *CurrentUser) (*CurrentUser, error) {
	if u == nil {
		u = NewCurrentUser()
	}
	row := db.NewStatement(`
		SELECT id,fbUserId,email,firstName,lastName,isAdmin,isSlackMember,
			emailFrequency,emailThreshold,ignoreMathjax,showAdvancedEditorMode
		FROM users
		WHERE id=?`).QueryRow(userID)
	exists, err := row.Scan(&u.ID, &u.FbUserID, &u.Email, &u.FirstName, &u.LastName,
		&u.IsAdmin, &u.IsSlackMember, &u.EmailFrequency, &u.EmailThreshold, &u.IgnoreMathjax,
		&u.ShowAdvancedEditorMode)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load user: %v", err)
	} else if !exists {
		return nil, fmt.Errorf("Couldn't find the user")
	}
	return u, nil
}

// LoadCurrentUser loads the user by their email via the cookie.
func LoadCurrentUser(w http.ResponseWriter, r *http.Request, db *database.DB) (userPtr *CurrentUser, err error) {
	// Load email from the cookie
	s, err := sessions.GetSession(r)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get session: %v", err)
	}
	u := NewCurrentUser()

	var cookie *CookieSession
	if cookieStruct, ok := s.Values[sessionKey]; !ok {
		sessionID, analyticsID, err := SaveCookie(w, r, "")
		u.SessionID = sessionID
		u.AnalyticsID = analyticsID
		return u, err
	} else {
		cookie = cookieStruct.(*CookieSession)
	}
	u.SessionID = cookie.SessionID
	u.AnalyticsID = cookie.AnalyticsID
	if cookie.Email == "" {
		return u, err
	}

	var pretendToBeUserID string
	row := db.NewStatement(`
		SELECT id,pretendToBeUserId,fbUserId,email,firstName,lastName,isAdmin,
			isSlackMember,emailFrequency,emailThreshold,ignoreMathjax,showAdvancedEditorMode
		FROM users
		WHERE email=?`).QueryRow(cookie.Email)
	exists, err := row.Scan(&u.ID, &pretendToBeUserID, &u.FbUserID, &u.Email, &u.FirstName, &u.LastName,
		&u.IsAdmin, &u.IsSlackMember, &u.EmailFrequency, &u.EmailThreshold, &u.IgnoreMathjax,
		&u.ShowAdvancedEditorMode)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve a user: %v", err)
	} else if !exists {
		return nil, fmt.Errorf("Couldn't find that email in DB")
	}

	// Admins can pretened to be certain users
	if u.IsAdmin && pretendToBeUserID != "" {
		u, err = LoadCurrentUserFromDb(db, pretendToBeUserID, u)
		if err != nil {
			return nil, fmt.Errorf("Couldn't pretend to be a user: %v", err)
		} else if u == nil {
			return nil, fmt.Errorf("Couldn't find user we are pretending to be")
		}
	}

	if err := LoadUserDomainMembership(db, &u.User, nil); err != nil {
		return nil, fmt.Errorf("Couldn't load group membership: %v", err)
	}

	return u, nil
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
		FROM pageInfos AS pi
		JOIN likes AS l
		ON pi.likeableId=l.likeableId
		JOIN users AS u
		ON l.userId=u.id
		WHERE`).AddPart(PageInfosFilter(user)).Add(`
			AND pi.createdBy=?`, user.ID).Add(`
			AND l.userId!=?`, user.ID).Add(`
			AND l.value=1
			AND l.updatedAt>?`, lastAchievementsView).ToStatement(db).QueryRow()
	_, err = row.Scan(&newLikeCount)
	if err != nil {
		return -1, err
	}

	var newChangeLogLikeCount int
	row = database.NewQuery(`
		SELECT COUNT(*)
		FROM likes AS l
		JOIN changeLogs AS cl
		ON cl.likeableId=l.likeableId
		WHERE cl.userId=?`, user.ID).Add(`
			AND l.value=1 AND l.userId!=?`, user.ID).Add(`
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
		JOIN pageInfos AS pi
		ON ump.taughtBy=pi.pageId
		JOIN users AS u
		ON ump.userId=u.id
		WHERE`).AddPart(PageInfosFilter(user)).Add(`
			AND pi.createdBy=?`, user.ID).Add(`
			AND ump.has=1 AND ump.userId!=?`, user.ID).Add(`
			AND ump.updatedAt>?`, lastAchievementsView).ToStatement(db).QueryRow()
	_, err = row.Scan(&newTaughtCount)
	if err != nil {
		return -1, err
	}

	newAchievementUpdateCount, err := LoadAchievementUpdateCount(db, user.ID, false)
	if err != nil {
		return -1, err
	}

	return newLikeCount + newTaughtCount + newChangeLogLikeCount + newAchievementUpdateCount, nil
}

func LoadNotificationCount(db *database.DB, userID string, includeOldAndDismissed bool) (int, error) {
	return loadUpdateCountInternal(db, userID, GetNotificationUpdateTypes(), includeOldAndDismissed)
}

func LoadMaintenanceUpdateCount(db *database.DB, userID string, includeOldAndDismissed bool) (int, error) {
	return loadUpdateCountInternal(db, userID, GetMaintenanceUpdateTypes(), includeOldAndDismissed)
}

func LoadAchievementUpdateCount(db *database.DB, userID string, includeOldAndDismissed bool) (int, error) {
	return loadUpdateCountInternal(db, userID, GetAchievementUpdateTypes(), includeOldAndDismissed)
}

func loadUpdateCountInternal(db *database.DB, userID string, updateTypes []string, includeOldAndDismissed bool) (int, error) {
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
		WHERE userId=?`, userID).Add(`
			AND type IN`).AddArgsGroupStr(updateTypes).Add(`
			AND`).Add(filterCondition).ToStatement(db).QueryRow()
	_, err := row.Scan(&updateCount)
	if err != nil {
		return -1, err
	}

	return updateCount, err
}

// Get invites filtered by the given condition.
func LoadInvitesWhere(db *database.DB, wherePart *database.QueryPart) ([]*Invite, error) {
	invites := make([]*Invite, 0)
	rows := database.NewQuery(`
		SELECT fromUserId,domainId,role,toEmail,createdAt,toUserId,claimedAt,emailSentAt
		FROM invites`).AddPart(wherePart).ToStatement(db).Query()
	err := rows.Process(func(db *database.DB, rows *database.Rows) error {
		invite := &Invite{}
		err := rows.Scan(&invite.FromUserID, &invite.DomainID, &invite.Role, &invite.ToEmail,
			&invite.CreatedAt, &invite.ToUserID, &invite.ClaimedAt, &invite.EmailSentAt)
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
	lifetimeMaintenanceUpdateCount, err := LoadMaintenanceUpdateCount(db, u.ID, true)
	if err != nil {
		return false, fmt.Errorf("Error while retrieving maintenance update count: %v", err)
	}

	return lifetimeMaintenanceUpdateCount > 0, nil
}

func LoadHasReceivedNotifications(db *database.DB, u *CurrentUser) (bool, error) {
	lifetimeNotificationCount, err := LoadNotificationCount(db, u.ID, true)
	if err != nil {
		return false, fmt.Errorf("Error while retrieving notification count: %v", err)
	}

	return lifetimeNotificationCount > 0, nil
}

func init() {
	// Need to register this so it can be stored inside a cookie
	gob.Register(&CookieSession{})
}
