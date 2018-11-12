package models

// CreateOrUpdateList updates a list or creates it if it doesn't exist
func CreateOrUpdateList(list *List) (err error) {

	// Check we have at least a title
	if list.Title == "" {
		return ErrListTitleCannotBeEmpty{}
	}

	// Check if the namespace exists
	if list.NamespaceID != 0 {
		_, err = GetNamespaceByID(list.NamespaceID)
		if err != nil {
			return err
		}
	}

	if list.ID == 0 {
		_, err = x.Insert(list)
	} else {
		_, err = x.ID(list.ID).Update(list)
	}

	if err != nil {
		return
	}

	err = list.ReadOne()
	return

}

// Update implements the update method of CRUDable
// @Summary Updates a list
// @Description Updates a list. This does not include adding a task (see below).
// @tags list
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "List ID"
// @Param list body models.List true "The list with updated values you want to update."
// @Success 200 {object} models.List "The updated list."
// @Failure 400 {object} models.HTTPError "Invalid list object provided."
// @Failure 403 {object} models.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id} [post]
func (l *List) Update() (err error) {
	// Check if it exists
	if err = l.GetSimpleByID(); err != nil {
		return
	}

	return CreateOrUpdateList(l)
}

// Create implements the create method of CRUDable
// @Summary Creates a new list
// @Description Creates a new list in a given namespace. The user needs write-access to the namespace.
// @tags list
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param namespaceID path int true "Namespace ID"
// @Param list body models.List true "The list you want to create."
// @Success 200 {object} models.List "The created list."
// @Failure 400 {object} models.HTTPError "Invalid list object provided."
// @Failure 403 {object} models.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/lists [put]
func (l *List) Create(doer *User) (err error) {
	// Check rights
	u, err := GetUserByID(doer.ID)
	if err != nil {
		return
	}

	l.OwnerID = u.ID
	l.Owner.ID = u.ID
	l.ID = 0 // Otherwise only the first time a new list would be created

	return CreateOrUpdateList(l)
}
