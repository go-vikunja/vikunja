/*
 * Copyright (c) 2018 the Vikunja Authors. All rights reserved.
 * Use of this source code is governed by a LPGLv3-style
 * license that can be found in the LICENSE file.
 */

package models

import "code.vikunja.io/web"

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

// ReadAllWithPriority gets all tasks for a user, sorted
// @Summary Get tasks sorted
// @Description Returns all tasks on any list the user has access to.
// @tags task
// @Accept json
// @Produce json
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search tasks by task text."
// @Param sortby path string true "The sorting parameter. Possible values to sort by are priority, prioritydesc, priorityasc, dueadate, dueadatedesc, dueadateasc."
// @Security ApiKeyAuth
// @Success 200 {array} models.List "The tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{sortby} [get]
func dummy() {
	// Dummy function for swaggo to pick up the docs comment
}

// ReadAll gets all tasks for a user
// @Summary Get tasks
// @Description Returns all tasks on any list the user has access to.
// @tags task
// @Accept json
// @Produce json
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search tasks by task text."
// @Security ApiKeyAuth
// @Success 200 {array} models.List "The tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks [get]
func (lt *ListTask) ReadAll(search string, a web.Auth, page int) (interface{}, error) {
	u, err := getUserWithError(a)
	if err != nil {
		return nil, err
	}

	var sortby SortBy
	switch lt.Sorting {
	case "priority":
		sortby = SortTasksByPriorityDesc
	case "prioritydesc":
		sortby = SortTasksByPriorityDesc
	case "priorityasc":
		sortby = SortTasksByPriorityAsc
	case "dueadate":
		sortby = SortTasksByDueDateDesc
	case "dueadatedesc":
		sortby = SortTasksByDueDateDesc
	case "duedateasc":
		sortby = SortTasksByDueDateAsc
	default:
		sortby = SortTasksByUnsorted
	}

	return GetTasksByUser(search, u, page, sortby)
}

//GetTasksByUser returns all tasks for a user
func GetTasksByUser(search string, u *User, page int, sortby SortBy) (tasks []*ListTask, err error) {
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

	// Then return all tasks for that lists
	if err := x.In("list_id", listIDs).Where("text LIKE ?", "%"+search+"%").OrderBy(orderby).Find(&tasks); err != nil {
		return nil, err
	}

	return tasks, err
}
