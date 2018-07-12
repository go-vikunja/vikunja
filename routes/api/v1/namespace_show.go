package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
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

func getNamespace(c echo.Context) (namespace models.Namespace, err error) {
	// Check if we have our ID
	id := c.Param("id")
	// Make int
	namespaceID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}

	// Get the namespace
	namespace, err = models.GetNamespaceByID(namespaceID)
	if err != nil {
		return
	}

	// Check if the user has acces to that namespace
	user, err := models.GetCurrentUser(c)
	if err != nil {
		return
	}
	if !namespace.CanRead(&user) {
		return
	}

	return
}
