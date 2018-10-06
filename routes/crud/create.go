package crud

import (
	"code.vikunja.io/api/models"
	"github.com/labstack/echo"
	"net/http"
	"reflect"
)

// CreateWeb is the handler to create an object
func (c *WebHandler) CreateWeb(ctx echo.Context) error {
	// Re-initialize our model
	p := reflect.ValueOf(c.CObject).Elem()
	p.Set(reflect.Zero(p.Type()))

	// Get the object & bind params to struct
	if err := ParamBinder(c.CObject, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided.")
	}

	// Get the user to pass for later checks
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}

	// Check rights
	if !c.CObject.CanCreate(&currentUser) {
		models.Log.Noticef("%s [ID: %d] tried to create while not having the rights for it", currentUser.Username, currentUser.ID)
		return echo.NewHTTPError(http.StatusForbidden)
	}

	// Create
	err = c.CObject.Create(&currentUser)
	if err != nil {
		return HandleHTTPError(err)
	}

	return ctx.JSON(http.StatusCreated, c.CObject)
}
