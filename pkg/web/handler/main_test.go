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

package handler

import (
	"os"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/testutil"
	"code.vikunja.io/api/pkg/user"
)

func TestMain(m *testing.M) {
	// Setup
	config.InitDefaultConfig()
	// We need to set the root path even if we're not using the config, otherwise fixtures are not loaded correctly
	if os.Getenv("VIKUNJA_SERVICE_ROOTPATH") == "" {
		config.ServiceRootpath.Set("../../../") // Default for running from pkg/web/handler
	} else {
		config.ServiceRootpath.Set(os.Getenv("VIKUNJA_SERVICE_ROOTPATH"))
	}

	// Initialize logger for tests
	log.InitLogger()

	// Initialize service dependency injection explicitly
	testutil.Init()

	// Some tests use the file engine, so we'll need to initialize that
	files.InitTests()
	user.InitTests()
	models.SetupTests()
	events.Fake()
	keyvalue.InitStorage()

	err := db.LoadFixtures()
	if err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	// Teardown (if needed)
	// Currently no teardown needed

	os.Exit(code)
}
