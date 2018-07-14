package models

func (t *Team) Create(doer *User, _ int64) (err error) {
	// Check if we have a name
	if t.Name == "" {
		return ErrTeamNameCannotBeEmpty{}
	}

	// Set the id to 0, otherwise the creation fails because of double keys
	t.CreatedByID = doer.ID
	t.CreatedBy = doer

	_, err = x.Insert(t)
	if err != nil {
		return
	}

	return
}
