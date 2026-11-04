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

package routes

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

// encoding/xml ignores json tags and would populate json:"-" fields; form and multipart only bind explicit form tags.
type bodyBinder struct {
	echo.DefaultBinder
}

func (b *bodyBinder) Bind(c *echo.Context, target any) error {
	req := c.Request()
	if req.ContentLength != 0 {
		// Parsed like echo.BindBody does, so both agree on the media type.
		base, _, _ := strings.Cut(req.Header.Get(echo.HeaderContentType), ";")
		switch strings.TrimSpace(base) {
		case echo.MIMEApplicationJSON, echo.MIMEApplicationForm, echo.MIMEMultipartForm:
		default:
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, "Only JSON request bodies are accepted.")
		}
	}
	return b.DefaultBinder.Bind(c, target)
}
