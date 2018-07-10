package crud

import (
	"github.com/labstack/echo"
	"git.kolaente.de/konrad/list/models"
	"net/http"
)

// ReadOneWeb is the webhandler to get one object
func (c *WebHandler) ReadOneWeb(ctx echo.Context) error {

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
