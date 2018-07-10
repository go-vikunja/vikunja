package v1

import (
	"git.kolaente.de/konrad/list/models"
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

	// Get the list
	var list *models.List

	if err := c.Bind(&list); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"No list model provided."})
	}

	// Get the namespace ID
	var err error
	list.NamespaceID, err = models.GetIntURLParam("nID", c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Invalid namespace ID."})
	}

	// Get the current user for later checks
	user, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}
	list.Owner = user

	// Get the namespace
	namespace, err := models.GetNamespaceByID(list.NamespaceID)
	if err != nil {
		if models.IsErrNamespaceDoesNotExist(err) {
			return c.JSON(http.StatusNotFound, models.Message{"Namespace not found."})
		}
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	// Check if the user has write acces to that namespace
	err = user.HasNamespaceWriteAccess(&namespace)
	if err != nil {
		if models.IsErrUserDoesNotHaveAccessToNamespace(err) {
			return c.JSON(http.StatusForbidden, models.Message{"You don't have access to this namespace."})
		}
		if models.IsErrUserDoesNotHaveWriteAccessToNamespace(err) {
			return c.JSON(http.StatusForbidden, models.Message{"You don't have write access to this namespace."})
		}
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	// Create the new list
	err = models.CreateOrUpdateList(list)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	return c.JSON(http.StatusOK, list)
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

	// Get the list
	var list *models.List

	if err := c.Bind(&list); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"No list model provided."})
	}

	// Get the list ID
	var err error
	list.ID, err = models.GetIntURLParam("id", c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	// Check if the list exists
	// ID = 0 means new list, no error
	var oldList models.List
	if list.ID != 0 {
		oldList, err = models.GetListByID(list.ID)
		if err != nil {
			if models.IsErrListDoesNotExist(err) {
				return c.JSON(http.StatusBadRequest, models.Message{"The list does not exist."})
			}
			return c.JSON(http.StatusInternalServerError, models.Message{"Could not check if the list exists."})
		}
	}

	// Get the current user for later checks
	user, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}
	list.Owner = user

	// Check if the user owns the list
	// TODO use list function for that
	if user.ID != oldList.Owner.ID {
		return c.JSON(http.StatusForbidden, models.Message{"You cannot edit a list you don't own."})
	}

	// Update the list
	err = models.CreateOrUpdateList(list)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	return c.JSON(http.StatusOK, list)
}
