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

// DeleteListByID ...
func DeleteListByID(c echo.Context) error {
	// swagger:operation DELETE /lists/{listID} lists deleteList
	// ---
	// summary: Deletes a list with all items on it
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: listID
	//   in: path
	//   description: ID of the list to delete
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

// GetListByID Adds or updates a new list
func GetListByID(c echo.Context) error {
	// swagger:operation GET /lists/{listID} lists getList
	// ---
	// summary: gets one list with all todo items
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: listID
	//   in: path
	//   description: ID of the list to show
	//   type: string
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/List"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	return echo.NewHTTPError(http.StatusNotImplemented)
}

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

	return echo.NewHTTPError(http.StatusNotImplemented)
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

	return echo.NewHTTPError(http.StatusNotImplemented)
}

// GetListsByUser gets all lists a user owns
func GetListsByUser(c echo.Context) error {
	// swagger:operation GET /lists lists getLists
	// ---
	// summary: Gets all lists owned by the current user
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// responses:
	//   "200":
	//     "$ref": "#/responses/List"
	//   "500":
	//     "$ref": "#/responses/Message"

	return echo.NewHTTPError(http.StatusNotImplemented)
}

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

// GetAllNamespacesByCurrentUser ...
func GetAllNamespacesByCurrentUser(c echo.Context) error {
	// swagger:operation GET /namespaces namespaces getNamespaces
	// ---
	// summary: Get all namespaces the currently logged in user has at least read access
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// responses:
	//   "200":
	//     "$ref": "#/responses/Namespace"
	//   "500":
	//     "$ref": "#/responses/Message"

	return echo.NewHTTPError(http.StatusNotImplemented)
}
