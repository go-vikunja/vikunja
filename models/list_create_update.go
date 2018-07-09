package models

// CreateOrUpdateList updates a list or creates it if it doesn't exist
func CreateOrUpdateList(list *List) (err error) {

	if list.ID == 0 {
		_, err = x.Insert(list)
	} else {
		_, err = x.ID(list.ID).Update(list)
	}

	if err != nil {
		return
	}

	*list, err = GetListByID(list.ID)

	return

}

func (l *List) Update(id int64, doer *User) (err error)  {
	l.ID = id

	// Check if it exists
	oldList, err := GetListByID(l.ID)
	if err != nil {
		return
	}

	// Check rights
	user, _, err := GetUserByID(doer.ID)
	if err != nil {
		return
	}

	if !oldList.IsAdmin(&user) {
		return ErrNeedToBeListAdmin{ListID:id, UserID:user.ID}
	}

	return CreateOrUpdateList(l)
}

func (l *List) Create(doer *User) (err error) {
	// Check rights
	user, _, err := GetUserByID(doer.ID)
	if err != nil {
		return
	}

	// Get the namespace of the list to check if the user can write to it
	namespace, err := GetNamespaceByID(l.NamespaceID)
	if err != nil {
		return
	}
	if !namespace.CanWrite(doer) {
		return ErrUserDoesNotHaveWriteAccessToNamespace{UserID:user.ID, NamespaceID:namespace.ID}
	}

	l.Owner.ID = user.ID

	return CreateOrUpdateList(l)
}