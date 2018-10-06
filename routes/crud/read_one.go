package crud

import (
	"code.vikunja.io/api/models"
	"github.com/labstack/echo"
	"net/http"
)

// ReadOneWeb is the webhandler to get one object
func (c *WebHandler) ReadOneWeb(ctx echo.Context) error {

	// Get the object & bind params to struct
	if err := ParamBinder(c.CObject, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided.")
	}

	// Get our object
	err := c.CObject.ReadOne()
	if err != nil {
		return HandleHTTPError(err)
	}

	// Check rights
	// We can only check the rights on a full object, which is why we need to check it afterwards
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}
	if !c.CObject.CanRead(&currentUser) {
		models.Log.Noticef("%s [ID: %d] tried to read while not having the rights for it", currentUser.Username, currentUser.ID)
		return echo.NewHTTPError(http.StatusForbidden, "You don't have the right to see this")
	}

	return ctx.JSON(http.StatusOK, c.CObject)
}
