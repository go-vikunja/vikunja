package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
)

// GetListsByUser gets all lists a user owns
func GetListsByUser(c echo.Context) error {

	currentUser, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"Could not determine the current user."})
	}

	allLists, err := models.GetListsByUser(&currentUser)
	if err != nil {

		if models.IsErrListDoesNotExist(err) {

		}

		return c.JSON(http.StatusInternalServerError, models.Message{"Could not get lists."})
	}

	return c.JSON(http.StatusOK, allLists)
}
