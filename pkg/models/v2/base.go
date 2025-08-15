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

package v2

import "github.com/labstack/echo/v4"

type Links struct {
	Self *Link `json:"self"`
}

type Link struct {
	Href string `json:"href"`
}

// GetPageAndPerPage gets the page and per_page query parameters from the context.
func GetPageAndPerPage(c echo.Context) (page int, perPage int) {
	page, _ = c.QueryParamInt("page", 1)
	perPage, _ = c.QueryParamInt("per_page", 20)
	return
}
