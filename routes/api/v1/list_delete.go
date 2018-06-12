package v1

import (
	"github.com/labstack/echo"
	"strconv"
	"net/http"
	"git.kolaente.de/konrad/list/models"
)

func DeleteListByID(c echo.Context) error {
	// Check if we have our ID
	id := c.Param("id")
	// Make int
	itemID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	// Check if the user has the right to delete that list
	user, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	err = models.DeleteListByID(itemID, &user)
	if err != nil {
		if models.IsErrNeedToBeListOwner(err) {
			return c.JSON(http.StatusForbidden, models.Message{"You need to be the list owner to delete a list."})
		}

		if models.IsErrListDoesNotExist(err) {
			return c.JSON(http.StatusNotFound, models.Message{"This list does not exist."})
		}

		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	return c.JSON(http.StatusOK, models.Message{"The list was deleted with success."})
}