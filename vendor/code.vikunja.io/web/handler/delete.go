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
	"github.com/labstack/echo/v4"
	"net/http"
)

type message struct {
	Message string `json:"message"`
}

// DeleteWeb is the web handler to delete something
func (c *WebHandler) DeleteWeb(ctx echo.Context) error {

	// Get our model
	currentStruct := c.EmptyStruct()

	// Bind params to struct
	if err := ParamBinder(currentStruct, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid URL param.")
	}

	// Check if the user has the right to delete
	currentAuth, err := config.AuthProvider.AuthObject(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	canDelete, err := currentStruct.CanDelete(currentAuth)
	if err != nil {
		return HandleHTTPError(err, ctx)
	}
	if !canDelete {
		config.LoggingProvider.Noticef("Tried to delete while not having the rights for it (User: %v)", currentAuth)
		return echo.NewHTTPError(http.StatusForbidden)
	}

	err = currentStruct.Delete()
	if err != nil {
		return HandleHTTPError(err, ctx)
	}

	return ctx.JSON(http.StatusOK, message{"Successfully deleted."})
}
