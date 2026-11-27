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

package caldav

import (
	"net/http"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/models"
)

func taskETag(t *models.Task) string {
	return `"` + strconv.FormatInt(t.ID, 10) + `-` + strconv.FormatInt(t.Updated.Unix(), 10) + `"`
}

// The ctag and sync-token derive from this, so a purge must move it too.
func collectionETag(p *models.ProjectWithTasksAndBuckets, tasks []*models.TaskWithComments) string {
	latest := p.Updated
	for _, t := range tasks {
		if t.Updated.After(latest) {
			latest = t.Updated
		}
	}

	etag := strconv.FormatInt(p.ID, 10) + "-" + strconv.FormatInt(latest.Unix(), 10)
	if !p.LastPurgedAt.IsZero() {
		etag += ";purge=" + strconv.FormatInt(p.LastPurgedAt.Unix(), 10)
	}
	return `"` + etag + `"`
}

// http.TimeFormat always says GMT, whatever location the time is in.
func lastModified(t time.Time) string {
	return t.UTC().Format(http.TimeFormat)
}

func renderTask(t *models.Task) string {
	project := models.ProjectWithTasksAndBuckets{Tasks: []*models.TaskWithComments{{Task: *t}}}
	return caldav.GetCaldavTodosForTasks(&project, project.Tasks)
}
