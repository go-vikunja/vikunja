package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

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

	// Check if we have our ID
	id := c.Param("id")
	// Make int
	itemID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
	}

	// Check if the user has the right to delete that list
	user, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	err = models.DeleteListByID(itemID, &user)
	if err != nil {
		if models.IsErrNeedToBeListOwner(err) {
			return c.JSON(http.StatusForbidden, models.Message{"You need to be the list owner to delete a list."})
		}

		if models.IsErrListDoesNotExist(err) {
			return c.JSON(http.StatusNotFound, models.Message{"This list does not exist."})
		}

		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}

	return c.JSON(http.StatusOK, models.Message{"The list was deleted with success."})
}
