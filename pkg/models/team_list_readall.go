package models

// ReadAll implements the method to read all teams of a list
// @Summary Get teams on a list
// @Description Returns a list with all teams which have access on a given list.
// @tags sharing
// @Accept json
// @Produce json
// @Param id path int true "List ID"
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search teams by its name."
// @Security ApiKeyAuth
// @Success 200 {array} models.TeamWithRight "The teams with their right."
// @Failure 403 {object} models.HTTPError "No right to see the list."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id}/teams [get]
func (tl *TeamList) ReadAll(search string, u *User, page int) (interface{}, error) {
	// Check if the user can read the namespace
	l := &List{ID: tl.ListID}
	if err := l.GetSimpleByID(); err != nil {
		return nil, err
	}
	if !l.CanRead(u) {
		return nil, ErrNeedToHaveListReadAccess{ListID: tl.ListID, UserID: u.ID}
	}

	// Get the teams
	all := []*TeamWithRight{}
	err := x.
		Table("teams").
		Join("INNER", "team_list", "team_id = teams.id").
		Where("team_list.list_id = ?", tl.ListID).
		Limit(getLimitFromPageIndex(page)).
		Where("teams.name LIKE ?", "%"+search+"%").
		Find(&all)

	return all, err
}
