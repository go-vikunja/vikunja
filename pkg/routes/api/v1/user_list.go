package v1

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes/crud"
	"github.com/labstack/echo"
	"net/http"
)

// UserList gets all information about a user
func UserList(c echo.Context) error {

	// swagger:operation GET /users user list
	// ---
	// summary: Lists all users
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: s
	//   description: A searchterm to search for a user by its username
	//   in: query
	// responses:
	//   "200":
	//     "$ref": "#/responses/User"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

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
