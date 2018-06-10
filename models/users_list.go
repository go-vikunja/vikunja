package models

// ListUsers returns a list with all users, filtered by an optional searchstring
func ListUsers(searchterm string) (users []User, err error) {

	if searchterm == "" {
		err = x.Find(&users)
	} else {
		err = x.
			Where("username LIKE ?", "%"+searchterm+"%").
			Or("name LIKE ?", "%"+searchterm+"%").
			Find(&users)
	}

	// Obfuscate the password. Selecting everything except the password didn't work.
	for i := range users {
		users[i].Password = ""
	}

	if err != nil {
		return []User{}, err
	}

	return users, nil
}
