// init.go runs the daemon to process the daemonQueue.
package queue_daemon

import (
	"appengine/taskqueue"
	"fmt"
	"net/http"
	"time"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

const (
	// TODO: read from queue.yaml config
	// TODO: refactor everything to use this variable; right now it's hardcoded in a bunch of places
	daemonQueueName = "daemonQueue"
	// Seconds between polling the queue for new tasks
	pollPeriod = 1
)

// processTask leases one task from the daemonQueue and processes it.
func processTask(c sessions.Context) error {
	leasedTasks, err := taskqueue.Lease(c, 1, daemonQueueName, 3600)
	if err != nil {
		return fmt.Errorf("Couldn't lease a task: %v", err)
	}
	if len(leasedTasks) == 0 {
		time.Sleep(pollPeriod * time.Second)
		return nil
	}
	leasedTask := leasedTasks[0]

	var task tasks.QueueTask
	// TODO: refactor tag strings into the corresponding files as consts
	if leasedTask.Tag == (&tasks.TickTask{}).Tag() {
		task = &tasks.TickTask{}
	} else if leasedTask.Tag == (&tasks.PopulateElasticTask{}).Tag() {
		task = &tasks.PopulateElasticTask{}
	} else if leasedTask.Tag == (&tasks.UpdateElasticPageTask{}).Tag() {
		task = &tasks.UpdateElasticPageTask{}
	} else if leasedTask.Tag == (&tasks.NewUpdateTask{}).Tag() {
		task = &tasks.NewUpdateTask{}
	} else if leasedTask.Tag == (&tasks.AtMentionUpdateTask{}).Tag() {
		task = &tasks.AtMentionUpdateTask{}
	} else if leasedTask.Tag == (&tasks.MemberUpdateTask{}).Tag() {
		task = &tasks.MemberUpdateTask{}
	} else if leasedTask.Tag == (&tasks.EmailUpdatesTask{}).Tag() {
		task = &tasks.EmailUpdatesTask{}
	} else if leasedTask.Tag == (&tasks.ProcessMarkTask{}).Tag() {
		task = &tasks.ProcessMarkTask{}
	} else if leasedTask.Tag == (&tasks.SendOneEmailTask{}).Tag() {
		task = &tasks.SendOneEmailTask{}
	} else if leasedTask.Tag == (&tasks.SendFeedbackEmailTask{}).Tag() {
		task = &tasks.SendFeedbackEmailTask{}
	} else if leasedTask.Tag == (&tasks.PropagateDomainTask{}).Tag() {
		task = &tasks.PropagateDomainTask{}
	} else if leasedTask.Tag == (&tasks.UpdateMetadataTask{}).Tag() {
		task = &tasks.UpdateMetadataTask{}
	} else if leasedTask.Tag == (&tasks.FixTextTask{}).Tag() {
		task = &tasks.FixTextTask{}
	} else if leasedTask.Tag == (&tasks.ResetPasswordsTask{}).Tag() {
		task = &tasks.ResetPasswordsTask{}
	} else {
		return fmt.Errorf("Unknown tag for the task: %s", leasedTask.Tag)
	}

	err = tasks.Decode(leasedTask, task)
	if err != nil {
		taskqueue.Delete(c, leasedTask, daemonQueueName)
		return fmt.Errorf("Couldn't decode a task: %v", err)
	}
	c.Debugf("Decoded a task: %v", task)

	// Open DB connection
	db, err := database.GetDB(c)
	if err != nil {
		return fmt.Errorf("ERROR: %v", err)
	}

	// Execute the task
	var retValue int
	retValue, err = task.Execute(db)
	if err != nil {
		return fmt.Errorf("Couldn't execute a task: %v", err)
	}
	if retValue == 0 {
		err = taskqueue.Delete(c, leasedTask, daemonQueueName)
		if err != nil {
			return fmt.Errorf("Couldn't delete a task from the daemon queue: %v", err)
		}
	} else if retValue > 0 {
		err = taskqueue.ModifyLease(c, leasedTask, daemonQueueName, retValue)
		if err != nil {
			return fmt.Errorf("Couldn't modify lease for a task: %v", err)
		}
	}
	return nil
}

// handler handles the /_ah/start GET request and never leaves.
func handler(w http.ResponseWriter, r *http.Request) {
	c := sessions.NewContext(r)

	// Insert the first Tick task
	// TODO: catch errors, but ignore "already added" error
	/*var tickTask tasks.TickTask
	err := tasks.EnqueueWithName(c, &tickTask, "tick", "tick")
	if err != nil {
		c.Debugf("TickTask enqueue error: %v", err)
	}*/
	var emailUpdatesTask tasks.EmailUpdatesTask
	err := tasks.Enqueue(c, &emailUpdatesTask, &tasks.TaskOptions{Name: emailUpdatesTask.Tag()})
	if err != nil {
		c.Debugf("EmailUpdatesTask enqueue error: %v", err)
	}

	for true {
		if err := processTask(c); err != nil {
			c.Debugf("ERROR: %v", err)
			c.Errorf("%v", err)
		}
	}
}

func init() {
	http.HandleFunc("/_ah/start", handler)
}
