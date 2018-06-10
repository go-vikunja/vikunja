package v1

import (
	"github.com/labstack/echo"
	"git.kolaente.de/konrad/list/models"
	"net/http"
)

func GetListsByUser(c echo.Context) error {

	currentUser, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"Could not determine the current user."})
	}

	allLists, err := models.GetListsByUser(&currentUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"Could not get lists."})
	}

	return c.JSON(http.StatusOK, allLists)
}