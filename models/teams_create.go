package models

// Create is the handler to create a team
func (t *Team) Create(doer *User) (err error) {
	// Check if we have a name
	if t.Name == "" {
		return ErrTeamNameCannotBeEmpty{}
	}

	t.CreatedByID = doer.ID
	t.CreatedBy = *doer

	_, err = x.Insert(t)
	if err != nil {
		return
	}

	// Insert the current user as member and admin
	tm := TeamMember{TeamID: t.ID, UserID: doer.ID, IsAdmin: true}
	err = tm.Create(doer)
	return
}
