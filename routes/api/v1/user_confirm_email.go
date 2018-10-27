package v1

import (
	"code.vikunja.io/api/models"
	"code.vikunja.io/api/routes/crud"
	"github.com/labstack/echo"
	"net/http"
)

// UserConfirmEmail is the handler to confirm a user email
func UserConfirmEmail(c echo.Context) error {
	// swagger:operation POST /user/confirm user confirmEmail
	// ---
	// summary: Confirms a users email address
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/EmailConfirm"
	// responses:
	//   "200":
	//     "$ref": "#/responses/Message"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "404":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	// Check for Request Content
	var emailConfirm models.EmailConfirm
	if err := c.Bind(&emailConfirm); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No token provided.")
	}

	err := models.UserEmailConfirm(&emailConfirm)
	if err != nil {
		return crud.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, models.Message{"The email was confirmed successfully."})
}
