package CRUD

import (
	"fmt"
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
)

// This does web stuff, aka returns json etc. Uses CRUDable Methods to get the data
type CRUDWebHandler struct {
	CObject interface{
		models.CRUDable
		models.Rights
	}
}

// This does json, handles the request
func (c *CRUDWebHandler) ReadOneWeb(ctx echo.Context) error {

	// Get the ID
	id, err := models.GetIntURLParam("id", ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	// TODO check rights

	// Get our object
	err = c.CObject.ReadOne(id)
	if err != nil {
		if models.IsErrListDoesNotExist(err) {
			return ctx.JSON(http.StatusNotFound, models.Message{"Not found."})
		}

		return ctx.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	return ctx.JSON(http.StatusOK, c.CObject)
}

// ReadAllWeb returns all elements of a type
func (c *CRUDWebHandler) ReadAllWeb(ctx echo.Context) error {
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.Message{"Could not determine the current user."})
	}

	lists, err := c.CObject.ReadAll(&currentUser)
	if err != nil {
		fmt.Println(err)
		return ctx.JSON(http.StatusInternalServerError, models.Message{"Could not get."})
	}

	return ctx.JSON(http.StatusOK, lists)
}

// UpdateWeb is the webhandler to update an object
func (c *CRUDWebHandler) UpdateWeb(ctx echo.Context) error {
	// Get the object
	if err := ctx.Bind(&c.CObject); err != nil {
		return ctx.JSON(http.StatusBadRequest, models.Message{"No model provided."})
	}

	// Get the ID
	id, err := models.GetIntURLParam("id", ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	// Check if the user has the right to do that
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.Message{"Could not determine the current user."})
	}

	// Do the update
	err = c.CObject.Update(id, &currentUser)
	if err != nil {
		if models.IsErrNeedToBeListAdmin(err) {
			return echo.NewHTTPError(http.StatusForbidden, "You need to be list admin to do that.")
		}

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, c.CObject)
}

// CreateWeb is the handler to create an object
func (c *CRUDWebHandler) CreateWeb(ctx echo.Context) error {
	// Get the object
	if err := ctx.Bind(&c.CObject); err != nil {
		return ctx.JSON(http.StatusBadRequest, models.Message{"No model provided."})
	}

	// Get the user to pass for later checks
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}

	// Create
	err = c.CObject.Create(&currentUser)
	if err != nil {
		fmt.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, c.CObject)
}

// DeleteWeb is the web handler to delete something
func (c *CRUDWebHandler) DeleteWeb(ctx echo.Context) error {
	// Get the ID
	id, err := models.GetIntURLParam("id", ctx)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	// Check if the user has the right to delete
	user, err := models.GetCurrentUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	err = c.CObject.Delete(id, &user)
	if err != nil {
		if models.IsErrNeedToBeListAdmin(err) {
			return echo.NewHTTPError(http.StatusForbidden, "You need to be the list admin to delete a list.")
		}

		if models.IsErrListDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound, "This list does not exist.")
		}

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, models.Message{"Successfully deleted."})
}