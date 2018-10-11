package crud

import (
	"code.vikunja.io/api/models"
	"github.com/labstack/echo"
	"net/http"
)

// WebHandler defines the webhandler object
// This does web stuff, aka returns json etc. Uses CRUDable Methods to get the data
type WebHandler struct {
	EmptyStruct func() CObject
}

// CObject is the definition of our object, holds the structs
type CObject interface {
	models.CRUDable
	models.Rights
}

// HandleHTTPError does what it says
func HandleHTTPError(err error) *echo.HTTPError {
	if a, has := err.(models.HTTPErrorProcessor); has {
		errDetails := a.HTTPError()
		return echo.NewHTTPError(errDetails.HTTPCode, errDetails)
	}
	models.Log.Error(err.Error())
	return echo.NewHTTPError(http.StatusInternalServerError)
}
