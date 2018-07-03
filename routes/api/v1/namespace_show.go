package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

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

	// Check if we have our ID
	id := c.Param("id")
	// Make int
	namespaceID, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	// Get the namespace
	namespace, err := models.GetNamespaceByID(namespaceID)
	if err != nil {
		if models.IsErrNamespaceDoesNotExist(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"The namespace does not exist."})
		}
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	// Check if the user has acces to that namespace
	user, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	has, err := user.HasNamespaceAccess(&namespace)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}
	if !has {
		return c.JSON(http.StatusForbidden, models.Message{"You don't have access to this namespace."})
	}

	return c.JSON(http.StatusOK, namespace)
}
