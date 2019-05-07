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
	"github.com/labstack/echo/v4"
	"net/http"
)

// WebHandler defines the webhandler object
// This does web stuff, aka returns json etc. Uses CRUDable Methods to get the data
type WebHandler struct {
	EmptyStruct func() CObject
}

// CObject is the definition of our object, holds the structs
type CObject interface {
	web.CRUDable
	web.Rights
}

// HandleHTTPError does what it says
func HandleHTTPError(err error, ctx echo.Context) *echo.HTTPError {
	if a, has := err.(web.HTTPErrorProcessor); has {
		errDetails := a.HTTPError()
		return echo.NewHTTPError(errDetails.HTTPCode, errDetails)
	}
	config.LoggingProvider.Error(err.Error())
	return echo.NewHTTPError(http.StatusInternalServerError)
}
