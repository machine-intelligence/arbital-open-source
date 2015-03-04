// init.go contains all the structs for the tasks that go into the queue.
package tasks

import (
	"encoding/gob"
)

func init() {
	gob.Register(&TickTask{})
	gob.Register(&NewUpdateTask{})
}
