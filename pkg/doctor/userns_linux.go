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

//go:build linux

package doctor

import (
	"code.vikunja.io/api/pkg/utils"
)

func checkUserNamespace() CheckResult {
	if !utils.IsUserNamespaceActive() {
		return CheckResult{
			Name:   "User namespace",
			Passed: true,
			Value:  "not active",
		}
	}

	return CheckResult{
		Name:   "User namespace",
		Passed: true,
		Value:  "active (" + utils.UIDMappingSummary() + ")",
		Lines:  []string{"UIDs inside this container are remapped. See directory ownership check for details."},
	}
}
