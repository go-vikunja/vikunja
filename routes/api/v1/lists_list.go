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

	return c.JSON(http.StatusOK, nil)
}
