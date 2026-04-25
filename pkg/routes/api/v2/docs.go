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

package apiv2

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v5"
)

//go:embed scalar/scalar.html
var scalarHTML string

//go:embed scalar/scalar.standalone.js
var scalarJS []byte

// ScalarUI renders the Scalar API reference shell HTML. Assets are
// fetched from /api/v2/docs/scalar.standalone.js — fully local, no CDN.
func ScalarUI(c *echo.Context) error {
	return c.HTML(http.StatusOK, scalarHTML)
}

// ScalarJS serves the embedded Scalar standalone JS bundle.
func ScalarJS(c *echo.Context) error {
	return c.Blob(http.StatusOK, echo.MIMEApplicationJavaScript, scalarJS)
}
