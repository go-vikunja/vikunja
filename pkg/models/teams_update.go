package models

// Update is the handler to create a team
func (t *Team) Update() (err error) {
	// Check if we have a name
	if t.Name == "" {
		return ErrTeamNameCannotBeEmpty{}
	}

	// Check if the team exists
	_, err = GetTeamByID(t.ID)
	if err != nil {
		return
	}

	_, err = x.ID(t.ID).Update(t)
	if err != nil {
		return
	}

	// Get the newly updated team
	*t, err = GetTeamByID(t.ID)

	return
}
