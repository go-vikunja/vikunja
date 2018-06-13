package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

// AddOrUpdateList Adds or updates a new list
func GetListByID(c echo.Context) error {
	// swagger:operation GET /lists/{listID} lists getList
	// ---
	// summary: gets one list with all todo items
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: listID
	//   in: path
	//   description: ID of the list to show
	//   type: string
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/List"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	// Check if we have our ID
	id := c.Param("id")
	// Make int
	listID, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	// Get the list
	list, err := models.GetListByID(listID)
	if err != nil {
		if models.IsErrListDoesNotExist(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"The list does not exist."})
		}

		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	return c.JSON(http.StatusOK, list)
}
