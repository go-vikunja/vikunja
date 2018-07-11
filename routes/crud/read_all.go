package crud

import (
	"fmt"
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
)

// ReadAllWeb is the webhandler to get all objects of a type
func (c *WebHandler) ReadAllWeb(ctx echo.Context) error {
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
