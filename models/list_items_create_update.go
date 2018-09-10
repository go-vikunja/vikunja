package models

// Create is the implementation to create a list task
func (i *ListTask) Create(doer *User) (err error) {
	i.ID = 0

	// Check if we have at least a text
	if i.Text == "" {
		return ErrListTaskCannotBeEmpty{}
	}

	// Check if the list exists
	_, err = GetListByID(i.ListID)
	if err != nil {
		return
	}

	user, err := GetUserByID(doer.ID)
	if err != nil {
		return err
	}

	i.CreatedByID = user.ID
	i.CreatedBy = user
	_, err = x.Insert(i)
	return err
}

// Update updates a list task
func (i *ListTask) Update() (err error) {
	// Check if the task exists
	_, err = GetListTaskByID(i.ID)
	if err != nil {
		return
	}

	// Do the update
	_, err = x.ID(i.ID).Update(i)
	return err
}
