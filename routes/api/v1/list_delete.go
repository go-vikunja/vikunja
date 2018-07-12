package v1

import (
	//	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	//	"net/http"
	//	"strconv"
	"net/http"
)

// DeleteListByID ...
func DeleteListByID(c echo.Context) error {
	// swagger:operation DELETE /lists/{listID} lists deleteList
	// ---
	// summary: Deletes a list with all items on it
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: listID
	//   in: path
	//   description: ID of the list to delete
	//   type: string
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/Message"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "403":
	//     "$ref": "#/responses/Message"
	//   "404":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	return echo.NewHTTPError(http.StatusNotImplemented)
}
