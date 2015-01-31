// newComment.go can change data for a comment.
package site

import (
	"encoding/json"
	"net/http"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
	"zanaduu3/src/user"
)

// newCommentData is the object that's put into the daemon queue.
type newCommentData struct {
	Text       string
	InputId    int64 `json:",string"`
	ReplyToId  int64 `json:",string"`
	QuestionId int64 `json:",string"`
}

// newCommentHandler renders the comment page.
func newCommentHandler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	var data newCommentData
	err := decoder.Decode(&data)
	if err != nil || data.Text == "" || data.QuestionId <= 0 {
		c.Inc("new_comment_fail")
		c.Errorf("Couldn't decode json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get user object
	var u *user.User
	u, err = user.LoadUser(w, r)
	if err != nil {
		c.Inc("new_comment_fail")
		c.Errorf("Couldn't load user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Add new comment
	hashmap := make(map[string]interface{})
	hashmap["inputId"] = data.InputId
	hashmap["createdAt"] = database.Now()
	hashmap["creatorId"] = u.Id
	hashmap["creatorName"] = u.FullName()
	hashmap["replyToId"] = data.ReplyToId
	hashmap["text"] = data.Text
	query := database.GetInsertSql("comments", hashmap)
	result, err2 := database.ExecuteSql(c, query)
	if err2 != nil {
		c.Inc("new_comment_fail")
		c.Errorf("Couldn't create new comment: %v", err2)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// If it's top level comment, subscribe the user to it.
	// If it's a reply to a comment, then subscribe the user to the parent.
	var subscribeCommentId int64
	if data.ReplyToId <= 0 {
		subscribeCommentId, _ = result.LastInsertId()
	} else {
		subscribeCommentId = data.ReplyToId
	}
	hashmap = make(map[string]interface{})
	hashmap["userId"] = u.Id
	hashmap["createdAt"] = database.Now()
	hashmap["commentId"] = subscribeCommentId
	// Note: if this subscription already exists, we update userId, which does nothing,
	// but also prevents an error from being generated.
	query = database.GetInsertSql("subscriptions", hashmap, "userId")
	if _, err = database.ExecuteSql(c, query); err != nil {
		c.Inc("new_subscription_fail")
		c.Errorf("Couldn't create new subscription: %v", err)
	}

	// Generate updates for people who are subscribed...
	var task tasks.NewUpdateTask
	task.UserId = u.Id
	task.QuestionId = data.QuestionId
	if data.ReplyToId <= 0 {
		// ... to this question.
		task.UpdateType = "topLevelComment"
	} else {
		// ... to the parent comment.
		task.CommentId = subscribeCommentId
		task.UpdateType = "reply"
	}
	if err := task.IsValid(); err != nil {
		c.Errorf("Invalid task created: %v", err)
	}
	if err := tasks.Enqueue(c, task, "newUpdate"); err != nil {
		c.Errorf("Couldn't enqueue a task: %v", err)
	}
}
