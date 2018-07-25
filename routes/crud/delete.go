package crud

import (
	"code.vikunja.io/api/models"
	"github.com/labstack/echo"
	"net/http"
)

// DeleteWeb is the web handler to delete something
func (c *WebHandler) DeleteWeb(ctx echo.Context) error {
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
