package v1

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes/crud"
	"github.com/labstack/echo"
	"net/http"
)

// RegisterUser is the register handler
// @Summary Register
// @Description Creates a new user account.
// @tags user
// @Accept json
// @Produce json
// @Param credentials body models.APIUserPassword true "The user credentials"
// @Success 200 {object} models.User
// @Failure 400 {object} models.HTTPError "No or invalid user register object provided / User already exists."
// @Failure 500 {object} models.Message "Internal error"
// @Router /register [post]
func RegisterUser(c echo.Context) error {
	// Check for Request Content
	var datUser *models.APIUserPassword
	if err := c.Bind(&datUser); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"No or invalid user model provided."})
	}

	// Insert the user
	newUser, err := models.CreateUser(datUser.APIFormat())
	if err != nil {
		return crud.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, newUser)
}
