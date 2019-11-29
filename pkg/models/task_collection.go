// Copyright 2019 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/web"
	"time"
)

// TaskCollection is a struct used to hold filter details and not clutter the Task struct with information not related to actual tasks.
type TaskCollection struct {
	ListID            int64  `param:"list"`
	Sorting           string `query:"sort"` // Parameter to sort by
	StartDateSortUnix int64  `query:"startdate"`
	EndDateSortUnix   int64  `query:"enddate"`
	Lists             []*List

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// ReadAll gets all tasks for a collection
// @Summary Get tasks on a list
// @Description Returns all tasks for the current list.
// @tags task
// @Accept json
// @Produce json
// @Param listID path int true "The list ID."
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search tasks by task text."
// @Param sort query string false "The sorting parameter. Possible values to sort by are priority, prioritydesc, priorityasc, duedate, duedatedesc, duedateasc."
// @Param startdate query int false "The start date parameter to filter by. Expects a unix timestamp. If no end date, but a start date is specified, the end date is set to the current time."
// @Param enddate query int false "The end date parameter to filter by. Expects a unix timestamp. If no start date, but an end date is specified, the start date is set to the current time."
// @Security JWTKeyAuth
// @Success 200 {array} models.Task "The tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/tasks [get]
func (tf *TaskCollection) ReadAll(a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	var sortby SortBy
	switch tf.Sorting {
	case "priority":
		sortby = SortTasksByPriorityDesc
	case "prioritydesc":
		sortby = SortTasksByPriorityDesc
	case "priorityasc":
		sortby = SortTasksByPriorityAsc
	case "duedate":
		sortby = SortTasksByDueDateDesc
	case "duedatedesc":
		sortby = SortTasksByDueDateDesc
	case "duedateasc":
		sortby = SortTasksByDueDateAsc
	default:
		sortby = SortTasksByUnsorted
	}

	taskopts := &taskOptions{
		search:    search,
		sortby:    sortby,
		startDate: time.Unix(tf.StartDateSortUnix, 0),
		endDate:   time.Unix(tf.EndDateSortUnix, 0),
		page:      page,
		perPage:   perPage,
	}

	shareAuth, is := a.(*LinkSharing)
	if is {
		list := &List{ID: shareAuth.ListID}
		err := list.GetSimpleByID()
		if err != nil {
			return nil, 0, 0, err
		}
		return getTasksForLists([]*List{list}, taskopts)
	}

	// If the list ID is not set, we get all tasks for the user.
	// This allows to use this function in Task.ReadAll with a possibility to deprecate the latter at some point.
	if tf.ListID == 0 {
		tf.Lists, _, _, err = getRawListsForUser("", &User{ID: a.GetID()}, -1, 0)
		if err != nil {
			return nil, 0, 0, err
		}
	} else {
		// Check the list exists and the user has acess on it
		list := &List{ID: tf.ListID}
		canRead, err := list.CanRead(a)
		if err != nil {
			return nil, 0, 0, err
		}
		if !canRead {
			return nil, 0, 0, ErrUserDoesNotHaveAccessToList{ListID: tf.ListID}
		}
		tf.Lists = []*List{{ID: tf.ListID}}
	}

	return getTasksForLists(tf.Lists, taskopts)
}
