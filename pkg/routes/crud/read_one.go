package crud

import (
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"github.com/labstack/echo"
	"net/http"
)

// ReadOneWeb is the webhandler to get one object
func (c *WebHandler) ReadOneWeb(ctx echo.Context) error {
	// Get our model
	currentStruct := c.EmptyStruct()

	// Get the object & bind params to struct
	if err := ParamBinder(currentStruct, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided.")
	}

	// Get our object
	err := currentStruct.ReadOne()
	if err != nil {
		return HandleHTTPError(err)
	}

	// Check rights
	// We can only check the rights on a full object, which is why we need to check it afterwards
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}
	if !currentStruct.CanRead(&currentUser) {
		log.Log.Noticef("%s [ID: %d] tried to read while not having the rights for it", currentUser.Username, currentUser.ID)
		return echo.NewHTTPError(http.StatusForbidden, "You don't have the right to see this")
	}

	return ctx.JSON(http.StatusOK, currentStruct)
}
