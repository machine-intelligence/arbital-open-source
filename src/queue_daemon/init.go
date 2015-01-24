// init.go runs the daemon to process the daemon-queue.
package queue_daemon

import (
	"appengine/taskqueue"
	"fmt"
	"net/http"
	"time"

	"zanaduu3/src/sessions"
	"zanaduu3/src/tasks"
)

const (
	// TODO: read from queue.yaml config
	// TODO: refactor everything to use this variable; right now it's hardcoded in a bunch of places
	daemonQueueName = "daemon-queue"
	// Seconds between polling the queue for new tasks
	pollPeriod = 60
)

// processTask leases one task from the daemon-queue and processes it.
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
	if leasedTask.Tag == "tick" {
		task = &tasks.TickTask{}
	} else {
		return fmt.Errorf("Unknown tag for the task: %s", leasedTask.Tag)
	}

	err = tasks.Dequeue(leasedTask, task)
	if err != nil {
		taskqueue.Delete(c, leasedTask, daemonQueueName)
		return fmt.Errorf("Couldn't decode a task: %v", err)
	}
	c.Debugf("Decoded a task: %v", task)
	var retValue int
	retValue, err = task.Execute(c)
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
	err := tasks.EnqueueWithName(c, tasks.TickTask{}, "tick", "tick")
	if err != nil {
		c.Debugf("TickTask enqueue error: %v", err)
	}

	for true {
		if err := processTask(c); err != nil {
			c.Errorf("%v", err)
		}
	}
}

func init() {
	http.HandleFunc("/_ah/start", handler)
}
