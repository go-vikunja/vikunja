package v1

import (
	"github.com/labstack/echo"
	"net/http"
)

// AddList ...
func AddList(c echo.Context) error {
	// swagger:operation PUT /namespaces/{namespaceID}/lists lists addList
	// ---
	// summary: Creates a new list owned by the currently logged in user in that namespace
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: namespaceID
	//   in: path
	//   description: ID of the namespace that list should belong to
	//   type: string
	//   required: true
	// - name: body
	//   in: body
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/List"
	// responses:
	//   "200":
	//     "$ref": "#/responses/List"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "403":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	return echo.NewHTTPError(http.StatusNotImplemented)
}

// UpdateList ...
func UpdateList(c echo.Context) error {
	// swagger:operation POST /lists/{listID} lists upadteList
	// ---
	// summary: Updates a list
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: listID
	//   in: path
	//   description: ID of the list to update
	//   type: string
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/List"
	// responses:
	//   "200":
	//     "$ref": "#/responses/List"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "403":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	return echo.NewHTTPError(http.StatusNotImplemented)
}
