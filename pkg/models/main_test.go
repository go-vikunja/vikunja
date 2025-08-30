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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
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

	// Set up a mock for the GetUsersOrLinkSharesFromIDsFunc for model tests,
	// as they should not depend on the services package.
	GetUsersOrLinkSharesFromIDsFunc = func(s *xorm.Session, ids []int64) (map[int64]*user.User, error) {
		usersMap := make(map[int64]*user.User)
		var userIDs []int64
		var linkShareIDs []int64
		for _, id := range ids {
			if id < 0 {
				linkShareIDs = append(linkShareIDs, id*-1)
				continue
			}
			userIDs = append(userIDs, id)
		}

		if len(userIDs) > 0 {
			var err error
			usersMap, err = user.GetUsersByIDs(s, userIDs)
			if err != nil {
				return nil, err
			}
		}

		if len(linkShareIDs) == 0 {
			return usersMap, nil
		}

		shares, err := GetLinkSharesByIDs(s, linkShareIDs)
		if err != nil {
			return nil, err
		}

		for _, share := range shares {
			usersMap[share.ID*-1] = share.ToUser()
		}

		return usersMap, nil
	}

	// Set up a mock for AddMoreInfoToTasksFunc for model tests,
	// as they should not depend on the services package.
	AddMoreInfoToTasksFunc = func(s *xorm.Session, taskMap map[int64]*Task, a web.Auth, view *ProjectView, expand []TaskCollectionExpandable) error {
		// This is a minimal mock that just returns nil - no additional task details are added in tests
		// Individual tests can override this if they need specific behavior
		return nil
	}

	events.Fake()

	os.Exit(m.Run())
}
