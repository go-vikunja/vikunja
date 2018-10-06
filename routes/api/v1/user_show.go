package v1

import (
	"code.vikunja.io/api/models"
	"code.vikunja.io/api/routes/crud"
	"github.com/labstack/echo"
	"net/http"
)

// UserShow gets all informations about the current user
func UserShow(c echo.Context) error {
	// swagger:operation GET /user user showUser
	// ---
	// summary: Shows the current user
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// responses:
	//   "200":
	//     "$ref": "#/responses/User"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	userInfos, err := models.GetCurrentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error getting current user.")
	}

	user, err := models.GetUserByID(userInfos.ID)
	if err != nil {
		return crud.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, user)
}
