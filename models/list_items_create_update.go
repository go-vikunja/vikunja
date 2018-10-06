package models

import (
	"github.com/imdario/mergo"
)

// Create is the implementation to create a list task
func (i *ListTask) Create(doer *User) (err error) {
	i.ID = 0

	// Check if we have at least a text
	if i.Text == "" {
		return ErrListTaskCannotBeEmpty{}
	}

	// Check if the list exists
	l := &List{ID: i.ListID}
	if err = l.GetSimpleByID(); err != nil {
		return
	}

	user, err := GetUserByID(doer.ID)
	if err != nil {
		return err
	}

	i.CreatedByID = user.ID
	i.CreatedBy = user
	_, err = x.Cols("text", "description", "done", "due_date_unix", "reminder_unix", "created_by_id", "list_id", "created", "updated").Insert(i)
	return err
}

// Update updates a list task
func (i *ListTask) Update() (err error) {
	// Check if the task exists
	ot, err := GetListTaskByID(i.ID)
	if err != nil {
		return
	}

	// For whatever reason, xorm dont detect if done is updated, so we need to update this every time by hand
	// Which is why we merge the actual task struct with the one we got from the user.
	// The user struct ovverrides values in the actual one.
	if err := mergo.Merge(&ot, i, mergo.WithOverride); err != nil {
		return err
	}

	// And because a false is considered to be a null value, we need to explicitly check that case here.
	if i.Done == false {
		ot.Done = false
	}

	_, err = x.ID(i.ID).Cols("text", "description", "done", "due_date_unix", "reminder_unix").Update(ot)
	*i = ot
	return
}
