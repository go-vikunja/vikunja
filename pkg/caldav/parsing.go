// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package caldav

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/utils"

	ics "github.com/arran4/golang-ical"
)

func GetCaldavTodosForTasks(project *models.ProjectWithTasksAndBuckets, projectTasks []*models.TaskWithComments) string {

	// Make caldav todos from Vikunja todos
	var caldavtodos []*Todo
	for _, t := range projectTasks {

		duration := t.EndDate.Sub(t.StartDate)
		var categories []string
		for _, label := range t.Labels {
			categories = append(categories, label.Title)
		}
		var alarms []Alarm
		for _, reminder := range t.Reminders {
			alarms = append(alarms, Alarm{
				Time:       reminder.Reminder,
				Duration:   time.Duration(reminder.RelativePeriod) * time.Second,
				RelativeTo: reminder.RelativeTo,
			})
		}

		var relations []Relation
		for reltype, tasks := range t.RelatedTasks {
			for _, r := range tasks {
				relations = append(relations, Relation{
					Type: reltype,
					UID:  r.UID,
				})
			}
		}

		caldavtodos = append(caldavtodos, &Todo{
			Timestamp:   t.Updated,
			UID:         t.UID,
			Summary:     t.Title,
			Description: t.Description,
			Completed:   t.DoneAt,
			// Organizer:     &t.CreatedBy, // Disabled until we figure out how this works
			Categories:  categories,
			Priority:    t.Priority,
			Start:       t.StartDate,
			End:         t.EndDate,
			Created:     t.Created,
			Updated:     t.Updated,
			DueDate:     t.DueDate,
			Duration:    duration,
			RepeatAfter: t.RepeatAfter,
			RepeatMode:  t.RepeatMode,
			Alarms:      alarms,
			Relations:   relations,
		})
	}

	caldavConfig := &Config{
		Name:   project.Title,
		ProdID: "Vikunja Todo App",
	}

	return ParseTodos(caldavConfig, caldavtodos)
}

func ParseTaskFromVTODO(content string) (vTask *models.Task, err error) {
	parsed, err := ics.ParseCalendar(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	vTodo, ok := parsed.Components[0].(*ics.VTodo)
	if !ok {
		return nil, errors.New("VTODO element not found")
	}
	// We put the vTodo details in a map to be able to handle them more easily
	task := make(map[string]ics.IANAProperty)
	var relations []ics.IANAProperty
	for _, c := range vTodo.UnknownPropertiesIANAProperties() {
		task[c.IANAToken] = c
		if strings.HasPrefix(c.IANAToken, "RELATED-TO") {
			relations = append(relations, c)
		}
	}

	// Parse the priority
	var priority int64
	if _, ok := task["PRIORITY"]; ok {
		priorityParsed, err := strconv.ParseInt(task["PRIORITY"].Value, 10, 64)
		if err != nil {
			return nil, err
		}

		priority = parseVTODOPriority(priorityParsed)
	}

	// Parse the enddate
	duration, _ := time.ParseDuration(task["DURATION"].Value)

	description := strings.ReplaceAll(task["DESCRIPTION"].Value, "\\,", ",")
	description = strings.ReplaceAll(description, "\\n", "\n")

	var labels []*models.Label
	if val, ok := task["CATEGORIES"]; ok {
		categories := strings.Split(val.Value, ",")
		labels = make([]*models.Label, 0, len(categories))
		for _, category := range categories {
			labels = append(labels, &models.Label{
				Title: category,
			})
		}
	}

	vTask = &models.Task{
		UID:         task["UID"].Value,
		Title:       task["SUMMARY"].Value,
		Description: description,
		Priority:    priority,
		Labels:      labels,
		DueDate:     caldavTimeToTimestamp(task["DUE"]),
		Updated:     caldavTimeToTimestamp(task["DTSTAMP"]),
		StartDate:   caldavTimeToTimestamp(task["DTSTART"]),
		DoneAt:      caldavTimeToTimestamp(task["COMPLETED"]),
	}

	for _, c := range relations {
		var relTypeStr string
		if _, ok := c.ICalParameters["RELTYPE"]; ok {
			if len(c.ICalParameters["RELTYPE"]) != 1 {
				continue
			}

			relTypeStr = c.ICalParameters["RELTYPE"][0]
		}

		var relationKind models.RelationKind
		switch relTypeStr {
		case "PARENT":
			relationKind = models.RelationKindParenttask
		case "CHILD":
			relationKind = models.RelationKindSubtask
		default:
			relationKind = models.RelationKindParenttask
		}

		if vTask.RelatedTasks == nil {
			vTask.RelatedTasks = make(map[models.RelationKind][]*models.Task)
		}

		vTask.RelatedTasks[relationKind] = append(vTask.RelatedTasks[relationKind], &models.Task{
			UID: c.Value,
		})
	}

	if task["STATUS"].Value == "COMPLETED" {
		vTask.Done = true
	}

	if duration > 0 && !vTask.StartDate.IsZero() {
		vTask.EndDate = vTask.StartDate.Add(duration)
	}

	for _, vAlarm := range vTodo.SubComponents() {
		if vAlarm, ok := vAlarm.(*ics.VAlarm); ok {
			vTask = parseVAlarm(vAlarm, vTask)
		}
	}

	return
}

func parseVAlarm(vAlarm *ics.VAlarm, vTask *models.Task) *models.Task {
	for _, property := range vAlarm.UnknownPropertiesIANAProperties() {
		if property.IANAToken != "TRIGGER" {
			continue
		}

		if contains(property.ICalParameters["VALUE"], "DATE-TIME") {
			// Example: TRIGGER;VALUE=DATE-TIME:20181201T011210Z
			vTask.Reminders = append(vTask.Reminders, &models.TaskReminder{
				Reminder: caldavTimeToTimestamp(property),
			})
			continue
		}

		duration := utils.ParseISO8601Duration(property.Value)

		if contains(property.ICalParameters["RELATED"], "END") {
			// Example: TRIGGER;RELATED=END:-P2D
			if vTask.EndDate.IsZero() {
				vTask.Reminders = append(vTask.Reminders, &models.TaskReminder{
					RelativePeriod: int64(duration.Seconds()),
					RelativeTo:     models.ReminderRelationDueDate})
			} else {
				vTask.Reminders = append(vTask.Reminders, &models.TaskReminder{
					RelativePeriod: int64(duration.Seconds()),
					RelativeTo:     models.ReminderRelationEndDate})
			}
			continue
		}

		// Example: TRIGGER;RELATED=START:-P2D
		// Example: TRIGGER:-PT60M
		vTask.Reminders = append(vTask.Reminders, &models.TaskReminder{
			RelativePeriod: int64(duration.Seconds()),
			RelativeTo:     models.ReminderRelationStartDate})
	}
	return vTask
}

func contains(array []string, str string) bool {
	for _, value := range array {
		if value == str {
			return true
		}
	}
	return false
}

// https://tools.ietf.org/html/rfc5545#section-3.3.5
func caldavTimeToTimestamp(ianaProperty ics.IANAProperty) time.Time {
	tstring := ianaProperty.Value
	if tstring == "" {
		return time.Time{}
	}

	format := DateFormat

	if strings.HasSuffix(tstring, "Z") {
		format = `20060102T150405Z`
	}

	if len(tstring) == 8 {
		format = `20060102`
	}

	var t time.Time
	var err error
	tzParameter := ianaProperty.ICalParameters["TZID"]
	if len(tzParameter) > 0 {
		loc, err := time.LoadLocation(tzParameter[0])
		if err != nil {
			log.Warningf("Error while parsing caldav timezone %s: %s", tzParameter[0], err)
		} else {
			t, err = time.ParseInLocation(format, tstring, loc)
			if err != nil {
				log.Warningf("Error while parsing caldav time %s to TimeStamp: %s at location %s", tstring, loc, err)
			} else {
				t = t.In(config.GetTimeZone())
				return t
			}
		}
	}
	t, err = time.Parse(format, tstring)
	if err != nil {
		log.Warningf("Error while parsing caldav time %s to TimeStamp: %s", tstring, err)
		return time.Time{}
	}
	return t
}
