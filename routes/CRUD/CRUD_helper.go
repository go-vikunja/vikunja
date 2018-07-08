package CRUD

import (
	"fmt"
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

// This does web stuff, aka returns json etc. Uses CRUDable Methods to get the data
type CRUDWebHandler struct {
	CObject interface{ models.CRUDable }
}

// This does json, handles the request
func (c *CRUDWebHandler) ReadOneWeb(ctx echo.Context) error {

	// Get the ID
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
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

//
func (c *CRUDWebHandler) ReadAllWeb(ctx echo.Context) error {
	currentUser, err := models.GetCurrentUser(ctx)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.Message{"Could not determine the current user."})
	}

	//c.CObject.IsAdmin()

	lists, err := c.CObject.ReadAll(&currentUser)
	if err != nil {
		fmt.Println(err)
		return ctx.JSON(http.StatusInternalServerError, models.Message{"Could not get."})
	}

	return ctx.JSON(http.StatusOK, lists)
}
