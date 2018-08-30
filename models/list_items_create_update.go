package models

// Create is the implementation to create a list task
func (i *ListTask) Create(doer *User) (err error) {
	//i.ListID = lID
	i.ID = 0

	return createOrUpdateListTask(i, doer)
}

// Update updates a list task
func (i *ListTask) Update() (err error) {
	// Check if the task exists
	_, err = GetListTaskByID(i.ID)
	if err != nil {
		return
	}

	return createOrUpdateListTask(i, &User{})
}

// Helper function for creation or updating of new lists as both methods share most of their logic
func createOrUpdateListTask(i *ListTask, doer *User) (err error) {

	// Check if we have at least a text
	if i.Text == "" {
		return ErrListTaskCannotBeEmpty{}
	}

	// Check if the list exists
	_, err = GetListByID(i.ListID)
	if err != nil {
		return
	}

	// Do the update
	if i.ID != 0 {
		_, err = x.ID(i.ID).Update(i)
	} else {
		user, _, err := GetUserByID(doer.ID)
		if err != nil {
			return err
		}

		i.CreatedByID = user.ID
		i.CreatedBy = user
		_, err = x.Insert(i)
	}

	return
}
