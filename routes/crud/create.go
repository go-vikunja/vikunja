package crud

import (
	"github.com/labstack/echo"
	"net/http"
	"fmt"
	"git.kolaente.de/konrad/list/models"
)

// CreateWeb is the handler to create an object
func (c *WebHandler) CreateWeb(ctx echo.Context) error {
	// Get the object
	if err := ctx.Bind(&c.CObject); err != nil {
		return ctx.JSON(http.StatusBadRequest, models.Message{"No model provided."})
	}

	// Get the user to pass for later checks
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}

	// Get an ID if we have one
	var id int64 = 0
	if ctx.Param("id") != "" {
		id, err := models.GetIntURLParam("id", ctx)
		if err != nil {

		}
	}

	// Create
	err = c.CObject.Create(&currentUser)
	if err != nil {
		fmt.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, c.CObject)
}