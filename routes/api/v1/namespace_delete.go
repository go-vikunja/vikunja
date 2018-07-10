package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
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

	// Check if we have our ID
	id := c.Param("id")
	// Make int
	itemID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	// Check if the user has the right to delete that namespace
	user, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	err = models.DeleteNamespaceByID(itemID, &user)
	if err != nil {
		if models.IsErrNeedToBeNamespaceOwner(err) {
			return c.JSON(http.StatusForbidden, models.Message{"You need to be the namespace owner to delete a namespace."})
		}

		if models.IsErrNamespaceDoesNotExist(err) {
			return c.JSON(http.StatusNotFound, models.Message{"This namespace does not exist."})
		}

		if models.IsErrUserNeedsToBeNamespaceAdmin(err) {
			return c.JSON(http.StatusForbidden, models.Message{"You need to be namespace admin to delete a namespace."})
		}

		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	return c.JSON(http.StatusOK, models.Message{"The namespace was deleted with success."})
}
