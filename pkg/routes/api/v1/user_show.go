package v1

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes/crud"
	"github.com/labstack/echo"
	"net/http"
)

// UserShow gets all informations about the current user
// @Summary Get user information
// @Description Returns the current user object.
// @tags user
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.User
// @Failure 404 {object} models.HTTPError "User does not exist."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user [get]
func UserShow(c echo.Context) error {
	userInfos, err := models.GetCurrentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error getting current user.")
	}

	user, err := models.GetUserByID(userInfos.ID)
	if err != nil {
		return crud.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, user)
}
