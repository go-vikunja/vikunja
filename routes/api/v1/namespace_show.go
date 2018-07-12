package v1

import (
	"github.com/labstack/echo"
	"net/http"
)

// ShowNamespace ...
func ShowNamespace(c echo.Context) error {
	// swagger:operation GET /namespaces/{namespaceID} namespaces getNamespace
	// ---
	// summary: gets one namespace with all todo items
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: namespaceID
	//   in: path
	//   description: ID of the namespace to show
	//   type: string
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/Namespace"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	return echo.NewHTTPError(http.StatusNotImplemented)
}
