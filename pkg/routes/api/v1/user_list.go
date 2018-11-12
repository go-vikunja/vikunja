package v1

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes/crud"
	"github.com/labstack/echo"
	"net/http"
)

// UserList gets all information about a user
// @Summary Get users
// @Description Lists all users (without emailadresses). Also possible to search for a specific user.
// @tags user
// @Accept json
// @Produce json
// @Param s query string false "Search for a user by its name."
// @Security ApiKeyAuth
// @Success 200 {array} models.User "All (found) users."
// @Failure 400 {object} models.HTTPError "Something's invalid."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /users [get]
func UserList(c echo.Context) error {
	s := c.QueryParam("s")
	users, err := models.ListUsers(s)
	if err != nil {
		return crud.HandleHTTPError(err)
	}

	// Obfuscate the mailadresses
	for in := range users {
		users[in].Email = ""
	}

	return c.JSON(http.StatusOK, users)
}
