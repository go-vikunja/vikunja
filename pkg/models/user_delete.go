package models

// DeleteUserByID deletes a user by its ID
func DeleteUserByID(id int64, doer *User) error {
	// Check if the id is 0
	if id == 0 {
		return ErrIDCannotBeZero{}
	}

	// Delete the user
	_, err := x.Id(id).Delete(&User{})

	if err != nil {
		return err
	}

	return err
}
