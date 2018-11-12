package models

// ReadAll gets all users who have access to a list
// @Summary Get users on a list
// @Description Returns a list with all users which have access on a given list.
// @tags sharing
// @Accept json
// @Produce json
// @Param id path int true "List ID"
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search users by its name."
// @Security ApiKeyAuth
// @Success 200 {array} models.UserWithRight "The users with the right they have."
// @Failure 403 {object} models.HTTPError "No right to see the list."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id}/users [get]
func (ul *ListUser) ReadAll(search string, u *User, page int) (interface{}, error) {
	// Check if the user has access to the list
	l := &List{ID: ul.ListID}
	if err := l.GetSimpleByID(); err != nil {
		return nil, err
	}
	if !l.CanRead(u) {
		return nil, ErrNeedToHaveListReadAccess{}
	}

	// Get all users
	all := []*UserWithRight{}
	err := x.
		Join("INNER", "users_list", "user_id = users.id").
		Where("users_list.list_id = ?", ul.ListID).
		Limit(getLimitFromPageIndex(page)).
		Where("users.username LIKE ?", "%"+search+"%").
		Find(&all)

	return all, err
}
