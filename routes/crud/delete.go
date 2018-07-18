package crud

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"fmt"
)

// DeleteWeb is the web handler to delete something
func (c *WebHandler) DeleteWeb(ctx echo.Context) error {
	// Get the ID
	/*id, err := models.GetIntURLParam("id", ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID.")
	}*/
	// Bind params to struct
	if err := ParamBinder(c.CObject, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid URL param.")
	}

	// Check if the user has the right to delete
	user, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if !c.CObject.CanDelete(&user) {
		return echo.NewHTTPError(http.StatusForbidden)
	}

	err = c.CObject.Delete()
	if err != nil {

		fmt.Println(err)

		if models.IsErrNeedToBeListAdmin(err) {
			return echo.NewHTTPError(http.StatusForbidden, "You need to be the list admin to delete a list.")
		}

		if models.IsErrListDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound, "This list does not exist.")
		}

		if models.IsErrTeamDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound, "This team does not exist.")
		}

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, models.Message{"Successfully deleted."})
}
