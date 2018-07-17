package models

// Create creates a new team <-> namespace relation
func (tn *TeamNamespace) Create(doer *User, nID int64) (err error) {

	// Check if the rights are valid
	if tn.Right != NamespaceRightAdmin && tn.Right != NamespaceRightRead && tn.Right != NamespaceRightWrite {
		return ErrInvalidTeamRight{tn.Right}
	}

	// Check if the team exists
	_, err = GetTeamByID(tn.TeamID)
	if err != nil {
		return
	}

	// Check if the namespace exists
	_, err = GetNamespaceByID(nID)
	if err != nil {
		return
	}
	tn.NamespaceID = nID

	// Insert the new team
	_, err = x.Insert(tn)
	return
}
