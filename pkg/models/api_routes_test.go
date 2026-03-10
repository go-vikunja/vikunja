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

package models

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanDoAPIRoute_BulkLabelTask(t *testing.T) {
	// Reset apiTokenRoutes to isolate this test
	apiTokenRoutes = make(map[string]APITokenRoute)

	// Register the standard CRUD routes for tasks_labels first
	CollectRoutesForAPITokenUsage(echo.RouteInfo{
		Method: "PUT",
		Path:   "/api/v1/tasks/:projecttask/labels",
	}, true)
	CollectRoutesForAPITokenUsage(echo.RouteInfo{
		Method: "DELETE",
		Path:   "/api/v1/tasks/:projecttask/labels/:label",
	}, true)

	// Now register the bulk route
	CollectRoutesForAPITokenUsage(echo.RouteInfo{
		Method: "POST",
		Path:   "/api/v1/tasks/:projecttask/labels/bulk",
	}, true)

	// Verify that the tasks_labels route group exists
	routes, has := apiTokenRoutes["tasks_labels"]
	require.True(t, has, "tasks_labels route group should exist")

	// The bulk route should be registered as "update_bulk" under tasks_labels
	bulkRoute, has := routes["update_bulk"]
	require.True(t, has, "update_bulk should exist in tasks_labels routes")
	assert.Equal(t, "/api/v1/tasks/:projecttask/labels/bulk", bulkRoute.Path)
	assert.Equal(t, "POST", bulkRoute.Method)
}
