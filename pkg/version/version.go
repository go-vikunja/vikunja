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

package version

import "code.vikunja.io/api/pkg/swagger"

// This package holds the version info
// It is an own package to avoid import cycles

// Version sets the version to be printed to the user. Gets overwritten by "make release" or "make build" with last git commit or tag.
var Version = "dev"

func init() {
	// Additional swagger information
	swagger.SwaggerInfo.Version = Version
}
