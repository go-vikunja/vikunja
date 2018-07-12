package v1

import (
	"github.com/labstack/echo"
	"net/http"
)

// GetListByID Adds or updates a new list
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

	return echo.NewHTTPError(http.StatusNotImplemented)
}
