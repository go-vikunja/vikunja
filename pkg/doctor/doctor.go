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

// Package doctor provides diagnostic checks for Vikunja installations.
package doctor

// Run executes all diagnostic checks in order and returns the results. Each group is
// passed to emit as soon as it completes, so a slow check does not withhold the
// groups before it.
func Run(emit func(CheckGroup)) []CheckGroup {
	var groups []CheckGroup
	collect := func(group CheckGroup) {
		groups = append(groups, group)
		emit(group)
	}

	collect(CheckSystem())
	collect(CheckConfig())
	collect(CheckDatabase())
	collect(CheckFiles())

	CheckOptionalServices(collect)

	return groups
}
