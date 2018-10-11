package crud

import (
	"code.vikunja.io/api/models"
	"github.com/labstack/echo"
	"net/http"
)

// ReadAllWeb is the webhandler to get all objects of a type
func (c *WebHandler) ReadAllWeb(ctx echo.Context) error {
	// Get our model
	currentStruct := c.EmptyStruct()

	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}

	// Get the object & bind params to struct
	if err := ParamBinder(currentStruct, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided.")
	}

	lists, err := currentStruct.ReadAll(&currentUser)
	if err != nil {
		return HandleHTTPError(err)
	}

	return ctx.JSON(http.StatusOK, lists)
}
