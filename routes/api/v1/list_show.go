package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

// AddOrUpdateList Adds or updates a new list
func GetListByID(c echo.Context) error {
	// Check if we have our ID
	id := c.Param("id")
	// Make int
	listID, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	// Get the list
	list, err := models.GetListByID(listID)
	if err != nil {
		if models.IsErrListDoesNotExist(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"The list does not exist."})
		}

		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	return c.JSON(http.StatusOK, list)
}
