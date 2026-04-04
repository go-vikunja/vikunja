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
	"bytes"
	_ "embed"
	"html/template"
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	_ "code.vikunja.io/api/pkg/swagger" // To make sure the swag files are properly registered

	"github.com/labstack/echo/v5"
	"github.com/swaggo/swag"
)

//go:embed redoc/redoc.html
var redocHTML string

var redocUITemplate = template.Must(template.New("redoc").Parse(redocHTML))

//go:embed redoc/redoc.standalone.js
var redocJS []byte

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
	publicURL := config.ServicePublicURL.GetString()
	docsURL := strings.TrimRight(publicURL, "/") + "/api/v1/docs.json"

	var buf bytes.Buffer
	data := map[string]string{"Url": docsURL}

	if err := redocUITemplate.Execute(&buf, data); err != nil {
		return err
	}

	return c.HTML(http.StatusOK, buf.String())
}

// RedocJS serves the embedded redoc standalone JavaScript bundle
func RedocJS(c *echo.Context) error {
	return c.Blob(http.StatusOK, echo.MIMEApplicationJavaScript, redocJS)
}
