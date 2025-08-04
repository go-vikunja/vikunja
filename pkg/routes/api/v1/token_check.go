// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"net/http"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// CheckToken checks prints a message if the token is valid or not. Currently only used for testing purposes.
func CheckToken(c echo.Context) error {

	user := c.Get("user").(*jwt.Token)

	log.Debugf("token valid: %t", user.Valid)

	return c.JSON(418, models.Message{Message: "🍵"})
}

// TestToken returns a simple test message. Used for testing purposes.
func TestToken(c echo.Context) error {
	return c.JSON(http.StatusOK, models.Message{Message: "ok"})
}
