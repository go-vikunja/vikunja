package models

// ReadAll implements the method to read all teams of a namespace
// @Summary Get teams on a namespace
// @Description Returns a namespace with all teams which have access on a given namespace.
// @tags sharing
// @Accept json
// @Produce json
// @Param id path int true "Namespace ID"
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search teams by its name."
// @Security ApiKeyAuth
// @Success 200 {array} models.TeamWithRight "The teams with the right they have."
// @Failure 403 {object} models.HTTPError "No right to see the namespace."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id}/teams [get]
func (tn *TeamNamespace) ReadAll(search string, user *User, page int) (interface{}, error) {
	// Check if the user can read the namespace
	n, err := GetNamespaceByID(tn.NamespaceID)
	if err != nil {
		return nil, err
	}
	if !n.CanRead(user) {
		return nil, ErrNeedToHaveNamespaceReadAccess{NamespaceID: tn.NamespaceID, UserID: user.ID}
	}

	// Get the teams
	all := []*TeamWithRight{}

	err = x.Table("teams").
		Join("INNER", "team_namespaces", "team_id = teams.id").
		Where("team_namespaces.namespace_id = ?", tn.NamespaceID).
		Limit(getLimitFromPageIndex(page)).
		Where("teams.name LIKE ?", "%"+search+"%").
		Find(&all)

	return all, err
}
