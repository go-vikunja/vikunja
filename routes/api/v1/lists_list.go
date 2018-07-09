package v1

import (
	"github.com/labstack/echo"
	"net/http"
)

// GetListsByUser gets all lists a user owns
func GetListsByUser(c echo.Context) error {
	// swagger:operation GET /lists lists getLists
	// ---
	// summary: Gets all lists owned by the current user
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// responses:
	//   "200":
	//     "$ref": "#/responses/List"
	//   "500":
	//     "$ref": "#/responses/Message"

	/*currentUser, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"Could not determine the current user."})
	}

	allLists, err := models.GetListsByUser(&currentUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"Could not get lists."})
	}*/

	return c.JSON(http.StatusOK, nil)
}
