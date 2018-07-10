package crud

import (
	"github.com/labstack/echo"
	"net/http"
	"git.kolaente.de/konrad/list/models"
)

// UpdateWeb is the webhandler to update an object
func (c *WebHandler) UpdateWeb(ctx echo.Context) error {
	// Get the object
	if err := ctx.Bind(&c.CObject); err != nil {
		return ctx.JSON(http.StatusBadRequest, models.Message{"No model provided."})
	}

	// Get the ID
	id, err := models.GetIntURLParam("id", ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	// Check if the user has the right to do that
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.Message{"Could not determine the current user."})
	}

	// Do the update
	err = c.CObject.Update(id, &currentUser)
	if err != nil {
		if models.IsErrNeedToBeListAdmin(err) {
			return echo.NewHTTPError(http.StatusForbidden, "You need to be list admin to do that.")
		}

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, c.CObject)
}
