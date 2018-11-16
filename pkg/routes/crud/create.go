package crud

import (
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"github.com/labstack/echo"
	"net/http"
)

// CreateWeb is the handler to create an object
func (c *WebHandler) CreateWeb(ctx echo.Context) error {
	// Get our model
	currentStruct := c.EmptyStruct()

	// Get the object & bind params to struct
	if err := ParamBinder(currentStruct, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided.")
	}

	// Validate the struct
	if err := ctx.Validate(currentStruct); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	// Get the user to pass for later checks
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}

	// Check rights
	if !currentStruct.CanCreate(&currentUser) {
		log.Log.Noticef("%s [ID: %d] tried to create while not having the rights for it", currentUser.Username, currentUser.ID)
		return echo.NewHTTPError(http.StatusForbidden)
	}

	// Create
	err = currentStruct.Create(&currentUser)
	if err != nil {
		return HandleHTTPError(err)
	}

	return ctx.JSON(http.StatusCreated, currentStruct)
}
