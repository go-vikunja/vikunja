package v1

import (
	"github.com/labstack/echo"
	"net/http"
)

// DeleteListItemByIDtemByID is the web handler to delete a list item
func DeleteListItemByIDtemByID(c echo.Context) error {
	// swagger:operation DELETE /item/{itemID} lists deleteListItem
	// ---
	// summary: Deletes a list item
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: itemID
	//   in: path
	//   description: ID of the list item to delete
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
