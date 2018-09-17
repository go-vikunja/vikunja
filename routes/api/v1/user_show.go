package v1

import (
	"code.vikunja.io/api/models"
	"github.com/labstack/echo"
	"net/http"
)

// UserShow gets all information about a user
func UserShow(c echo.Context) error {
	userInfos, err := models.GetCurrentUser(c)
	if err != nil {
		if models.IsErrUserDoesNotExist(err) {
			return c.JSON(http.StatusNotFound, models.Message{"The user does not exist."})
		}
		return c.JSON(http.StatusInternalServerError, models.Message{"Error getting user infos."})
	}

	user, err := models.GetUserByID(userInfos.ID)
	if err != nil {
		if models.IsErrUserDoesNotExist(err) {
			return c.JSON(http.StatusNotFound, models.Message{"The user does not exist."})
		}
		return c.JSON(http.StatusInternalServerError, models.Message{"Error getting user infos."})
	}
	// Obfuscate his password
	user.Password = ""

	return c.JSON(http.StatusOK, user)
}
