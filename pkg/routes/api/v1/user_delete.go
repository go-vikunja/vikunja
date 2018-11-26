//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes/crud"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

// UserDelete is the handler to delete a user
func UserDelete(c echo.Context) error {

	// TODO: only allow users to allow itself

	id := c.Param("id")

	// Make int
	userID, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"User ID is invalid."})
	}

	// Check if the user exists
	_, err = models.GetUserByID(userID)

	if err != nil {
		if models.IsErrUserDoesNotExist(err) {
			return c.JSON(http.StatusNotFound, models.Message{"The user does not exist."})
		}
		return c.JSON(http.StatusInternalServerError, models.Message{"Could not get user."})
	}

	// Get the doer options
	doer, err := models.GetCurrentUser(c)
	if err != nil {
		return err
	}

	// Delete it
	err = models.DeleteUserByID(userID, &doer)

	if err != nil {
		return crud.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, models.Message{"success"})
}
