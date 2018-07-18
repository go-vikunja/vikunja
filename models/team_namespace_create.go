package models

import "fmt"

// Create creates a new team <-> namespace relation
func (tn *TeamNamespace) Create(doer *User) (err error) {

	// Check if the rights are valid
	if tn.Right != NamespaceRightAdmin && tn.Right != NamespaceRightRead && tn.Right != NamespaceRightWrite {
		return ErrInvalidTeamRight{tn.Right}
	}

	fmt.Println(tn.NamespaceID)

	// Check if the team exists
	_, err = GetTeamByID(tn.TeamID)
	if err != nil {
		return
	}

	// Check if the namespace exists
	_, err = GetNamespaceByID(tn.NamespaceID)
	if err != nil {
		return
	}

	// Insert the new team
	_, err = x.Insert(tn)
	return
}
