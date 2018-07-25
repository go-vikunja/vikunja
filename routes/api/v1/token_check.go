package v1

import (
	"fmt"
	"code.vikunja.io/api/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// CheckToken checks prints a message if the token is valid or not. Currently only used for testing pourposes.
func CheckToken(c echo.Context) error {

	user := c.Get("user").(*jwt.Token)

	fmt.Println(user.Valid)

	return c.JSON(418, models.Message{"🍵"})
}
