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

package caldav

import (
	"code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"github.com/laurent22/ical-go"
	"strconv"
	"time"
)

func getCaldavTodosForTasks(list *models.List) string {

	// Make caldav todos from Vikunja todos
	var caldavtodos []*caldav.Todo
	for _, t := range list.Tasks {

		durationString := t.EndDateUnix - t.StartDateUnix
		duration, _ := time.ParseDuration(strconv.FormatInt(durationString, 10) + `s`)

		caldavtodos = append(caldavtodos, &caldav.Todo{
			TimestampUnix: t.Updated,
			UID:           t.UID,
			Summary:       t.Text,
			Description:   t.Description,
			CompletedUnix: t.DoneAtUnix,
			// Organizer:     &t.CreatedBy, // Disabled until we figure out how this works
			Priority:    t.Priority,
			StartUnix:   t.StartDateUnix,
			EndUnix:     t.EndDateUnix,
			CreatedUnix: t.Created,
			UpdatedUnix: t.Updated,
			DueDateUnix: t.DueDateUnix,
			Duration:    duration,
		})
	}

	caldavConfig := &caldav.Config{
		Name:   list.Title,
		ProdID: "Vikunja Todo App",
	}

	return caldav.ParseTodos(caldavConfig, caldavtodos)
}

func parseTaskFromVTODO(content string) (vTask *models.Task, err error) {
	parsed, err := ical.ParseCalendar(content)
	if err != nil {
		return nil, err
	}

	// We put the task details in a map to be able to handle them more easily
	task := make(map[string]string)
	for _, c := range parsed.Children {
		if c.Name == "VTODO" {
			for _, entry := range c.Children {
				task[entry.Name] = entry.Value
			}
			// Breaking, to only process the first task
			break
		}
	}

	// Parse the UID
	var priority int64
	if _, ok := task["PRIORITY"]; ok {
		priority, err = strconv.ParseInt(task["PRIORITY"], 10, 64)
		if err != nil {
			return nil, err
		}
	}

	// Parse the enddate
	duration, _ := time.ParseDuration(task["DURATION"])

	vTask = &models.Task{
		UID:           task["UID"],
		Text:          task["SUMMARY"],
		Description:   task["DESCRIPTION"],
		Priority:      priority,
		DueDateUnix:   caldavTimeToUnixTimestamp(task["DUE"]),
		Updated:       caldavTimeToUnixTimestamp(task["DTSTAMP"]),
		StartDateUnix: caldavTimeToUnixTimestamp(task["DTSTART"]),
		DoneAtUnix:    caldavTimeToUnixTimestamp(task["COMPLETED"]),
	}

	if task["STATUS"] == "COMPLETED" {
		vTask.Done = true
	}

	if duration > 0 && vTask.StartDateUnix > 0 {
		vTask.EndDateUnix = vTask.StartDateUnix + int64(duration.Seconds())
	}

	return
}

func caldavTimeToUnixTimestamp(tstring string) int64 {
	if tstring == "" {
		return 0
	}

	t, err := time.Parse(caldav.DateFormat, tstring)
	if err != nil {
		log.Warningf("Error while parsing caldav time %s to unix time: %s", tstring, err)
		return 0
	}
	return t.Unix()
}
