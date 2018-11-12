package models

// ReadAll gets all users who have access to a namespace
// @Summary Get users on a namespace
// @Description Returns a namespace with all users which have access on a given namespace.
// @tags sharing
// @Accept json
// @Produce json
// @Param id path int true "Namespace ID"
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search users by its name."
// @Security ApiKeyAuth
// @Success 200 {array} models.UserWithRight "The users with the right they have."
// @Failure 403 {object} models.HTTPError "No right to see the namespace."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id}/users [get]
func (un *NamespaceUser) ReadAll(search string, u *User, page int) (interface{}, error) {
	// Check if the user has access to the namespace
	l, err := GetNamespaceByID(un.NamespaceID)
	if err != nil {
		return nil, err
	}
	if !l.CanRead(u) {
		return nil, ErrNeedToHaveNamespaceReadAccess{}
	}

	// Get all users
	all := []*UserWithRight{}
	err = x.
		Join("INNER", "users_namespace", "user_id = users.id").
		Where("users_namespace.namespace_id = ?", un.NamespaceID).
		Limit(getLimitFromPageIndex(page)).
		Where("users.username LIKE ?", "%"+search+"%").
		Find(&all)

	return all, err
}
