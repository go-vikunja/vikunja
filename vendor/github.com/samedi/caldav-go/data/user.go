package data

// CalUser represents the calendar user. It is used, for example, to
// keep track globally what is the current user interacting with the calendar.
// This user data can be used in various places, including in some of the CALDAV responses.
type CalUser struct {
	Name string
}
