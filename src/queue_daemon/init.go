// init.go runs the daemon to process the daemonQueue.
package queue_daemon

import (
	"appengine/taskqueue"
	"fmt"
	"net/http"
	"reflect"
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

	// Go through all possible tasks and decode into correct type based on the tag string
	var task tasks.QueueTask
	taskPrototypes := []tasks.QueueTask{
		tasks.AtMentionUpdateTask{},
		tasks.CheckAnsweredMarksTask{},
		tasks.DomainWideNewUpdateTask{},
		tasks.EmailUpdatesTask{},
		tasks.FixTextTask{},
		tasks.MemberUpdateTask{},
		tasks.NewUpdateTask{},
		tasks.PopulateElasticTask{},
		tasks.PropagateDomainTask{},
		tasks.PublishPagePairTask{},
		tasks.ResetPasswordsTask{},
		tasks.SendFeedbackEmailTask{},
		tasks.SendInviteTask{},
		tasks.SendOneEmailTask{},
		tasks.TickTask{},
		tasks.UpdateElasticPageTask{},
		tasks.UpdateFeaturedPagesTask{},
		tasks.UpdateMetadataTask{},
		tasks.UpdatePagePairsTask{},
	}
	taskPrototypeMap := make(map[string]tasks.QueueTask)
	for _, prototype := range taskPrototypes {
		key := prototype.Tag()
		if _, ok := taskPrototypeMap[key]; !ok {
			taskPrototypeMap[key] = prototype
		} else {
			return fmt.Errorf("Some tasks are registering a duplicate tag: %s", prototype.Tag())
		}
	}
	if taskPrototype, ok := taskPrototypeMap[leasedTask.Tag]; ok {
		task = reflect.New(reflect.TypeOf(taskPrototype)).Interface().(tasks.QueueTask)
	} else {
		return fmt.Errorf("Unknown tag for the task: %s", leasedTask.Tag)
	}

	// Decode the task
	err = tasks.Decode(leasedTask, task)
	if err != nil {
		taskqueue.Delete(c, leasedTask, daemonQueueName)
		return fmt.Errorf("Couldn't decode a task: %v", err)
	}
	c.Debugf("Decoded a task: %+v", task)

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

	// Insert multiple tasks that need to be always running.
	// TODO: catch errors, but ignore "already added" error
	var tickTask tasks.TickTask
	err := tasks.Enqueue(c, &tickTask, &tasks.TaskOptions{Name: tickTask.Tag()})
	if err != nil {
		c.Debugf("TickTask enqueue error: %v", err)
	}
	var emailUpdatesTask tasks.EmailUpdatesTask
	err = tasks.Enqueue(c, &emailUpdatesTask, &tasks.TaskOptions{Name: emailUpdatesTask.Tag()})
	if err != nil {
		c.Debugf("EmailUpdatesTask enqueue error: %v", err)
	}
	var checkMarksTask tasks.CheckAnsweredMarksTask
	err = tasks.Enqueue(c, &checkMarksTask, &tasks.TaskOptions{Name: checkMarksTask.Tag()})
	if err != nil {
		c.Debugf("CheckAnsweredMarksTask enqueue error: %v", err)
	}
	var updateFeaturedPagesTask tasks.UpdateFeaturedPagesTask
	err = tasks.Enqueue(c, &updateFeaturedPagesTask, &tasks.TaskOptions{Name: updateFeaturedPagesTask.Tag()})
	if err != nil {
		c.Debugf("UpdateFeaturedPagesTask enqueue error: %v", err)
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
