// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type message struct {
	Message string `json:"message"`
}

// DeleteWeb is the web handler to delete something
func (c *WebHandler) DeleteWeb(ctx echo.Context) error {

	// Get our model
	currentStruct := c.EmptyStruct()

	// Bind params to struct
	if err := ctx.Bind(currentStruct); err != nil {
		config.LoggingProvider.Debugf("Invalid model error. Internal error was: %s", err.Error())
		if he, is := err.(*echo.HTTPError); is {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid model provided. Error was: %s", he.Message))
		}
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid model provided."))
	}

	// Check if the user has the right to delete
	currentAuth, err := config.AuthProvider.AuthObject(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	// Create the db session
	s := config.SessionFactory()
	defer func() {
		err = s.Close()
		if err != nil {
			config.LoggingProvider.Errorf("Could not close session: %s", err)
		}
	}()

	canDelete, err := currentStruct.CanDelete(s, currentAuth)
	if err != nil {
		_ = s.Rollback()
		return HandleHTTPError(err)
	}
	if !canDelete {
		_ = s.Rollback()
		config.LoggingProvider.Noticef("Tried to delete while not having the rights for it (User: %v)", currentAuth)
		return echo.NewHTTPError(http.StatusForbidden)
	}

	err = currentStruct.Delete(s, currentAuth)
	if err != nil {
		_ = s.Rollback()
		return HandleHTTPError(err)
	}

	err = s.Commit()
	if err != nil {
		return HandleHTTPError(err)
	}

	err = ctx.JSON(http.StatusOK, message{"Successfully deleted."})
	if err != nil {
		return HandleHTTPError(err)
	}
	return err
}
