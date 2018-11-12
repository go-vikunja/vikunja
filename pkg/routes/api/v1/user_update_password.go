package v1

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes/crud"
	"github.com/labstack/echo"
	"net/http"
)

// UserPassword holds a user password. Used to update it.
type UserPassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// UserChangePassword is the handler to change a users password
// @Summary Change password
// @Description Lets the current user change its password.
// @tags user
// @Accept json
// @Produce json
// @Param userPassword body v1.UserPassword true "The current and new password."
// @Security ApiKeyAuth
// @Success 200 {object} models.Message
// @Failure 400 {object} models.HTTPError "Something's invalid."
// @Failure 404 {object} models.HTTPError "User does not exist."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/password [post]
func UserChangePassword(c echo.Context) error {
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

	// Check the current password
	if _, err = models.CheckUserCredentials(&models.UserLogin{Username: doer.Username, Password: newPW.OldPassword}); err != nil {
		return crud.HandleHTTPError(err)
	}

	// Update the password
	if err = models.UpdateUserPassword(&doer, newPW.NewPassword); err != nil {
		return crud.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, models.Message{"The password was updated successfully."})
}
