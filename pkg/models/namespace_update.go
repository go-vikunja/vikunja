package models

// Update implements the update method via the interface
// @Summary Updates a namespace
// @Description Updates a namespace.
// @tags namespace
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Namespace ID"
// @Param namespace body models.Namespace true "The namespace with updated values you want to update."
// @Success 200 {object} models.Namespace "The updated namespace."
// @Failure 400 {object} models.HTTPError "Invalid namespace object provided."
// @Failure 403 {object} models.HTTPError "The user does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespace/{id} [post]
func (n *Namespace) Update() (err error) {
	// Check if we have at least a name
	if n.Name == "" {
		return ErrNamespaceNameCannotBeEmpty{NamespaceID: n.ID}
	}

	// Check if the namespace exists
	currentNamespace, err := GetNamespaceByID(n.ID)
	if err != nil {
		return
	}

	// Check if the (new) owner exists
	n.OwnerID = n.Owner.ID
	if currentNamespace.OwnerID != n.OwnerID {
		n.Owner, err = GetUserByID(n.OwnerID)
		if err != nil {
			return
		}
	}

	// Do the actual update
	_, err = x.ID(currentNamespace.ID).Update(n)
	return
}
