package v1

import (
	"github.com/labstack/echo"
	"net/http"
)

// AddListItem ...
func AddListItem(c echo.Context) error {
	// swagger:operation PUT /lists/{listID} lists addListItem
	// ---
	// summary: Adds an item to a list
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: listID
	//   in: path
	//   description: ID of the list to use
	//   type: string
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/ListItem"
	// responses:
	//   "200":
	//     "$ref": "#/responses/ListItem"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	return echo.NewHTTPError(http.StatusNotImplemented)
}

// UpdateListItem ...
func UpdateListItem(c echo.Context) error {
	// swagger:operation PUT /item/{itemID} lists updateListItem
	// ---
	// summary: Updates a list item
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: itemID
	//   in: path
	//   description: ID of the item to update
	//   type: string
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/ListItem"
	// responses:
	//   "200":
	//     "$ref": "#/responses/ListItem"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	return echo.NewHTTPError(http.StatusNotImplemented)
}