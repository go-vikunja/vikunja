//  Copyright (c) 2018 Vikunja and contributors.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Lesser General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Lesser General Public License for more details.
//
//  You should have received a copy of the GNU Lesser General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package handler

import (
	"github.com/labstack/echo"
	"net/http"
)

// ReadOneWeb is the webhandler to get one object
func (c *WebHandler) ReadOneWeb(ctx echo.Context) error {
	// Get our model
	currentStruct := c.EmptyStruct()

	// Get the object & bind params to struct
	if err := ParamBinder(currentStruct, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided.")
	}

	// Get our object
	err := currentStruct.ReadOne()
	if err != nil {
		return HandleHTTPError(err, ctx)
	}

	// Check rights
	// We can only check the rights on a full object, which is why we need to check it afterwards
	currentAuth, err := config.AuthProvider.AuthObject(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}
	canRead, err := currentStruct.CanRead(currentAuth)
	if err != nil {
		return HandleHTTPError(err, ctx)
	}
	if canRead {
		config.LoggingProvider.Noticef("Tried to create while not having the rights for it (User: %v)", currentAuth)
		return echo.NewHTTPError(http.StatusForbidden, "You don't have the right to see this")
	}

	return ctx.JSON(http.StatusOK, currentStruct)
}
