package v1

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes/crud"
	"github.com/labstack/echo"
	"net/http"
)

// UserConfirmEmail is the handler to confirm a user email
// @Summary Confirm the email of a new user
// @Description Confirms the email of a newly registered user.
// @tags user
// @Accept json
// @Produce json
// @Param credentials body models.EmailConfirm true "The token."
// @Success 200 {object} models.Message
// @Failure 412 {object} models.HTTPError "Bad token provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /user/confirm [post]
func UserConfirmEmail(c echo.Context) error {
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
