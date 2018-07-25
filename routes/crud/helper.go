package crud

import (
	"code.vikunja.io/api/models"
)

// WebHandler defines the webhandler object
// This does web stuff, aka returns json etc. Uses CRUDable Methods to get the data
type WebHandler struct {
	CObject interface {
		models.CRUDable
		models.Rights
	}
}
