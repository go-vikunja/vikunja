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

package feeds

import (
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

// CalendarContentType is the content type of the iCalendar task feeds.
const CalendarContentType = "text/calendar; charset=utf-8"

// BuildTasksCalendar renders the tasks of one project (or, with projectID 0,
// of every project u can read) as an iCalendar feed of VEVENTs. Done tasks are
// left out unless includeDone is set.
func BuildTasksCalendar(s *xorm.Session, u *user.User, projectID int64, includeDone bool) (string, error) {
	name := "Vikunja"
	if projectID != 0 {
		project := &models.Project{ID: projectID}
		canRead, _, err := project.CanRead(s, u)
		if err != nil {
			return "", err
		}
		if !canRead {
			return "", models.ErrUserDoesNotHaveAccessToProject{
				ProjectID: projectID,
				UserID:    u.ID,
			}
		}
		if err := project.ReadOne(s, u); err != nil {
			return "", err
		}
		name = project.Title
	}

	tc := &models.TaskCollection{ProjectID: projectID}
	if !includeDone {
		tc.Filter = "done = false"
	}
	result, _, _, err := tc.ReadAll(s, u, "", 0, -1)
	if err != nil {
		return "", err
	}
	tasks, ok := result.([]*models.Task)
	if !ok {
		return "", fmt.Errorf("task collection returned unexpected type %T", result)
	}

	loc := config.GetTimeZone()
	if u.Timezone != "" {
		if userLoc, err := time.LoadLocation(u.Timezone); err == nil {
			loc = userLoc
		}
	}

	events := caldav.GetEventsForTasks(tasks, loc, config.ServicePublicURL.GetString())
	return caldav.ParseEvents(&caldav.Config{
		Name:   name,
		ProdID: "Vikunja Todo App",
	}, events), nil
}
