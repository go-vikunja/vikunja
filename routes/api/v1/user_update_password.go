package v1

import (
	"net/http"
	"strconv"

	"code.vikunja.io/api/models"
	"github.com/labstack/echo"
)

type datPassword struct {
	Password string `json:"password"`
}

// UserChangePassword is the handler to add a user
func UserChangePassword(c echo.Context) error {

	// Get the ID
	user := c.Param("id")

	if user == "" {
		return c.JSON(http.StatusBadRequest, models.Message{"User ID cannot be empty."})
	}

	// Make int
	userID, err := strconv.ParseInt(user, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"User ID is invalid."})
	}

	// Check if the user is itself
	userJWTinfo, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"Error getting current user."})
	}

	if userJWTinfo.ID != userID {
		return echo.ErrUnauthorized
	}

	// Check for Request Content
	pwFromString := c.FormValue("password")
	var datPw datPassword

	if pwFromString == "" {
		if err := c.Bind(&datPw); err != nil {
			return c.JSON(http.StatusBadRequest, models.Message{"No password provided."})
		}
	} else {
		// Take the value directly from the input
		datPw.Password = pwFromString
	}

	// Get User Infos
	_, err = models.GetUserByID(userID)

	if err != nil {
		if models.IsErrUserDoesNotExist(err) {
			return c.JSON(http.StatusNotFound, models.Message{"The user does not exist."})
		}
		return c.JSON(http.StatusInternalServerError, models.Message{"Error getting user infos."})
	}

	// Get the doer options
	doer, err := models.GetCurrentUser(c)
	if err != nil {
		return err
	}

	err = models.UpdateUserPassword(userID, datPw.Password, &doer)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{"The password was updated successfully"})
}
