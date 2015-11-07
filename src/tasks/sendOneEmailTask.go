// sendOneEmailTask.go sends one email, with no delay before the task can run again
package tasks

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"

	"appengine/mail"
	"appengine/urlfetch"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/user"
)

// SendOneEmailTask is the object that's put into the daemon queue.
type SendOneEmailTask struct {
	UserId int64
}

// Check if this task is valid, and we can safely execute it.
func (task *SendOneEmailTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *SendOneEmailTask) Execute(db *database.DB) (delay int, err error) {
	delay = 0
	c := db.C

	if err = task.IsValid(); err != nil {
		return
	}

	c.Debugf("==== SEND EMAIL START ====")
	defer c.Debugf("==== SEND EMAIL COMPLETED SUCCESSFULLY ====")

	SendTheEmail(db, task.UserId)

	return
}

func SendTheEmail(db *database.DB, userId int64) (retErr error, resultData string) {
	c := db.C

	resultData = ""

	c.Debugf("starting sendTheEmail")

	u := &user.User{}
	row := db.NewStatement(`
		SELECT id,email,emailFrequency,emailThreshold
		FROM users
		WHERE id=?`).QueryRow(userId)
	_, err := row.Scan(&u.Id, &u.Email, &u.EmailFrequency, &u.EmailThreshold)
	if err != nil {
		return fmt.Errorf("Couldn't retrieve a user: %v", err), resultData
	}

	c.Debugf("u.Id: %v", u.Id)
	c.Debugf("u.Email: %v", u.Email)
	c.Debugf("u.EmailThreshold: %v", u.EmailThreshold)

	// Load the groups the user belongs to.
	if err = core.LoadUserGroupIds(db, u); err != nil {
		return fmt.Errorf("Couldn't load user groups: %v", err), resultData
	}

	// Update database first, even though we might fail to send the email. This
	// way we definitely won't accidentally email a person twice.
	statement := db.NewStatement(`
		UPDATE users
		SET updateEmailSentAt=NOW()
		WHERE id=?`)
	_, err = statement.Exec(u.Id)
	if err != nil {
		return fmt.Errorf("failed to update updateEmailSentAt: %v", err), resultData
	}

	c.Debugf("updated database")

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
	masteryMap := make(map[int64]*core.Mastery)
	updateRows, err := core.LoadUpdateRows(db, u.Id, pageMap, userMap, true)
	if err != nil {
		return fmt.Errorf("failed to load updates: %v", err), resultData
	}

	c.Debugf("loaded updates")
	/*
		c.Debugf("len(updateRows): %v", len(updateRows))
		c.Debugf("Data: %+v", data)
		c.Debugf("updateRows: %+v", updateRows)
		for key, value := range updateRows {
			c.Debugf("updateRows[%v]: %+v", key, value)
		}
		c.Debugf("pageMap: %+v", pageMap)
	*/

	// Check to make sure there are enough updates
	data.UpdateCount = len(updateRows)
	if data.UpdateCount < u.EmailThreshold {
		c.Debugf("no updates to email")
		return nil, resultData
	}
	data.UpdateGroups = core.ConvertUpdateRowsToGroups(updateRows, pageMap)
	// Mark loaded updates as emailed
	updateIds := make([]interface{}, 0)
	for _, row := range updateRows {
		updateIds = append(updateIds, row.Id)
	}
	statement = database.NewQuery(`
			UPDATE updates
			SET emailed=true
			WHERE id IN`).AddArgsGroup(updateIds).ToStatement(db)
	_, err = statement.Exec()
	if err != nil {
		return fmt.Errorf("Couldn't update updates an emailed: %v", err), resultData
	}
	c.Debugf("marked updates as emailed")

	// Load pages.
	err = core.ExecuteLoadPipeline(db, u, pageMap, userMap, masteryMap)
	if err != nil {
		return fmt.Errorf("Pipeline error: %v", err), resultData
	}

	// Load the template file
	var templateBytes []byte
	if sessions.Live {
		resp, err := urlfetch.Client(c).Get(fmt.Sprintf("%s/static/updatesEmailInlined.html", sessions.GetDomain()))
		if err != nil {
			return fmt.Errorf("Couldn't load the email template form URL: %v", err), resultData
		}
		defer resp.Body.Close()
		templateBytes, err = ioutil.ReadAll(resp.Body)
	} else {
		templateBytes, err = ioutil.ReadFile("../site/static/updatesEmailInlined.html")
	}
	if err != nil {
		return fmt.Errorf("Couldn't load the email template from file: %v", err), resultData
	}

	funcMap := template.FuncMap{
		//"UserFirstName": func() int64 { return u.Id },
		"GetUserUrl": func(userId int64) string {
			return fmt.Sprintf(`%s/user/%d`, sessions.GetDomainForTestEmail(), userId)
		},
		"GetUserName": func(userId int64) string {
			return userMap[userId].FullName()
		},
		"GetPageUrl": func(pageId int64) string {
			return fmt.Sprintf("%s/pages/%d", sessions.GetDomainForTestEmail(), pageId)
		},
		"GetPageTitle": func(pageId int64) string {
			return pageMap[pageId].Title
		},
	}

	// Create and execute template
	buffer := &bytes.Buffer{}
	t, err := template.New("email").Funcs(funcMap).Parse(string(templateBytes))
	if err != nil {
		return fmt.Errorf("Couldn't parse template: %v", err), resultData
	}
	err = t.Execute(buffer, data)
	if err != nil {
		return fmt.Errorf("Couldn't execute template: %v", err), resultData
	}

	c.Debugf("finished loading")

	if sessions.Live {

		c.Debugf("Sending email, live")

		// Create mail message
		subject := fmt.Sprintf("%d new updates on Arbital", data.UpdateCount)
		msg := &mail.Message{
			Sender:   "Arbital <updates@zanaduu3.appspotmail.com>",
			To:       []string{u.Email},
			Bcc:      []string{"alexei@arbital.com"},
			Subject:  subject,
			HTMLBody: buffer.String(),
		}

		// Ship it!
		err = mail.Send(c, msg)
		if err != nil {
			c.Inc("email_send_fail")
			return fmt.Errorf("Couldn't send email: %v", err), resultData
		}

	} else {
		// If not live, then write the email to an html file

		c.Debugf("Sending email, not live")
		/*
			c.Debugf("Data: %+v", data)
			c.Debugf("UpdateGroups: %+v", data.UpdateGroups)
			for key1, value := range data.UpdateGroups {
				c.Debugf("UpdateGroups[%v]: %+v", key1, value)
				c.Debugf("UpdateGroups[%v].Key: %+v", key1, value.Key)
				for key2, value := range value.Updates {
					c.Debugf("UpdateGroups[%v].Updates[%v]: %+v", key1, key2, value)
				}
			}
			c.Debugf("userMap: %+v", userMap)
			for key1, value := range userMap {
				c.Debugf("userMap[%v]: %+v", key1, value)
			}
			c.Debugf("pageMap: %+v", pageMap)
			for key1, value := range pageMap {
				c.Debugf("pageMap[%v]: %+v", key1, value)
			}
			//c.Debugf("sessions.GetDomain(): %v", sessions.GetDomain())
		*/

		resultData = buffer.String()

	}

	c.Inc("email_send_success")
	c.Debugf("Email sent!")

	return nil, resultData
}
