package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

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

	// Check if we have our ID
	id := c.Param("id")
	// Make int
	itemID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	// Check if the user has the right to delete that list item
	user, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	err = models.DeleteListItemByID(itemID, &user)
	if err != nil {
		if models.IsErrListItemDoesNotExist(err) {
			return c.JSON(http.StatusNotFound, models.Message{"List item does not exist."})
		}

		if models.IsErrNeedToBeItemOwner(err) {
			return c.JSON(http.StatusForbidden, models.Message{"You need to own the list item in order to be able to delete it."})
		}

		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	return c.JSON(http.StatusOK, models.Message{"The item was deleted with success."})
}
