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

var cssColorsToHex map[string]string

func init() {
	cssColorsToHex = map[string]string{
		"aliceblue":            "f0f8ff",
		"antiquewhite":         "faebd7",
		"aqua":                 "00ffff",
		"aquamarine":           "7fffd4",
		"azure":                "f0ffff",
		"beige":                "f5f5dc",
		"bisque":               "ffe4c4",
		"black":                "000000",
		"blanchedalmond":       "ffebcd",
		"blue":                 "0000ff",
		"blueviolet":           "8a2be2",
		"brown":                "a52a2a",
		"burlywood":            "deb887",
		"cadetblue":            "5f9ea0",
		"chartreuse":           "7fff00",
		"chocolate":            "d2691e",
		"coral":                "ff7f50",
		"cornflowerblue":       "6495ed",
		"cornsilk":             "fff8dc",
		"crimson":              "dc143c",
		"cyan":                 "00ffff",
		"darkblue":             "00008b",
		"darkcyan":             "008b8b",
		"darkgoldenrod":        "b8860b",
		"darkgray":             "a9a9a9",
		"darkgreen":            "006400",
		"darkgrey":             "a9a9a9",
		"darkkhaki":            "bdb76b",
		"darkmagenta":          "8b008b",
		"darkolivegreen":       "556b2f",
		"darkorange":           "ff8c00",
		"darkorchid":           "9932cc",
		"darkred":              "8b0000",
		"darksalmon":           "e9967a",
		"darkseagreen":         "8fbc8f",
		"darkslateblue":        "483d8b",
		"darkslategray":        "2f4f4f",
		"darkslategrey":        "2f4f4f",
		"darkturquoise":        "00ced1",
		"darkviolet":           "9400d3",
		"deeppink":             "ff1493",
		"deepskyblue":          "00bfff",
		"dimgray":              "696969",
		"dimgrey":              "696969",
		"dodgerblue":           "1e90ff",
		"firebrick":            "b22222",
		"floralwhite":          "fffaf0",
		"forestgreen":          "228b22",
		"fuchsia":              "ff00ff",
		"gainsboro":            "dcdcdc",
		"ghostwhite":           "f8f8ff",
		"gold":                 "ffd700",
		"goldenrod":            "daa520",
		"gray":                 "808080",
		"green":                "008000",
		"greenyellow":          "adff2f",
		"grey":                 "808080",
		"honeydew":             "f0fff0",
		"hotpink":              "ff69b4",
		"indianred":            "cd5c5c",
		"indigo":               "4b0082",
		"ivory":                "fffff0",
		"khaki":                "f0e68c",
		"lavender":             "e6e6fa",
		"lavenderblush":        "fff0f5",
		"lawngreen":            "7cfc00",
		"lemonchiffon":         "fffacd",
		"lightblue":            "add8e6",
		"lightcoral":           "f08080",
		"lightcyan":            "e0ffff",
		"lightgoldenrodyellow": "fafad2",
		"lightgray":            "d3d3d3",
		"lightgreen":           "90ee90",
		"lightgrey":            "d3d3d3",
		"lightpink":            "ffb6c1",
		"lightsalmon":          "ffa07a",
		"lightseagreen":        "20b2aa",
		"lightskyblue":         "87cefa",
		"lightslategray":       "778899",
		"lightslategrey":       "778899",
		"lightsteelblue":       "b0c4de",
		"lightyellow":          "ffffe0",
		"lime":                 "00ff00",
		"limegreen":            "32cd32",
		"linen":                "faf0e6",
		"magenta":              "ff00ff",
		"maroon":               "800000",
		"mediumaquamarine":     "66cdaa",
		"mediumblue":           "0000cd",
		"mediumorchid":         "ba55d3",
		"mediumpurple":         "9370db",
		"mediumseagreen":       "3cb371",
		"mediumslateblue":      "7b68ee",
		"mediumspringgreen":    "00fa9a",
		"mediumturquoise":      "48d1cc",
		"mediumvioletred":      "c71585",
		"midnightblue":         "191970",
		"mintcream":            "f5fffa",
		"mistyrose":            "ffe4e1",
		"moccasin":             "ffe4b5",
		"navajowhite":          "ffdead",
		"navy":                 "000080",
		"oldlace":              "fdf5e6",
		"olive":                "808000",
		"olivedrab":            "6b8e23",
		"orange":               "ffa500",
		"orangered":            "ff4500",
		"orchid":               "da70d6",
		"palegoldenrod":        "eee8aa",
		"palegreen":            "98fb98",
		"paleturquoise":        "afeeee",
		"palevioletred":        "db7093",
		"papayawhip":           "ffefd5",
		"peachpuff":            "ffdab9",
		"peru":                 "cd853f",
		"pink":                 "ffc0cb",
		"plum":                 "dda0dd",
		"powderblue":           "b0e0e6",
		"purple":               "800080",
		"red":                  "ff0000",
		"rosybrown":            "bc8f8f",
		"royalblue":            "4169e1",
		"saddlebrown":          "8b4513",
		"salmon":               "fa8072",
		"sandybrown":           "f4a460",
		"seagreen":             "2e8b57",
		"seashell":             "fff5ee",
		"sienna":               "a0522d",
		"silver":               "c0c0c0",
		"skyblue":              "87ceeb",
		"slateblue":            "6a5acd",
		"slategray":            "708090",
		"slategrey":            "708090",
		"snow":                 "fffafa",
		"springgreen":          "00ff7f",
		"steelblue":            "4682b4",
		"tan":                  "d2b48c",
		"teal":                 "008080",
		"thistle":              "d8bfd8",
		"tomato":               "ff6347",
		"turquoise":            "40e0d0",
		"violet":               "ee82ee",
		"wheat":                "f5deb3",
		"white":                "ffffff",
		"whitesmoke":           "f5f5f5",
		"yellow":               "ffff00",
		"yellowgreen":          "9acd32",
	}
}

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

func getHexColorFromCaldavColor(caldavColor string) string {
	if caldavColor == "" {
		return ""
	}

	if caldavColor[:1] == "#" {
		caldavColor = strings.TrimPrefix(caldavColor, "#")
		if len(caldavColor) > 6 {
			caldavColor = caldavColor[:6]
		}
		return caldavColor
	}

	hexColor, has := cssColorsToHex[caldavColor]
	if !has {
		return ""
	}

	return hexColor
}

func ParseTaskFromVTODO(content string) (vTask *models.Task, err error) {
	parsed, err := ics.ParseCalendar(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	if len(parsed.Components) == 0 {
		return nil, errors.New("VTODO element does seem not contain any components")
	}
	vTodo, ok := parsed.Components[0].(*ics.VTodo)
	if !ok {
		return nil, errors.New("VTODO element not found")
	}
	// We put the vTodo details in a map to be able to handle them more easily
	task := make(map[string]ics.IANAProperty)

	var relations []ics.IANAProperty
	var color string
	for _, c := range vTodo.UnknownPropertiesIANAProperties() {
		task[c.IANAToken] = c
		if strings.HasPrefix(c.IANAToken, "RELATED-TO") {
			relations = append(relations, c)
		}
		if c.IANAToken == "X-APPLE-CALENDAR-COLOR" {
			color = c.Value
		}
		if c.IANAToken == "X-OUTLOOK-COLOR" {
			color = c.Value
		}
		if c.IANAToken == "X-FUNAMBOL-COLOR" {
			color = c.Value
		}
		if c.IANAToken == "COLOR" {
			color = c.Value
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
		HexColor:    getHexColorFromCaldavColor(color),
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
