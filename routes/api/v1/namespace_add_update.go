package v1

import (
	"github.com/labstack/echo"
	"net/http"
)

// AddNamespace ...
func AddNamespace(c echo.Context) error {
	// swagger:operation PUT /namespaces namespaces addNamespace
	// ---
	// summary: Creates a new namespace owned by the currently logged in user
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/Namespace"
	// responses:
	//   "200":
	//     "$ref": "#/responses/Namespace"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "403":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	return echo.NewHTTPError(http.StatusNotImplemented)
}

// UpdateNamespace ...
func UpdateNamespace(c echo.Context) error {
	// swagger:operation POST /namespaces/{namespaceID} namespaces upadteNamespace
	// ---
	// summary: Updates a namespace
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: namespaceID
	//   in: path
	//   description: ID of the namespace to update
	//   type: string
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/Namespace"
	// responses:
	//   "200":
	//     "$ref": "#/responses/Namespace"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "403":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	return echo.NewHTTPError(http.StatusNotImplemented)
}
