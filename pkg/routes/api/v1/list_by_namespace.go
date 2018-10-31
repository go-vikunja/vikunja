package v1

import (
	"code.vikunja.io/api/pkg/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

// GetListsByNamespaceID is the web handler to delete a namespace
func GetListsByNamespaceID(c echo.Context) error {
	// swagger:operation GET /namespaces/{namespaceID}/lists namespaces getListsByNamespace
	// ---
	// summary: gets all lists belonging to that namespace
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: namespaceID
	//   in: path
	//   description: ID of the namespace
	//   type: string
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/Namespace"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	// Get our namespace
	namespace, err := getNamespace(c)
	if err != nil {
		if models.IsErrNamespaceDoesNotExist(err) {
			return c.JSON(http.StatusNotFound, models.Message{"Namespace not found."})
		}
		if models.IsErrUserDoesNotHaveAccessToNamespace(err) {
			return c.JSON(http.StatusForbidden, models.Message{"You don't have access to this namespace."})
		}
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occurred."})
	}

	// Get the lists
	lists, err := models.GetListsByNamespaceID(namespace.ID)
	if err != nil {
		if models.IsErrNamespaceDoesNotExist(err) {
			return c.JSON(http.StatusNotFound, models.Message{"Namespace not found."})
		}
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occurred."})
	}
	return c.JSON(http.StatusOK, lists)
}

func getNamespace(c echo.Context) (namespace models.Namespace, err error) {
	// Check if we have our ID
	id := c.Param("namespace")
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
