package models

// CanCreate checks if the user can create a new team
func (n *Team) CanCreate(user *User, id int64) bool {
	// This is currently a dummy function, later on we could imagine global limits etc.
	return true
}