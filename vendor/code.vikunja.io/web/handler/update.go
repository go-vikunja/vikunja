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

// UpdateWeb is the webhandler to update an object
func (c *WebHandler) UpdateWeb(ctx echo.Context) error {

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

	// Check if the user has the right to do that
	authprovider := ctx.Get("AuthProvider").(*web.Auths)
	currentAuth, err := authprovider.AuthObject(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}
	if !currentStruct.CanUpdate(currentAuth) {
		ctx.Get("LoggingProvider").(*logging.Logger).Noticef("Tried to update while not having the rights for it", currentAuth)
		return echo.NewHTTPError(http.StatusForbidden)
	}

	// Do the update
	err = currentStruct.Update()
	if err != nil {
		return HandleHTTPError(err, ctx)
	}

	return ctx.JSON(http.StatusOK, currentStruct)
}
