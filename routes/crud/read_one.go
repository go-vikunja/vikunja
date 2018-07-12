package crud

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
)

// ReadOneWeb is the webhandler to get one object
func (c *WebHandler) ReadOneWeb(ctx echo.Context) error {

	// Get the ID
	id, err := models.GetIntURLParam("id", ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID.")
	}

	// Get our object
	err = c.CObject.ReadOne(id)
	if err != nil {
		if models.IsErrListDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		if models.IsErrNamespaceDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "An error occured.")
	}

	// Check rights
	// We can only check the rights on a full object, which is why we need to check it afterwards
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}
	if !c.CObject.CanRead(&currentUser) {
		return echo.NewHTTPError(http.StatusForbidden, "You don't have the right to see this")
	}

	return ctx.JSON(http.StatusOK, c.CObject)
}
