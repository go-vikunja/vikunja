package v1

import (
	"code.vikunja.io/api/models"
	"crypto/md5"
	"encoding/hex"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"net/http"
	"time"
)

// Login is the login handler
func Login(c echo.Context) error {
	// swagger:operation POST /login user login
	// ---
	// summary: Logs a user in. Returns a JWT-Token to authenticate requests
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/UserLogin"
	// responses:
	//   "200":
	//     "$ref": "#/responses/Token"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "403":
	//     "$ref": "#/responses/Message"

	u := models.UserLogin{}
	if err := c.Bind(&u); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"Please provide a username and password."})
	}

	// Check user
	user, err := models.CheckUserCredentials(&u)
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
