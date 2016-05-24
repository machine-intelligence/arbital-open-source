// resetPasswordsTask.go adds all the pages to the elastic index.
package tasks

import (
	//"fmt"
	"math/rand"
	"time"

	"zanaduu3/src/database"
	//"zanaduu3/src/stormpath"
)

// ResetPasswordsTask is the object that's put into the daemon queue.
type ResetPasswordsTask struct {
}

func (task ResetPasswordsTask) Tag() string {
	return "resetPasswords"
}

// Check if this task is valid, and we can safely execute it.
func (task ResetPasswordsTask) IsValid() error {
	return nil
}

// Execute this task. Called by the actual daemon worker, don't call on BE.
// For comments on return value see tasks.QueueTask
func (task ResetPasswordsTask) Execute(db *database.DB) (delay int, err error) {
	c := db.C

	if err = task.IsValid(); err != nil {
		return 0, err
	}

	c.Infof("==== RESET PASSWORDS START ====")
	defer c.Infof("==== RESET PASSWORDS COMPLETED ====")

	rand.Seed(time.Now().UnixNano())

	// Compute all priors.
	rows := db.NewStatement(`
		SELECT id,firstName,lastName,email
		FROM users
		WHERE NOT resetEmailSent AND firstName!=""`).Query()
	err = rows.Process(resetPasswordsProcessPage)
	if err != nil {
		c.Errorf("ERROR: %v", err)
		// Error or not, we don't want to rerun this.
	}
	return 0, err
}

func resetPasswordsProcessPage(db *database.DB, rows *database.Rows) error {
	/*var id, firstName, lastName, email string
	if err := rows.Scan(&id, &firstName, &lastName, &email); err != nil {
		return fmt.Errorf("Failed to scan for page: %v", err)
	}

	hashmap := make(database.InsertMap)
	hashmap["id"] = id
	hashmap["resetEmailSent"] = true
	statement := db.NewInsertStatement("users", hashmap, "resetEmailSent")
	if _, err := statement.Exec(); err != nil {
		return fmt.Errorf("Couldn't update user's record: %v", err)
	}

	err := stormpath.CreateNewUser(db.C, firstName, lastName, email, fmt.Sprintf("Pwd%d", rand.Int63()))
	if err != nil {
		return fmt.Errorf("Couldn't create a new user: %v", err)
	}

	err = stormpath.ForgotPassword(db.C, email)
	if err != nil {
		return fmt.Errorf("Couldn't forget password", err)
	}*/

	return nil
}
