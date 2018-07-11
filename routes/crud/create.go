package crud

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
)

// CreateWeb is the handler to create an object
func (c *WebHandler) CreateWeb(ctx echo.Context) error {
	// Get the object
	if err := ctx.Bind(&c.CObject); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided.")
	}

	// Get the user to pass for later checks
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}

	// Get an ID if we have one
	var id int64
	if ctx.Param("id") != "" {
		id, err = models.GetIntURLParam("id", ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Bad id.")
		}
	}

	// Create
	err = c.CObject.Create(&currentUser, id)
	if err != nil {
		if models.IsErrListDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "The list does not exist.")
		}
		if models.IsErrListItemCannotBeEmpty(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "You must provide at least a list item text.")
		}
		if models.IsErrUserDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "The user does not exist.")
		}
		if models.IsErrNeedToBeListWriter(err) {
			return echo.NewHTTPError(http.StatusForbidden, "You need to have write access on that list.")
		}

		if models.IsErrNamespaceNameCannotBeEmpty(err) {
			return echo.NewHTTPError(http.StatusNotFound, "The namespace name cannot be empty.")
		}

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, c.CObject)
}
