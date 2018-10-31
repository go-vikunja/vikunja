package models

import (
	"code.vikunja.io/api/pkg/log"
)

// CanCreate checks if the user can add a new tem member
func (tm *TeamMember) CanCreate(u *User) bool {
	return tm.IsAdmin(u)
}

// CanDelete checks if the user can delete a new team member
func (tm *TeamMember) CanDelete(u *User) bool {
	return tm.IsAdmin(u)
}

// IsAdmin checks if the user is team admin
func (tm *TeamMember) IsAdmin(u *User) bool {
	// A user can add a member to a team if he is admin of that team
	exists, err := x.Where("user_id = ? AND team_id = ? AND admin = ?", u.ID, tm.TeamID, true).
		Get(&TeamMember{})
	if err != nil {
		log.Log.Error("Error occurred during IsAdmin for TeamMember: %s", err)
		return false
	}
	return exists
}
