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
		return echo.NewHTTPError(http.StatusForbidden)
	}

	// Create
	err = c.CObject.Create(&currentUser)
	if err != nil {
		if models.IsErrListDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "The list does not exist.")
		}
		if models.IsErrListTitleCannotBeEmpty(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "You must provide at least a list title.")
		}
		if models.IsErrListTaskCannotBeEmpty(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "You must provide at least a list task text.")
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

		if models.IsErrTeamNameCannotBeEmpty(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "The team name cannot be empty.")
		}

		if models.IsErrTeamAlreadyHasAccess(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "This team already has access.")
		}
		if models.IsErrUserIsMemberOfTeam(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "This user is already a member of that team.")
		}

		if models.IsErrUserAlreadyHasAccess(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "This user already has access to this list.")
		}
		if models.IsErrUserAlreadyHasNamespaceAccess(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "This user already has access to this namespace.")
		}
		if models.IsErrInvalidUserRight(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "The right is invalid.")
		}

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, c.CObject)
}
