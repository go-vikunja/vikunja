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

type (
	sortParam struct {
		sortBy  string
		orderBy sortOrder // asc or desc
	}

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
	taskPropertyPosition      string = "position"
)

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
	return validateTaskField(sp.sortBy)
}
