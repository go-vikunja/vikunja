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

package humaapi

import (
	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
)

// RegisterLabelRoutes wires Huma-flavoured Label CRUD operations onto the
// given Huma API. Runs alongside (not replacing) the legacy swag-driven
// routes for the duration of the spike.
func RegisterLabelRoutes(api huma.API) {
	Register(api, Config[*models.Label, SingleID]{
		Tag:      "labels",
		BasePath: "/labels",
		ItemPath: "/labels/{id}",
		New:      func() *models.Label { return &models.Label{} },
		ApplyPath: func(l *models.Label, p SingleID) {
			l.ID = p.ID
		},
	})
}
