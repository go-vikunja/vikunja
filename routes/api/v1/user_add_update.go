package v1

import (
	"encoding/json"
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
	"strings"
	"fmt"
)

// UserAddOrUpdate is the handler to add a user
func UserAddOrUpdate(c echo.Context) error {

	// TODO: prevent everyone from updating users

	// Check for Request Content
	userFromString := c.FormValue("user")
	var datUser *models.User

	if userFromString == "" {
		// b := new(models.User)
		if err := c.Bind(&datUser); err != nil {
			return c.JSON(http.StatusBadRequest, models.Message{"No user model provided."})
		}
	} else {
		// Decode the JSON
		dec := json.NewDecoder(strings.NewReader(userFromString))
		err := dec.Decode(&datUser)

		if err != nil {
			return c.JSON(http.StatusBadRequest, models.Message{"Error decoding user: " + err.Error()})
		}
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

	fmt.Println(exists)

	// Insert or update the user
	var newUser models.User
	if exists {
		newUser, err = models.UpdateUser(*datUser)
	} else {
		newUser, err = models.CreateUser(*datUser)
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

	// Obfuscate his password
	newUser.Password = ""

	return c.JSON(http.StatusOK, newUser)
}
