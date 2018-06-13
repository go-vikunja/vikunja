package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

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

	// TODO: return 403 if you dont have the right to add an item to that list

	// Get the list ID
	id := c.Param("id")
	// Make int
	listID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	return updateOrCreateListItemHelper(listID, 0, c)
}

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

	// Get the item ID
	id := c.Param("id")
	// Make int
	itemID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	return updateOrCreateListItemHelper(0, itemID, c)
}

func updateOrCreateListItemHelper(listID, itemID int64, c echo.Context) error {

	// Get the list item
	var listItem *models.ListItem

	if err := c.Bind(&listItem); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"No list model provided."})
	}

	// Creating
	if listID != 0 {
		listItem.ListID = listID

		// Set the user
		user, err := models.GetCurrentUser(c)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
		}
		listItem.CreatedBy = user
	}

	// Updating
	if itemID != 0 {
		listItem.ID = itemID
	}

	finalItem, err := models.CreateOrUpdateListItem(listItem)
	if err != nil {
		if models.IsErrListDoesNotExist(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"The list does not exist."})
		}
		if models.IsErrListItemCannotBeEmpty(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"You must provide at least a list item text."})
		}
		if models.IsErrUserDoesNotExist(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"The user does not exist."})
		}

		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	return c.JSON(http.StatusOK, finalItem)
}
