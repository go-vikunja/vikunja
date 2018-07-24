package models

// Create implements the create method to assign a user to a team
func (tm *TeamMember) Create(doer *User) (err error) {
	// TODO: Check if it exists etc
	_, err = x.Insert(tm)
	return
}
