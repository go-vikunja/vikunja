package v1

import (
	"crypto/md5"
	"encoding/hex"
	"git.kolaente.de/konrad/list/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"net/http"
	"time"
)

// Login is the login handler
func Login(c echo.Context) error {
	u := new(models.UserLogin)
	if err := c.Bind(u); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Please provide a username and password."})
	}

	// Check user
	user, err := models.CheckUserCredentials(u)

	if err != nil {
		return c.JSON(http.StatusUnauthorized, models.Message{"Wrong username or password."})
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["email"] = user.Email
	claims["id"] = user.ID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	avatar := md5.Sum([]byte(user.Email))
	claims["avatar"] = hex.EncodeToString(avatar[:])

	// Generate encoded token and send it as response.
	t, err := token.SignedString(models.Config.JWTLoginSecret)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": t,
	})
}
