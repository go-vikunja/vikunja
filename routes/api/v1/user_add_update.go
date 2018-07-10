package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

// RegisterUser ...
func RegisterUser(c echo.Context) error {

	// swagger:operation POST /register user register
	// ---
	// summary: Creates a new user account
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/APIUserPassword"
	// responses:
	//   "200":
	//     "$ref": "#/responses/User"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	return userAddOrUpdate(c)
}

// userAddOrUpdate is the handler to add a user
func userAddOrUpdate(c echo.Context) error {

	// TODO: prevent everyone from updating users

	// Check for Request Content
	var datUser *models.APIUserPassword

	if err := c.Bind(&datUser); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"No user model provided."})
	}

	// Check if we have an ID other than the one in the struct
	id := c.Param("id")
	if id != "" {
		// Make int
		userID, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
		}
		datUser.ID = userID
	}

	// Check if the user exists
	_, exists, err := models.GetUserByID(datUser.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"Could not check if the user exists."})
	}

	// Insert or update the user
	var newUser models.User
	if exists {
		newUser, err = models.UpdateUser(datUser.APIFormat())
	} else {
		newUser, err = models.CreateUser(datUser.APIFormat())
	}

	if err != nil {
		// Check for user already exists
		if models.IsErrUsernameExists(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"A user with this username already exists."})
		}

		// Check for user with that email already exists
		if models.IsErrUserEmailExists(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"A user with this email address already exists."})
		}

		// Check for no username provided
		if models.IsErrNoUsername(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"Please specify a username."})
		}

		// Check for no username or password provided
		if models.IsErrNoUsernamePassword(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"Please specify a username and a password."})
		}

		// Check for user does not exist
		if models.IsErrUserDoesNotExist(err) {
			return c.JSON(http.StatusBadRequest, models.Message{"The user does not exist."})
		}

		return c.JSON(http.StatusInternalServerError, models.Message{"Error"})
	}

	return c.JSON(http.StatusOK, newUser)
}
