//lint:file-ignore ST1018 The const below is not ours

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

package v1

import (
	_ "embed"
	"net/http"

	"code.vikunja.io/api/pkg/log"
	_ "code.vikunja.io/api/pkg/swagger" // To make sure the swag files are properly registered

	"github.com/labstack/echo/v5"
	"github.com/swaggo/swag"
)

// DocsJSON serves swagger doc json specs
func DocsJSON(c *echo.Context) error {
	doc, err := swag.ReadDoc()
	if err != nil {
		log.Error(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error").Wrap(err)
	}

	return c.Blob(http.StatusOK, echo.MIMEApplicationJSON, []byte(doc))
}

// RedocUI serves everything needed to provide the redoc ui
func RedocUI(c *echo.Context) error {
	return c.HTML(http.StatusOK, RedocUITemplate)
}

// RedocUITemplate contains the html + js needed for redoc ui
//
//go:embed templates/redoc.html
var RedocUITemplate string
