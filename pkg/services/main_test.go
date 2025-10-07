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

package services

import (
	"os"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"
)

var testEngine *xorm.Engine

func TestMain(m *testing.M) {
	// Initialize logger for tests
	log.InitLogger()

	// Set default config
	config.InitDefaultConfig()
	// We need to set the root path even if we're not using the config, otherwise fixtures are not loaded correctly
	config.ServiceRootpath.Set(os.Getenv("VIKUNJA_SERVICE_ROOTPATH"))

	i18n.Init()

	// Some tests use the file engine, so we'll need to initialize that
	files.InitTests()

	user.InitTests()

	models.SetupTests()
	events.Fake()

	// Initialize service dependency injection in the correct order
	// First, wire up model/service dependencies
	InitializeDependencies()
	// Then initialize service instances
	InitUserService()
	InitSavedFilterService()
	InitTaskService()
	InitProjectService()
	InitKanbanService()
	InitProjectDuplicateService()
	InitAttachmentService()

	// Initialize testEngine for service tests
	testEngine = db.GetEngine()

	// Register API routes for permission validation in tests
	// This populates the apiTokenRoutes map needed for PermissionsAreValid()
	registerTestAPIRoutes()

	os.Exit(m.Run())
}

// registerTestAPIRoutes manually registers API routes for testing
// This is needed because tests don't start the web server, but API token
// permission validation requires the apiTokenRoutes map to be populated
func registerTestAPIRoutes() {
	// Register v1 task routes
	models.CollectRoute("PUT", "/api/v1/projects/:project/tasks", "create")
	models.CollectRoute("GET", "/api/v1/tasks/:taskid", "read_one")
	models.CollectRoute("POST", "/api/v1/tasks/:taskid", "update")
	models.CollectRoute("DELETE", "/api/v1/tasks/:taskid", "delete")

	// Register other commonly used routes for API token tests
	// Add more as needed for comprehensive test coverage
	models.CollectRoute("GET", "/api/v1/projects", "read_all")
	models.CollectRoute("PUT", "/api/v1/projects", "create")
	models.CollectRoute("GET", "/api/v1/projects/:project", "read_one")
	models.CollectRoute("POST", "/api/v1/projects/:project", "update")
	models.CollectRoute("DELETE", "/api/v1/projects/:project", "delete")
}
