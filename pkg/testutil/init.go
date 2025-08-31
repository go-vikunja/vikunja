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

// Package testutil provides explicit initialization for test environments.
// This package ensures that service dependency injection is set up in a
// deterministic order, replacing the fragile init() function pattern.
package testutil

import "code.vikunja.io/api/pkg/services"

// Init initializes all service dependency injection in a deterministic order.
// This function replaces the fragile init() function pattern and must be called
// explicitly in TestMain functions that need dependency injection to work.
//
// The initialization order is carefully chosen to respect dependencies:
// 1. UserService (foundational, used by others)
// 2. TaskService (depends on user service)
// 3. ProjectService (depends on user service)
// 4. KanbanService (depends on user and task services)
func Init() {
	services.InitUserService()
	services.InitTaskService()
	services.InitProjectService()
	services.InitKanbanService()
}