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
	"code.vikunja.io/web"
	"github.com/labstack/echo"
	"github.com/op/go-logging"
	"net/http"
)

// CreateWeb is the handler to create an object
func (c *WebHandler) CreateWeb(ctx echo.Context) error {
	// Get our model
	currentStruct := c.EmptyStruct()

	// Get the object & bind params to struct
	if err := ParamBinder(currentStruct, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided.")
	}

	// Validate the struct
	if err := ctx.Validate(currentStruct); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	// Get the user to pass for later checks
	authprovider := ctx.Get("AuthProvider").(*web.Auths)
	currentAuth, err := authprovider.AuthObject(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}

	// Check rights
	if !currentStruct.CanCreate(currentAuth) {
		ctx.Get("LoggingProvider").(*logging.Logger).Noticef("Tried to create while not having the rights for it", currentAuth)
		return echo.NewHTTPError(http.StatusForbidden)
	}

	// Create
	err = currentStruct.Create(currentAuth)
	if err != nil {
		return HandleHTTPError(err, ctx)
	}

	return ctx.JSON(http.StatusCreated, currentStruct)
}
