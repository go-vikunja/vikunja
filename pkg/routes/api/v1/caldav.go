//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes/crud"
	"github.com/labstack/echo"
	"net/http"
	"time"
)

// Caldav returns a caldav-readable format with all tasks
// @Summary CalDAV-readable format with all tasks as calendar events.
// @Description Returns a calDAV-parsable format with all tasks as calendar events. Only returns tasks with a due date. Also creates reminders when the task has one.
// @tags task
// @Produce text/plain
// @Security BasicAuth
// @Success 200 {string} string "The caldav events."
// @Failure 403 {string} string "Unauthorized."
// @Router /tasks/caldav [get]
func Caldav(c echo.Context) error {

	// Request basic auth
	user, pass, ok := c.Request().BasicAuth()

	// Check credentials
	creds := &models.UserLogin{
		Username: user,
		Password: pass,
	}
	u, err := models.CheckUserCredentials(creds)

	if !ok || err != nil {
		c.Response().Header().Set("WWW-Authenticate", `Basic realm="Vikunja cal"`)
		return c.String(http.StatusUnauthorized, "Unauthorized.")
	}

	// Get all tasks for that user
	tasks, err := models.GetTasksByUser("", &u, -1)
	if err != nil {
		return crud.HandleHTTPError(err)
	}

	hour := int64(time.Hour.Seconds())
	var caldavTasks []*caldav.Event
	for _, t := range tasks {
		if t.DueDateUnix != 0 {
			event := &caldav.Event{
				Summary:       t.Text,
				Description:   t.Description,
				UID:           "",
				TimestampUnix: t.Updated,
				StartUnix:     t.DueDateUnix,
				EndUnix:       t.DueDateUnix + hour,
			}

			if len(t.RemindersUnix) > 0 {
				for _, rem := range t.RemindersUnix {
					event.Alarms = append(event.Alarms, caldav.Alarm{TimeUnix: rem})
				}
			}

			caldavTasks = append(caldavTasks, event)
		}
	}

	caldavConfig := &caldav.Config{
		Name:   "Vikunja Calendar for " + u.Username,
		ProdID: "Vikunja Todo App",
	}

	return c.String(http.StatusOK, caldav.ParseEvents(caldavConfig, caldavTasks))
}
