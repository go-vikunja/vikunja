package crud

import (
	"code.vikunja.io/api/models"
	"fmt"
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
		fmt.Println(err)
		if models.IsErrNeedToBeListAdmin(err) {
			return echo.NewHTTPError(http.StatusForbidden, "You need to be the list admin to delete a list.")
		}

		if models.IsErrListDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound, "This list does not exist.")
		}
		if models.IsErrTeamDoesNotHaveAccessToList(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "This team does not have access to the list.")
		}

		if models.IsErrTeamDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound, "This team does not exist.")
		}

		if models.IsErrCannotDeleteLastTeamMember(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "You cannot delete the last member of a team.")
		}

		if models.IsErrUserDoesNotHaveAccessToList(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "This user does not have access to the list.")
		}

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, models.Message{"Successfully deleted."})
}
