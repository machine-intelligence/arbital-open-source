// refer_email.go handles the task of sending a referral email for the daemon queue.
package tasks

import (
	"fmt"
	"html"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"appengine/mail"
	"appengine/urlfetch"

	"xelaie/src/go/database"
	"xelaie/src/go/sessions"
)

// ReferEmailtask is the object that's put into the daemon queue.
type ReferEmailTask struct {
	FromId         int64  `json:",string"` // Twitter id
	FromScreenName string // Twitter @handle
	FromName       string // Normal name the user enters
	ToName         string // Normal name the user enters
	ToEmail        string // Email of the recepient
}

// Check if this task is valid, and we can safely execute it.
func (task *ReferEmailTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task *ReferEmailTask) Execute(c sessions.Context) (int, error) {
	if err := task.IsValid(); err != nil {
		return -1, fmt.Errorf("This task is not valid")
	}

	rand.Seed(time.Now().UnixNano())
	referralId := rand.Int63()
	referral := make(database.InsertMap)
	referral["referralId"] = referralId
	referral["fromUserId"] = task.FromId
	referral["fromScreenName"] = task.FromScreenName
	referral["toEmail"] = task.ToEmail
	referral["toName"] = task.ToName
	referral["fromName"] = task.FromName
	referral["createdAt"] = database.Now()

	err := database.ExecuteSql(c, database.GetInsertSql("referrals_sent", referral))
	if err != nil {
		return 60, fmt.Errorf("Couldn't add the referral to the db: %v", err)
	}

	var bytes []byte
	if sessions.Live {
		resp, err := urlfetch.Client(c).Get("http://rewards.xelaie.com/static/email_template_inlined.html")
		if err != nil {
			return 60, fmt.Errorf("Couldn't load the email template form URL: %v", err)
		}
		defer resp.Body.Close()
		bytes, err = ioutil.ReadAll(resp.Body)
	} else {
		bytes, err = ioutil.ReadFile("../site/static/email_template_inlined.html")
		if err != nil {
			return 60, fmt.Errorf("Couldn't load the email template from file: %v", err)
		}
	}
	emailTemplate := string(bytes)
	body := strings.Replace(emailTemplate, "{{FromName}}", task.FromName, -1)
	body = strings.Replace(body, "{{ToName}}", task.ToName, -1)
	body = strings.Replace(body, "{{ReferralId}}", fmt.Sprintf("%d", referralId), -1)
	subject := fmt.Sprintf("%s invites you to join Xelaie for a chance to win a $5 gift card", html.UnescapeString(task.FromName))
	msg := &mail.Message{
		Sender:   "Xelaie Rewards <referral@xelaie.com>",
		To:       []string{task.ToEmail},
		Subject:  subject,
		HTMLBody: body,
	}

	err = mail.Send(c, msg)
	if err != nil {
		c.Inc("email_send_fail")
		c.Errorf("Couldn't send email: %v", err)
		return 600, err
	}
	c.Inc("email_send_success")
	c.Debugf("Email sent!")

	return 0, nil
}
