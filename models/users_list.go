package models

// ListUsers returns a list with all users, filtered by an optional searchstring
func ListUsers(searchterm string) (users []User, err error) {

	if searchterm == "" {
		err = x.Find(&users)
	} else {
		err = x.
			Where("username LIKE ?", "%"+searchterm+"%").
			Find(&users)
	}

	if err != nil {
		return []User{}, err
	}

	return users, nil
}
