package v1

import (
	"code.vikunja.io/api/models"
	"github.com/labstack/echo"
	"net/http"
)

type UserPassword struct {
	Password string `json:"password"`
}

// UserChangePassword is the handler to change a users password
func UserChangePassword(c echo.Context) error {
	// swagger:operation POST /user/password user updatePassword
	// ---
	// summary: Shows the current user
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/Password"
	// responses:
	//   "200":
	//     "$ref": "#/responses/Message"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "404":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	// Check if the user is itself
	doer, err := models.GetCurrentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error getting current user.")
	}

	// Check for Request Content
	var newPW UserPassword
	if err := c.Bind(&newPW); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No password provided.")
	}

	// Update the password
	err = models.UpdateUserPassword(&doer, newPW.Password)
	if err != nil {
		if models.IsErrUserDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound, "The user does not exist.")
		}

		models.Log.Error("Error updating a users password, user: %d", doer.ID)
		return echo.NewHTTPError(http.StatusInternalServerError, "An error occurred.")
	}

	return c.JSON(http.StatusOK, models.Message{"The password was updated successfully."})
}
