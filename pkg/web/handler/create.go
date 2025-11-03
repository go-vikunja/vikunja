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

package handler

import (
	"errors"
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v4"
)

// CreateWeb is the handler to create an object
func (c *WebHandler) CreateWeb(ctx echo.Context) error {
	// Get our model
	currentStruct := c.EmptyStruct()

	// Get the object & bind params to struct
	if err := ctx.Bind(currentStruct); err != nil {
		log.Debugf("Invalid model error. Internal error was: %s", err.Error())
		var he *echo.HTTPError
		if errors.As(err, &he) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid model provided. Error was: %s", he.Message)).SetInternal(err)
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid model provided.").SetInternal(err)
	}

	// Validate the struct
	if err := ctx.Validate(currentStruct); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err).SetInternal(err)
	}

	// Get the user to pass for later checks
	currentAuth, err := auth.GetAuthFromClaims(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.").SetInternal(err)
	}

	// Create the db session
	s := db.NewSession()
	defer func() {
		err = s.Close()
		if err != nil {
			log.Errorf("Could not close session: %s", err)
		}
	}()

	// Check permissions
	canCreate, err := currentStruct.CanCreate(s, currentAuth)
	if err != nil {
		_ = s.Rollback()
		return HandleHTTPError(err)
	}
	if !canCreate {
		_ = s.Rollback()
		log.Warningf("Tried to create while not having the permissions for it (User: %v)", currentAuth)
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
