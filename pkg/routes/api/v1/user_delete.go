package v1

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes/crud"
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
	_, err = models.GetUserByID(userID)

	if err != nil {
		if models.IsErrUserDoesNotExist(err) {
			return c.JSON(http.StatusNotFound, models.Message{"The user does not exist."})
		}
		return c.JSON(http.StatusInternalServerError, models.Message{"Could not get user."})
	}

	// Get the doer options
	doer, err := models.GetCurrentUser(c)
	if err != nil {
		return err
	}

	// Delete it
	err = models.DeleteUserByID(userID, &doer)

	if err != nil {
		return crud.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, models.Message{"success"})
}
