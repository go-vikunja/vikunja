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

import "github.com/danielgtaylor/huma/v2"

var routeRegistrars []func(huma.API)

// AddRouteRegistrar records a resource's route-registration function. Each
// resource file calls this from an init() so new resources never touch the
// central wiring.
func AddRouteRegistrar(f func(huma.API)) {
	routeRegistrars = append(routeRegistrars, f)
}

// RegisterAll runs every registrar collected via AddRouteRegistrar, then
// enables AutoPatch. Registrars run in init() order (filename order across the
// package); the order they register routes in is irrelevant. AutoPatch runs
// last so it can synthesise PATCH counterparts for all GET + PUT pairs.
func RegisterAll(api huma.API) {
	for _, r := range routeRegistrars {
		r(api)
	}
	EnableAutoPatch(api)
}
