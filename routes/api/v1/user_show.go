package v1

import (
	"code.vikunja.io/api/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

// UserShow gets all informations about a user
func UserShow(c echo.Context) error {

	// TODO: only allow users to show itself/with privacy options

	user := c.Param("id")

	if user == "" {
		return c.JSON(http.StatusBadRequest, models.Message{"User ID cannot be empty."})
	}

	// Make int
	userID, err := strconv.ParseInt(user, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"User ID is invalid."})
	}

	// Get User Infos
	userInfos, err := models.GetUserByID(userID)

	if err != nil {
		if models.IsErrUserDoesNotExist(err) {
			return c.JSON(http.StatusNotFound, models.Message{"The user does not exist."})
		}
		return c.JSON(http.StatusInternalServerError, models.Message{"Error getting user infos."})
	}

	// Obfucate his password
	userInfos.Password = ""

	return c.JSON(http.StatusOK, userInfos)
}
