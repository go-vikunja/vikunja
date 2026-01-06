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
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
)

// RegisterWikiPageRoutes registers all routes for wiki pages
func RegisterWikiPageRoutes(a *echo.Group) {
	wikiHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.WikiPage{}
		},
	}

	// Create wiki page
	a.PUT("/projects/:project/wiki", wikiHandler.CreateWeb)
	// Get all wiki pages in a project
	a.GET("/projects/:project/wiki", wikiHandler.ReadAllWeb)
	// Get a single wiki page
	a.GET("/projects/:project/wiki/:page", wikiHandler.ReadOneWeb)
	// Update a wiki page
	a.POST("/projects/:project/wiki/:page", wikiHandler.UpdateWeb)
	// Delete a wiki page
	a.DELETE("/projects/:project/wiki/:page", wikiHandler.DeleteWeb)
}
