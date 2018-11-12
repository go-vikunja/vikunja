package models

// Create implements the creation method via the interface
// @Summary Creates a new namespace
// @Description Creates a new namespace.
// @tags namespace
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param namespace body models.Namespace true "The namespace you want to create."
// @Success 200 {object} models.Namespace "The created namespace."
// @Failure 400 {object} models.HTTPError "Invalid namespace object provided."
// @Failure 403 {object} models.HTTPError "The user does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces [put]
func (n *Namespace) Create(doer *User) (err error) {
	// Check if we have at least a name
	if n.Name == "" {
		return ErrNamespaceNameCannotBeEmpty{NamespaceID: 0, UserID: doer.ID}
	}
	n.ID = 0 // This would otherwise prevent the creation of new lists after one was created

	// Check if the User exists
	n.Owner, err = GetUserByID(doer.ID)
	if err != nil {
		return
	}
	n.OwnerID = n.Owner.ID

	// Insert
	_, err = x.Insert(n)
	return
}
