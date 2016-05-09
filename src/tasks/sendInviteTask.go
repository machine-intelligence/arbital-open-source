// sendInviteTask.go sends a invite email
package tasks

import (
	"fmt"

	"appengine/mail"

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// SendInviteTask is the object that's put into the daemon queue.
type SendInviteTask struct {
	FromUserId string
	DomainIds  []string
	ToEmail    string
}

func (task SendInviteTask) Tag() string {
	return "sendInvite"
}

// Check if this task is valid, and we can safely execute it.
func (task SendInviteTask) IsValid() error {
	if !core.IsIdValid(task.FromUserId) {
		return fmt.Errorf("Invalid FromUserId")
	}
	for _, domainId := range task.DomainIds {
		if !core.IsIdValid(domainId) {
			return fmt.Errorf("Invalid domain id: %v", domainId)
		}
	}
	if task.ToEmail == "" {
		return fmt.Errorf("Invalid ToEmail")
	}

	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task SendInviteTask) Execute(db *database.DB) (delay int, err error) {
	delay = 0
	c := db.C

	if err = task.IsValid(); err != nil {
		return
	}

	c.Debugf("==== SEND INVITE START ====")
	defer c.Debugf("==== SEND INVITE COMPLETED ====")

	senderUser, err := core.LoadUser(db, task.FromUserId, task.FromUserId)
	if err != nil {
		return -1, fmt.Errorf("Couldn't load sender user info: %v", err)
	}

	pageMap := make(map[string]*core.Page)
	for _, domainId := range task.DomainIds {
		core.AddPageIdToMap(domainId, pageMap)
	}
	err = core.LoadPages(db, &core.CurrentUser{Id: task.FromUserId}, pageMap)
	if err != nil {
		return -1, fmt.Errorf("Couldn't load domain info: %v", err)
	}

	domainsDesc := ""
	if len(pageMap) > 0 {
		for index, domainId := range task.DomainIds {
			if index > 0 {
				if index == len(task.DomainIds)-1 {
					domainsDesc += " and "
				} else {
					domainsDesc += ", "
				}
			}
			domainsDesc += pageMap[domainId].Title
		}
	}

	errMessage, err := db.Transaction(func(tx *database.Tx) (string, error) {
		bodyText := fmt.Sprintf(`
Hello!

%s just sent you an invite to Arbital, an ambitious effort to solve online discussion.

This invite gives you the permission to create and edit %s pages.

Visit https://arbital.com/signup to create your account.

We're excited to have you with us!

—Team Arbital`, senderUser.FullName(), domainsDesc)

		if sessions.Live {
			// Create mail message
			msg := &mail.Message{
				Sender:  "alexei@arbital.com",
				To:      []string{task.ToEmail},
				Bcc:     []string{"alexei@arbital.com"},
				Subject: fmt.Sprintf("Arbital invite from %s", senderUser.FullName()),
				Body:    bodyText,
			}

			// Ship it!
			err = mail.Send(c, msg)
			if err != nil {
				c.Inc("email_send_fail")
				return "Couldn't send email", err
			}
		} else {
			// If not live, then do nothing, for now
		}
		return "", nil
	})
	if errMessage != "" {
		return -1, fmt.Errorf(errMessage, err)
	}

	c.Inc("invite_send_success")
	c.Debugf("Invite sent!")

	return
}
