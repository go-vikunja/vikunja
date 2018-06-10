package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

func AddOrUpdateListItem(c echo.Context) error {
	// Get the list item
	var listItem *models.ListItem

	if err := c.Bind(&listItem); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"No list model provided."})
	}

	// Get the list ID
	id := c.Param("id")
	// Make int
	listID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}
	listItem.ListID = listID

	// Set the user
	user, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}
	listItem.CreatedBy = user

	err = models.CreateOrUpdateListItem(listItem)
	if err != nil {
		if models.IsErrListDoesNotExist(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"The list does not exist."})
		}
		if models.IsErrListItemCannotBeEmpty(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"You must provide at least a list item text."})
		}
		if models.IsErrUserDoesNotExist(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"The user does not exist."})
		}

		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	return c.JSON(http.StatusOK, listItem)
}
