package crud

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"reflect"
)

// UpdateWeb is the webhandler to update an object
func (c *WebHandler) UpdateWeb(ctx echo.Context) error {
	// Re-initialize our model
	p := reflect.ValueOf(c.CObject).Elem()
	p.Set(reflect.Zero(p.Type()))

	// Get the object
	if err := ctx.Bind(&c.CObject); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided.")
	}

	// Get the ID
	id, err := models.GetIntURLParam("id", ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID.")
	}

	// Check if the user has the right to do that
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}
	if !c.CObject.CanUpdate(&currentUser, id) {
		return echo.NewHTTPError(http.StatusForbidden)
	}

	// Do the update
	err = c.CObject.Update(id)
	if err != nil {
		if models.IsErrNeedToBeListAdmin(err) {
			return echo.NewHTTPError(http.StatusForbidden, "You need to be list admin to do that.")
		}

		if models.IsErrNamespaceDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound, "The namespace does not exist.")
		}
		if models.IsErrNamespaceNameCannotBeEmpty(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "The namespace name cannot be empty.")
		}
		if models.IsErrNamespaceOwnerCannotBeEmpty(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "The namespace owner cannot be empty.")
		}

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, c.CObject)
}
