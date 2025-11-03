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
	"net/http"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/web"

	"github.com/labstack/echo/v4"
)

// WebHandler defines the webhandler object
// This does web stuff, aka returns json etc. Uses CRUDable Methods to get the data
type WebHandler struct {
	EmptyStruct func() CObject
}

// CObject is the definition of our object, holds the structs
type CObject interface {
	web.CRUDable
	web.Permissions
}

// HandleHTTPError does what it says
func HandleHTTPError(err error) *echo.HTTPError {
	log.Error(err.Error())
	if a, has := err.(web.HTTPErrorProcessor); has {
		errDetails := a.HTTPError()
		return echo.NewHTTPError(errDetails.HTTPCode, errDetails).SetInternal(err)
	}
	return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
}
