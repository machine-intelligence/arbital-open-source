// task.go defines a task that can go into a queue
package tasks

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"google.golang.org/appengine/taskqueue"

	"zanaduu3/src/database"
	"zanaduu3/src/sessions"
)

// QueueTask is the object that's put into the daemon queue.
type QueueTask interface {
	Tag() string
	IsValid() error
	// If the int returned is:
	// > 0:  the task will be put back into the queue with the given number of seconds
	// == 0: the task will be deleted from the queue
	// < 0:  the task will remain leased for the default period of time
	Execute(db *database.DB) (int, error)
}

// TaskOptions specify how to add a task to the queue.
type TaskOptions struct {
	Name  string
	Delay int // seconds
}

// Add the task to the queue.
func Enqueue(c sessions.Context, task QueueTask, options *TaskOptions) error {
	if options == nil {
		options = &TaskOptions{}
	}
	if err := task.IsValid(); err != nil {
		return fmt.Errorf("Attempting to enqueue invalid task: %v", err)
	}
	buffer := new(bytes.Buffer)
	err := gob.NewEncoder(buffer).Encode(task)
	if err != nil {
		return fmt.Errorf("Couldn't encode the task.")
	}
	newTask := &taskqueue.Task{
		Method:  "PULL",
		Payload: buffer.Bytes(),
		Tag:     task.Tag(),
		Delay:   time.Duration(options.Delay) * time.Second,
	}
	if options.Name != "" {
		newTask.Name = options.Name
	}

	newTask, err = taskqueue.Add(c, newTask, "daemonQueue")
	if err != nil {
		return fmt.Errorf("Failed to insert task: %v", err)
	}
	return nil
}

// Convert byte stream into a QueueTask.
func Decode(leasedTask *taskqueue.Task, task QueueTask) (err error) {
	buffer := bytes.NewBuffer(leasedTask.Payload)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(task)
	if err != nil {
		err = fmt.Errorf("Couldn't decode a task: %v", err)
		return
	}
	if err = task.IsValid(); err != nil {
		err = fmt.Errorf("Attempting to decode invalid task: %v", err)
		return
	}
	return
}
