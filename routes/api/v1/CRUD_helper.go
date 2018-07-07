package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

// Basic Method definitions
type CRUD interface {
	Create()
	Read(int64) (error)
	Update()
	Delete()
}

// We use this to acces the default methods
type DefaultCRUD struct {
	CRUD
	Target interface{}
}

// This method gets our data, which will be called by ReadWeb()
func (d *DefaultCRUD) Read(id int64) (err error) {
	return models.GetByID(id, d.Target)
}

// This does web stuff, aka returns json etc. Uses DefaultCRUD Methods to get the data
type CRUDWebHandler struct {
	CObject CRUD
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
	err = c.CObject.Read(id)
	if err != nil {
		if models.IsErrListDoesNotExist(err) {
			return ctx.JSON(http.StatusNotFound, models.Message{"Not found."})
		}

		return ctx.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	// TODO how can we return c.CObject.Targetdirectly?
	return ctx.JSON(http.StatusOK, c.CObject)
}
