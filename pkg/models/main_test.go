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
	"fmt"
	"os"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
)

func setupTime() {
	var err error
	loc, err := time.LoadLocation("GMT")
	if err != nil {
		fmt.Printf("Error setting up time: %s", err)
		os.Exit(1)
	}
	testCreatedTime, err = time.ParseInLocation(time.RFC3339Nano, "2018-12-01T15:13:12.0+00:00", loc)
	if err != nil {
		fmt.Printf("Error setting up time: %s", err)
		os.Exit(1)
	}
	testCreatedTime = testCreatedTime.In(loc)
	testUpdatedTime, err = time.ParseInLocation(time.RFC3339Nano, "2018-12-02T15:13:12.0+00:00", loc)
	if err != nil {
		fmt.Printf("Error setting up time: %s", err)
		os.Exit(1)
	}
	testUpdatedTime = testUpdatedTime.In(loc)
}

func TestMain(m *testing.M) {

	setupTime()

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

	SetupTests()

	events.Fake()

	os.Exit(m.Run())
}
