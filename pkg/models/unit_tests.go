// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	_ "code.vikunja.io/api/pkg/config" // To trigger its init() which initializes the config
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/mail"
)

// SetupTests takes care of seting up the db, fixtures etc.
// This is an extra function to be able to call the fixtures setup from the integration tests.
func SetupTests() {
	var err error
	x, err = db.CreateTestEngine()
	if err != nil {
		log.Fatal(err)
	}

	err = x.Sync2(GetTables()...)
	if err != nil {
		log.Fatal(err)
	}

	err = db.InitTestFixtures(
		"files",
		"label_task",
		"labels",
		"link_sharing",
		"list",
		"namespaces",
		"task_assignees",
		"task_attachments",
		"task_comments",
		"task_relations",
		"task_reminders",
		"tasks",
		"team_list",
		"team_members",
		"team_namespaces",
		"teams",
		"users",
		"users_list",
		"users_namespace")
	if err != nil {
		log.Fatal(err)
	}

	// Start the pseudo mail queue
	mail.StartMailDaemon()
}
