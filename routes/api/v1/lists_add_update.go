package v1

import (
	"net/http"
	"github.com/labstack/echo"
	"git.kolaente.de/konrad/list/models"
	"strconv"
	"fmt"
)

func AddOrUpdateList(c echo.Context) error {

	// Get the list
	var list *models.List

	if err := c.Bind(&list); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"No list model provided."})
	}

	// Check if we have an ID other than the one in the struct
	id := c.Param("id")
	if id != "" {
		// Make int
		listID, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
		}
		list.ID = listID
	}

	// Check if the list exists
	// ID = 0 means new list, no error
	if list.ID != 0 {
		_, err := models.GetListByID(list.ID)
		if err != nil {
			if models.IsErrListDoesNotExist(err) {
				return c.JSON(http.StatusBadRequest, models.Message{"The list does not exist."})
			} else {
				return c.JSON(http.StatusInternalServerError, models.Message{"Could not check if the list exists."})
			}
		}
	}

	// Get the current user for later checks
	user, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}
	list.Owner = user

	// update or create...
	if list.ID == 0 {
		err = models.CreateOrUpdateList(list)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
		}
	} else {
		// Check if the user owns the list
		oldList, err := models.GetListByID(list.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
		}
		if user.ID != oldList.Owner.ID {
			return c.JSON(http.StatusForbidden, models.Message{"You cannot edit a list you don't own."})
		}

		err = models.CreateOrUpdateList(list)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
		}
	}

	return c.JSON(http.StatusOK, list)
}