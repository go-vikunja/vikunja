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
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// CreateWeb is the handler to create an object
func (c *WebHandler) CreateWeb(ctx echo.Context) error {
	// Get our model
	currentStruct := c.EmptyStruct()

	// Get the object & bind params to struct
	if err := ctx.Bind(currentStruct); err != nil {
		config.LoggingProvider.Debugf("Invalid model error. Internal error was: %s", err.Error())
		var he *echo.HTTPError
		if errors.As(err, &he) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid model provided. Error was: %s", he.Message))
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid model provided.")
	}

	// Validate the struct
	if err := ctx.Validate(currentStruct); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	// Get the user to pass for later checks
	currentAuth, err := config.AuthProvider.AuthObject(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}

	// Create the db session
	s := config.SessionFactory()
	defer func() {
		err = s.Close()
		if err != nil {
			config.LoggingProvider.Errorf("Could not close session: %s", err)
		}
	}()

	// Check rights
	canCreate, err := currentStruct.CanCreate(s, currentAuth)
	if err != nil {
		_ = s.Rollback()
		return HandleHTTPError(err)
	}
	if !canCreate {
		_ = s.Rollback()
		config.LoggingProvider.Noticef("Tried to create while not having the rights for it (User: %v)", currentAuth)
		return echo.NewHTTPError(http.StatusForbidden)
	}

	// Create
	err = currentStruct.Create(s, currentAuth)
	if err != nil {
		_ = s.Rollback()
		return HandleHTTPError(err)
	}

	err = s.Commit()
	if err != nil {
		return HandleHTTPError(err)
	}

	err = ctx.JSON(http.StatusCreated, currentStruct)
	if err != nil {
		return HandleHTTPError(err)
	}
	return err
}
