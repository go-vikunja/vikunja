package models

import (
	"code.vikunja.io/api/pkg/log"
)

// CanCreate checks if the user can create a new team
func (t *Team) CanCreate(u *User) bool {
	// This is currently a dummy function, later on we could imagine global limits etc.
	return true
}

// CanUpdate checks if the user can update a team
func (t *Team) CanUpdate(u *User) bool {

	// Check if the current user is in the team and has admin rights in it
	exists, err := x.Where("team_id = ?", t.ID).
		And("user_id = ?", u.ID).
		And("admin = ?", true).
		Get(&TeamMember{})
	if err != nil {
		log.Log.Error("Error occurred during CanUpdate for Team: %s", err)
		return false
	}

	return exists
}

// CanDelete checks if a user can delete a team
func (t *Team) CanDelete(u *User) bool {
	return t.IsAdmin(u)
}

// IsAdmin returns true when the user is admin of a team
func (t *Team) IsAdmin(u *User) bool {
	exists, err := x.Where("team_id = ?", t.ID).
		And("user_id = ?", u.ID).
		And("admin = ?", true).
		Get(&TeamMember{})
	if err != nil {
		log.Log.Error("Error occurred during CanUpdate for Team: %s", err)
		return false
	}
	return exists
}

// CanRead returns true if the user has read access to the team
func (t *Team) CanRead(user *User) bool {
	// Check if the user is in the team
	exists, err := x.Where("team_id = ?", t.ID).
		And("user_id = ?", user.ID).
		Get(&TeamMember{})
	if err != nil {
		log.Log.Error("Error occurred during CanUpdate for Team: %s", err)
		return false
	}
	return exists
}
