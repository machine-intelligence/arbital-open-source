// task.go defines a task that can go into a queue
package tasks

import (
	"appengine/taskqueue"
	"bytes"
	"encoding/gob"
	"fmt"

	"zanaduu3/src/sessions"
)

// QueueTask is the object that's put into the daemon queue.
type QueueTask interface {
	IsValid() error
	// If the int returned is:
	// > 0:  the task will be put back into the queue with the given number of seconds
	// == 0: the task will be deleted from the queue
	// < 0:  the task will remain leased for the default period of time
	Execute(c sessions.Context) (int, error)
}

// Add the task to the queue.
func Enqueue(c sessions.Context, task interface{}, tag string) (err error) {
	buffer := new(bytes.Buffer)
	err = gob.NewEncoder(buffer).Encode(task)
	if err != nil {
		err = fmt.Errorf("Couldn't encode the task.")
		return
	}
	newTask := &taskqueue.Task{
		Method:  "PULL",
		Payload: buffer.Bytes(),
		Tag:     tag}
	newTask, err = taskqueue.Add(c, newTask, "daemon-queue")
	if err != nil {
		err = fmt.Errorf("Failed to insert task: %v", err)
		return
	}
	return
}

// Convert byte stream into a QueueTask.
func Dequeue(leasedTask *taskqueue.Task, task interface{}) (err error) {
	buffer := bytes.NewBuffer(leasedTask.Payload)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(task)
	if err != nil {
		err = fmt.Errorf("Couldn't decode a task: %v", err)
		return
	}
	return
}
