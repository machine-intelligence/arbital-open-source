// emailUpdatesTask.go sends an email to every user with the new updates
package tasks

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"

	"appengine/mail"
	"appengine/urlfetch"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

const (
	// Need at least this many new updates to send an email
	emailUpdateThreshold = 3
	emailUpdatesPeriod   = 60 * 60 * 24
)

// EmailUpdatesTask is the object that's put into the daemon queue.
type EmailUpdatesTask struct {
}

// Check if this task is valid, and we can safely execute it.
func (task *EmailUpdatesTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *EmailUpdatesTask) Execute(c sessions.Context) (delay int, err error) {
	delay = emailUpdatesPeriod
	if err = task.IsValid(); err != nil {
		return
	}

	c.Debugf("==== EMAIL UPDATES START ====")
	defer c.Debugf("==== EMAIL UPDATES COMPLETED SUCCESSFULLY ====")

	// Find all users who need emailing.
	err = database.QuerySql(c, `
		SELECT id,email
		FROM users
		WHERE DATEDIFF(NOW(),updateEmailSentAt)>=1`, emailUpdatesProcessUser)
	if err != nil {
		c.Debugf("ERROR: %v", err)
	}
	return
}

func emailUpdatesProcessUser(c sessions.Context, rows *sql.Rows) error {
	var userId int64
	var userEmail string
	err := rows.Scan(&userId, &userEmail)
	if err != nil {
		return fmt.Errorf("failed to scan a user id: %v", err)
	}

	// Update database first, even though we might fail to send the email. This
	// way we definitely won't accidentally email a person twice.
	query := fmt.Sprintf(`
		UPDATE users
		SET updateEmailSentAt=NOW()
		WHERE id=%d`, userId)
	_, err = database.ExecuteSql(c, query)
	if err != nil {
		return fmt.Errorf("failed to update updateEmailSentAt: %v", err)
	}

	// Data used for populating email template
	data := struct {
		UpdateCount  int
		UpdateGroups []*core.UpdateGroup
	}{
		UpdateCount:  0,
		UpdateGroups: nil,
	}

	// Load updates and populate the maps
	pageMap := make(map[int64]*core.Page)
	userMap := make(map[int64]*core.User)
	updateRows, err := core.LoadUpdateRows(c, userId, pageMap, userMap, true)
	if err != nil {
		return fmt.Errorf("failed to load updates: %v", err)
	}

	// Check to make sure there are enough updates
	data.UpdateCount = len(updateRows)
	if data.UpdateCount < emailUpdateThreshold {
		return nil
	}
	data.UpdateGroups = core.ConvertUpdateRowsToGroups(updateRows, nil)

	// Load pages.
	err = core.LoadPages(c, pageMap, userId, core.LoadPageOptions{})
	if err != nil {
		return fmt.Errorf("error while loading pages: %v", err)
	}

	// Load the names for all users.
	userMap[userId] = &core.User{Id: userId}
	for _, p := range pageMap {
		userMap[p.CreatorId] = &core.User{Id: p.CreatorId}
	}
	err = core.LoadUsers(c, userMap)
	if err != nil {
		return fmt.Errorf("error while loading users: %v", err)
	}

	// Load the template file
	var templateBytes []byte
	if sessions.Live {
		resp, err := urlfetch.Client(c).Get("http://zanaduu3.appspot.com/static/updatesEmailInlined.html")
		if err != nil {
			return fmt.Errorf("Couldn't load the email template form URL: %v", err)
		}
		defer resp.Body.Close()
		templateBytes, err = ioutil.ReadAll(resp.Body)
	} else {
		templateBytes, err = ioutil.ReadFile("../site/static/updatesEmailInlined.html")
	}
	if err != nil {
		return fmt.Errorf("Couldn't load the email template from file: %v", err)
	}

	funcMap := template.FuncMap{
		//"UserFirstName": func() int64 { return u.Id },
		"GetUserUrl": func(userId int64) string {
			return fmt.Sprintf(`http://zanaduu3.appspot.com/filter?user=%d`, userId)
		},
		"GetUserName": func(userId int64) string {
			return userMap[userId].FullName()
		},
		"GetPageUrl": func(pageId int64) string {
			return fmt.Sprintf("http://zanaduu3.appspot.com/pages/%d", pageId)
		},
		"GetPageTitle": func(pageId int64) string {
			return pageMap[pageId].Title
		},
	}

	// Create and execute template
	buffer := &bytes.Buffer{}
	t, err := template.New("email").Funcs(funcMap).Parse(string(templateBytes))
	if err != nil {
		return fmt.Errorf("Couldn't create template: %v", err)
	}
	err = t.Execute(buffer, data)
	if err != nil {
		return fmt.Errorf("Couldn't execute template: %v", err)
	}

	// Create mail message
	subject := fmt.Sprintf("%d new updates on Zanaduu", data.UpdateCount)
	msg := &mail.Message{
		Sender:   "Zanaduu <updates@zanaduu3.appspotmail.com>",
		To:       []string{userEmail},
		Subject:  subject,
		HTMLBody: buffer.String(),
	}

	// Ship it!
	err = mail.Send(c, msg)
	if err != nil {
		c.Inc("email_send_fail")
		return fmt.Errorf("Couldn't send email: %v", err)
	}
	c.Inc("email_send_success")
	c.Debugf("Email sent!")

	return nil
}
