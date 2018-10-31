package models

// TeamRight defines the rights teams can have for lists/namespaces
type TeamRight int

// define unknown team right
const (
	TeamRightUnknown = -1
)

// Enumerate all the team rights
const (
	// Can read lists in a Team
	TeamRightRead TeamRight = iota
	// Can write tasks in a Team like lists and todo tasks. Cannot create new lists.
	TeamRightWrite
	// Can manage a list/namespace, can do everything
	TeamRightAdmin
)

func (r TeamRight) isValid() error {
	if r != TeamRightAdmin && r != TeamRightRead && r != TeamRightWrite {
		return ErrInvalidTeamRight{r}
	}

	return nil
}
