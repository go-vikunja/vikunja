package v1

import (
	"github.com/labstack/echo"
	"net/http"
)

// DeleteNamespaceByID ...
func DeleteNamespaceByID(c echo.Context) error {
	// swagger:operation DELETE /namespaces/{namespaceID} namespaces deleteNamespace
	// ---
	// summary: Deletes a namespace with all lists
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: namespaceID
	//   in: path
	//   description: ID of the namespace to delete
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
