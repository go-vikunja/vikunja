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
	_ "code.vikunja.io/api/pkg/config" // To trigger its init() which initializes the config
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/notifications"
)

// SetupTests takes care of seting up the db, fixtures etc.
// This is an extra function to be able to call the fixtures setup from the web tests.
func SetupTests() {
	var err error
	x, err = db.CreateTestEngine()
	if err != nil {
		log.Fatal(err)
	}

	tables := []interface{}{}
	tables = append(tables, GetTables()...)
	tables = append(tables, notifications.GetTables()...)

	err = x.Sync2(tables...)
	if err != nil {
		log.Fatal(err)
	}

	err = db.CreateParadeDBIndexes()
	if err != nil {
		log.Fatal(err)
	}

	err = db.InitTestFixtures(
		"files",
		"label_tasks",
		"labels",
		"link_shares",
		"projects",
		"task_assignees",
		"task_attachments",
		"task_comments",
		"task_relations",
		"task_reminders",
		"tasks",
		"team_projects",
		"team_members",
		"teams",
		"users",
		"user_tokens",
		"users_projects",
		"buckets",
		"saved_filters",
		"subscriptions",
		"favorites",
		"api_tokens",
		"reactions",
		"project_views",
		"task_positions",
		"task_buckets",
	)
	if err != nil {
		log.Fatal(err)
	}

	// Start the pseudo mail queue
	mail.StartMailDaemon()
}
