package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

func AddList(c echo.Context) error {
	// swagger:operation PUT /lists lists addList
	// ---
	// summary: Creates a new list owned by the currently logged in user
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
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

	return addOrUpdateList(c)
}

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

	return addOrUpdateList(c)
}

// AddOrUpdateList Adds or updates a new list
func addOrUpdateList(c echo.Context) error {

	// Get the list
	var list *models.List

	if err := c.Bind(&list); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"No list model provided."})
	}

	// Check if we have an ID other than the one in the struct
	id := c.Param("id")
	if id != "" {
		// Make int
		listID, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
		}
		list.ID = listID
	}

	// Check if the list exists
	// ID = 0 means new list, no error
	if list.ID != 0 {
		_, err := models.GetListByID(list.ID)
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

	// update or create...
	if list.ID == 0 {
		err = models.CreateOrUpdateList(list)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
		}
	} else {
		// Check if the user owns the list
		oldList, err := models.GetListByID(list.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
		}
		if user.ID != oldList.Owner.ID {
			return c.JSON(http.StatusForbidden, models.Message{"You cannot edit a list you don't own."})
		}

		err = models.CreateOrUpdateList(list)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
		}
	}

	return c.JSON(http.StatusOK, list)
}
