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

package cron

import (
	"github.com/robfig/cron/v3"
)

var c *cron.Cron

// Init starts the cron
func Init() {
	c = cron.New()
	c.Start()
}

// Schedule schedules a job as a cron job
func Schedule(schedule string, f func()) (err error) {
	_, err = c.AddFunc(schedule, f)
	return
}

// Stop stops the cron scheduler
func Stop() {
	c.Stop()
}
