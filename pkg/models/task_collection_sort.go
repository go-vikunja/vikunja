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
	"code.vikunja.io/api/pkg/timeutil"
	"fmt"
	"reflect"
	"sort"
)

type (
	sortParam struct {
		sortBy  sortProperty
		orderBy sortOrder // asc or desc
	}

	sortProperty string

	sortOrder string
)

const (
	taskPropertyID            string = "id"
	taskPropertyText          string = "text"
	taskPropertyDescription   string = "description"
	taskPropertyDone          string = "done"
	taskPropertyDoneAtUnix    string = "done_at_unix"
	taskPropertyDueDateUnix   string = "due_date_unix"
	taskPropertyCreatedByID   string = "created_by_id"
	taskPropertyListID        string = "list_id"
	taskPropertyRepeatAfter   string = "repeat_after"
	taskPropertyPriority      string = "priority"
	taskPropertyStartDateUnix string = "start_date_unix"
	taskPropertyEndDateUnix   string = "end_date_unix"
	taskPropertyHexColor      string = "hex_color"
	taskPropertyPercentDone   string = "percent_done"
	taskPropertyUID           string = "uid"
	taskPropertyCreated       string = "created"
	taskPropertyUpdated       string = "updated"
)

func (p sortProperty) String() string {
	return string(p)
}

const (
	orderInvalid    sortOrder = "invalid"
	orderAscending  sortOrder = "asc"
	orderDescending sortOrder = "desc"
)

func (o sortOrder) String() string {
	return string(o)
}

func getSortOrderFromString(s string) sortOrder {
	if s == "asc" {
		return orderAscending
	}
	if s == "desc" {
		return orderDescending
	}
	return orderInvalid
}

func (sp *sortParam) validate() error {
	if sp.orderBy != orderDescending && sp.orderBy != orderAscending {
		return ErrInvalidSortOrder{OrderBy: sp.orderBy}
	}
	return validateTaskField(string(sp.sortBy))
}

type taskComparator func(lhs, rhs *Task) int64

func mustMakeComparator(fieldName string) taskComparator {
	field, ok := reflect.TypeOf(&Task{}).Elem().FieldByName(fieldName)
	if !ok {
		panic(fmt.Sprintf("Field '%s' has not been found on Task", fieldName))
	}

	extractProp := func(task *Task) interface{} {
		return reflect.ValueOf(task).Elem().FieldByIndex(field.Index).Interface()
	}

	// Special case for handling TimeStamp types
	if field.Type.Name() == "TimeStamp" {
		return func(lhs, rhs *Task) int64 {
			return int64(extractProp(lhs).(timeutil.TimeStamp)) - int64(extractProp(rhs).(timeutil.TimeStamp))
		}
	}

	switch field.Type.Kind() {
	case reflect.Int64:
		return func(lhs, rhs *Task) int64 {
			return extractProp(lhs).(int64) - extractProp(rhs).(int64)
		}
	case reflect.Float64:
		return func(lhs, rhs *Task) int64 {
			floatLHS, floatRHS := extractProp(lhs).(float64), extractProp(rhs).(float64)
			if floatLHS > floatRHS {
				return 1
			} else if floatLHS < floatRHS {
				return -1
			}
			return 0
		}
	case reflect.String:
		return func(lhs, rhs *Task) int64 {
			strLHS, strRHS := extractProp(lhs).(string), extractProp(rhs).(string)
			if strLHS > strRHS {
				return 1
			} else if strLHS < strRHS {
				return -1
			}
			return 0
		}
	case reflect.Bool:
		return func(lhs, rhs *Task) int64 {
			boolLHS, boolRHS := extractProp(lhs).(bool), extractProp(rhs).(bool)
			if !boolLHS && boolRHS {
				return -1
			} else if boolLHS && !boolRHS {
				return 1
			}
			return 0
		}
	default:
		panic(fmt.Sprintf("Unsupported type for sorting: %s", field.Type.Name()))
	}
}

// This is a map of properties that can be sorted by
// and their appropriate comparator function.
// The comparator function sorts in ascending mode.
var propertyComparators = map[string]taskComparator{
	taskPropertyID:            mustMakeComparator("ID"),
	taskPropertyText:          mustMakeComparator("Text"),
	taskPropertyDescription:   mustMakeComparator("Description"),
	taskPropertyDone:          mustMakeComparator("Done"),
	taskPropertyDoneAtUnix:    mustMakeComparator("DoneAt"),
	taskPropertyDueDateUnix:   mustMakeComparator("DueDate"),
	taskPropertyCreatedByID:   mustMakeComparator("CreatedByID"),
	taskPropertyListID:        mustMakeComparator("ListID"),
	taskPropertyRepeatAfter:   mustMakeComparator("RepeatAfter"),
	taskPropertyPriority:      mustMakeComparator("Priority"),
	taskPropertyStartDateUnix: mustMakeComparator("StartDate"),
	taskPropertyEndDateUnix:   mustMakeComparator("EndDate"),
	taskPropertyHexColor:      mustMakeComparator("HexColor"),
	taskPropertyPercentDone:   mustMakeComparator("PercentDone"),
	taskPropertyUID:           mustMakeComparator("UID"),
	taskPropertyCreated:       mustMakeComparator("Created"),
	taskPropertyUpdated:       mustMakeComparator("Updated"),
}

// Creates a taskComparator that sorts by the first comparator and falls back to
// the second one (and so on...) if the properties were equal.
func combineComparators(comparators ...taskComparator) taskComparator {
	return func(lhs, rhs *Task) int64 {
		for _, compare := range comparators {
			res := compare(lhs, rhs)
			if res != 0 {
				return res
			}
		}
		return 0
	}
}

func sortTasks(tasks []*Task, by []*sortParam) {

	// Always sort at least by id asc so we have a consistent order of items every time
	// If we would not do this, we would get a different order for items with the same content every time
	// the slice is sorted. To circumvent this, we always order at least by ID.
	if len(by) == 0 ||
		(len(by) > 0 && by[len(by)-1].sortBy != sortProperty(taskPropertyID)) { // Don't sort by ID last if the id parameter is already passed as the last parameter.
		by = append(by, &sortParam{sortBy: sortProperty(taskPropertyID), orderBy: orderAscending})
	}

	comparators := make([]taskComparator, 0, len(by))
	for _, param := range by {
		comparator, ok := propertyComparators[string(param.sortBy)]
		if !ok {
			panic("No suitable comparator for sortBy found! Param was " + param.sortBy)
		}

		// This is a descending sort, so we need to negate the comparator (i.e. switch the inputs).
		if param.orderBy == orderDescending {
			oldComparator := comparator
			comparator = func(lhs, rhs *Task) int64 {
				return oldComparator(lhs, rhs) * -1
			}
		}

		comparators = append(comparators, comparator)
	}

	combinedComparator := combineComparators(comparators...)

	sort.Slice(tasks, func(i, j int) bool {
		lhs, rhs := tasks[i], tasks[j]

		res := combinedComparator(lhs, rhs)
		return res <= 0
	})
}
