package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

// UserDelete is the handler to delete a user
func UserDelete(c echo.Context) error {

	// TODO: only allow users to allow itself

	id := c.Param("id")

	// Make int
	userID, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"User ID is invalid."})
	}

	// Check if the user exists
	_, exists, err := models.GetUserByID(userID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"Could not get user."})
	}

	if !exists {
		return c.JSON(http.StatusNotFound, models.Message{"The user does not exist."})
	}

	// Get the doer options
	doer, err := models.GetCurrentUser(c)
	if err != nil {
		return err
	}

	// Delete it
	err = models.DeleteUserByID(userID, &doer)

	if err != nil {
		if models.IsErrIDCannotBeZero(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"Id cannot be 0"})
		}

		if models.IsErrCannotDeleteLastUser(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"Cannot delete last user."})
		}

		return c.JSON(http.StatusInternalServerError, models.Message{"Could not delete user."})
	}

	return c.JSON(http.StatusOK, models.Message{"success"})
}
