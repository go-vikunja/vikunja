// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
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

package models

type (
	sortParam struct {
		sortBy  string
		orderBy sortOrder // asc or desc
	}

	sortOrder string
)

const (
	taskPropertyID          string = "id"
	taskPropertyTitle       string = "title"
	taskPropertyDescription string = "description"
	taskPropertyDone        string = "done"
	taskPropertyDoneAt      string = "done_at"
	taskPropertyDueDate     string = "due_date"
	taskPropertyCreatedByID string = "created_by_id"
	taskPropertyListID      string = "list_id"
	taskPropertyRepeatAfter string = "repeat_after"
	taskPropertyPriority    string = "priority"
	taskPropertyStartDate   string = "start_date"
	taskPropertyEndDate     string = "end_date"
	taskPropertyHexColor    string = "hex_color"
	taskPropertyPercentDone string = "percent_done"
	taskPropertyUID         string = "uid"
	taskPropertyCreated     string = "created"
	taskPropertyUpdated     string = "updated"
	taskPropertyPosition    string = "position"
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
