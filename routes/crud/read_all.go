package crud

import (
	"fmt"
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
)

// ReadAllWeb is the webhandler to get all objects of a type
func (c *WebHandler) ReadAllWeb(ctx echo.Context) error {
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}

	// Get the object & bind params to struct
	if err := ParamBinder(c.CObject, ctx); err != nil {
		fmt.Println(err)
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided.")
	}

	lists, err := c.CObject.ReadAll(&currentUser)
	if err != nil {
		fmt.Println(err)

		return echo.NewHTTPError(http.StatusInternalServerError, "An error occured.")
	}

	return ctx.JSON(http.StatusOK, lists)
}
