package v1

import (
	"git.kolaente.de/konrad/list/models"
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
	userInfos, exists, err := models.GetUserByID(userID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"Error getting user infos."})
	}

	// Check if it exists
	if !exists {
		return c.JSON(http.StatusNotFound, models.Message{"User not found."})
	}

	// Obfucate his password
	userInfos.Password = ""

	return c.JSON(http.StatusOK, userInfos)
}
