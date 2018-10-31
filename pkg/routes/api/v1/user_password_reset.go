package v1

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes/crud"
	"github.com/labstack/echo"
	"net/http"
)

// UserResetPassword is the handler to change a users password
func UserResetPassword(c echo.Context) error {
	// swagger:operation POST /user/password/reset user updatePassword
	// ---
	// summary: Resets a users password
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/PasswordReset"
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
	var pwReset models.PasswordReset
	if err := c.Bind(&pwReset); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No password provided.")
	}

	err := models.UserPasswordReset(&pwReset)
	if err != nil {
		return crud.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, models.Message{"The password was updated successfully."})
}

// UserRequestResetPasswordToken is the handler to change a users password
func UserRequestResetPasswordToken(c echo.Context) error {
	// swagger:operation POST /user/password/token user requestUpdatePasswordToken
	// ---
	// summary: Requests a token to reset a users password
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/PasswordTokenRequest"
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
	var pwTokenReset models.PasswordTokenRequest
	if err := c.Bind(&pwTokenReset); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No user ID provided.")
	}

	err := models.RequestUserPasswordResetToken(&pwTokenReset)
	if err != nil {
		return crud.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, models.Message{"Token was sent."})
}
