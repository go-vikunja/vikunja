package crud

import (
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

// ReadAllWeb is the webhandler to get all objects of a type
func (c *WebHandler) ReadAllWeb(ctx echo.Context) error {
	// Get our model
	currentStruct := c.EmptyStruct()

	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}

	// Get the object & bind params to struct
	if err := ParamBinder(currentStruct, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided.")
	}

	// Pagination
	page := ctx.QueryParam("page")
	if page == "" {
		page = "1"
	}
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		log.Log.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Bad page requested.")
	}
	if pageNumber < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad page requested.")
	}

	// Search
	search := ctx.QueryParam("s")

	lists, err := currentStruct.ReadAll(search, &currentUser, pageNumber)
	if err != nil {
		return HandleHTTPError(err)
	}

	return ctx.JSON(http.StatusOK, lists)
}
