package crud

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"fmt"
)

// ReadOneWeb is the webhandler to get one object
func (c *WebHandler) ReadOneWeb(ctx echo.Context) error {

	// Get the ID
	id, err := models.GetIntURLParam("id", ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID.")
	}

	// TODO check rights
	//c.CObject.CanRead(doer)

	// Get our object
	err = c.CObject.ReadOne(id)
	if err != nil {
		if models.IsErrListDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		if models.IsErrNamespaceDoesNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		fmt.Println(err)

		return echo.NewHTTPError(http.StatusInternalServerError, "An error occured.")
	}

	return ctx.JSON(http.StatusOK, c.CObject)
}
