/*
 * Copyright (c) 2018 the Vikunja Authors. All rights reserved.
 * Use of this source code is governed by a LPGLv3-style
 * license that can be found in the LICENSE file.
 */

package models

import (
	"code.vikunja.io/web"
	"sort"
	"time"
)

// SortBy declares constants to sort
type SortBy int

// These are possible sort options
const (
	SortTasksByUnsorted   SortBy = -1
	SortTasksByDueDateAsc        = iota
	SortTasksByDueDateDesc
	SortTasksByPriorityAsc
	SortTasksByPriorityDesc
)

// ReadAll gets all tasks for a user
// @Summary Get tasks
// @Description Returns all tasks on any list the user has access to.
// @tags task
// @Accept json
// @Produce json
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search tasks by task text."
// @Param sort query string false "The sorting parameter. Possible values to sort by are priority, prioritydesc, priorityasc, duedate, duedatedesc, duedateasc."
// @Param startdate query int false "The start date parameter to filter by. Expects a unix timestamp. If no end date, but a start date is specified, the end date is set to the current time."
// @Param enddate query int false "The end date parameter to filter by. Expects a unix timestamp. If no start date, but an end date is specified, the start date is set to the current time."
// @Security JWTKeyAuth
// @Success 200 {array} models.ListTask "The tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/all [get]
func (t *ListTask) ReadAll(search string, a web.Auth, page int) (interface{}, error) {
	u, err := getUserWithError(a)
	if err != nil {
		return nil, err
	}

	var sortby SortBy
	switch t.Sorting {
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

	return GetTasksByUser(search, u, page, sortby, time.Unix(t.StartDateSortUnix, 0), time.Unix(t.EndDateSortUnix, 0))
}

//GetTasksByUser returns all tasks for a user
func GetTasksByUser(search string, u *User, page int, sortby SortBy, startDate time.Time, endDate time.Time) ([]*ListTask, error) {
	// Get all lists
	lists, err := getRawListsForUser("", u, page)
	if err != nil {
		return nil, err
	}

	// Get all list IDs and get the tasks
	var listIDs []int64
	for _, l := range lists {
		listIDs = append(listIDs, l.ID)
	}

	var orderby string
	switch sortby {
	case SortTasksByPriorityDesc:
		orderby = "priority desc"
	case SortTasksByPriorityAsc:
		orderby = "priority asc"
	case SortTasksByDueDateDesc:
		orderby = "due_date_unix desc"
	case SortTasksByDueDateAsc:
		orderby = "due_date_unix asc"
	}

	taskMap := make(map[int64]*ListTask)

	// Then return all tasks for that lists
	if startDate.Unix() != 0 || endDate.Unix() != 0 {

		startDateUnix := time.Now().Unix()
		if startDate.Unix() != 0 {
			startDateUnix = startDate.Unix()
		}

		endDateUnix := time.Now().Unix()
		if endDate.Unix() != 0 {
			endDateUnix = endDate.Unix()
		}

		if err := x.In("list_id", listIDs).
			Where("text LIKE ?", "%"+search+"%").
			And("((due_date_unix BETWEEN ? AND ?) OR "+
				"(start_date_unix BETWEEN ? and ?) OR "+
				"(end_date_unix BETWEEN ? and ?))", startDateUnix, endDateUnix, startDateUnix, endDateUnix, startDateUnix, endDateUnix).
			And("(parent_task_id = 0 OR parent_task_id IS NULL)").
			OrderBy(orderby).
			Find(&taskMap); err != nil {
			return nil, err
		}
	} else {
		if err := x.In("list_id", listIDs).
			Where("text LIKE ?", "%"+search+"%").
			And("(parent_task_id = 0 OR parent_task_id IS NULL)").
			OrderBy(orderby).
			Find(&taskMap); err != nil {
			return nil, err
		}
	}

	tasks, err := addMoreInfoToTasks(taskMap)
	if err != nil {
		return nil, err
	}
	// Because the list is sorted by id which we don't want (since we're dealing with maps)
	// we have to manually sort the tasks again here.
	sortTasks(tasks, sortby)

	return tasks, err
}

func sortTasks(tasks []*ListTask, by SortBy) {
	switch by {
	case SortTasksByPriorityDesc:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Priority > tasks[j].Priority
		})
	case SortTasksByPriorityAsc:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Priority < tasks[j].Priority
		})
	case SortTasksByDueDateDesc:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].DueDateUnix > tasks[j].DueDateUnix
		})
	case SortTasksByDueDateAsc:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].DueDateUnix < tasks[j].DueDateUnix
		})
	}
}
