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

package models

import (
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
	"xorm.io/xorm"
)

// TaskCollection is a struct used to hold filter details and not clutter the Task struct with information not related to actual tasks.
type TaskCollection struct {
	ProjectID int64 `param:"project" json:"-"`

	// The query parameter to sort by. This is for ex. done, priority, etc.
	SortBy    []string `query:"sort_by" json:"sort_by"`
	SortByArr []string `query:"sort_by[]" json:"-"`
	// The query parameter to order the items by. This can be either asc or desc, with asc being the default.
	OrderBy    []string `query:"order_by" json:"order_by"`
	OrderByArr []string `query:"order_by[]" json:"-"`

	// The filter query to match tasks by. Check out https://vikunja.io/docs/filters for a full explanation.
	Filter string `query:"filter" json:"filter"`
	// The time zone which should be used for date match (statements like "now" resolve to different actual times)
	FilterTimezone string `query:"filter_timezone" json:"-"`

	// If set to true, the result will also include null values
	FilterIncludeNulls bool `query:"filter_include_nulls" json:"filter_include_nulls"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

func validateTaskField(fieldName string) error {
	switch fieldName {
	case
		taskPropertyID,
		taskPropertyTitle,
		taskPropertyDescription,
		taskPropertyDone,
		taskPropertyDoneAt,
		taskPropertyDueDate,
		taskPropertyCreatedByID,
		taskPropertyProjectID,
		taskPropertyRepeatAfter,
		taskPropertyPriority,
		taskPropertyStartDate,
		taskPropertyEndDate,
		taskPropertyHexColor,
		taskPropertyPercentDone,
		taskPropertyUID,
		taskPropertyCreated,
		taskPropertyUpdated,
		taskPropertyPosition,
		taskPropertyKanbanPosition,
		taskPropertyBucketID,
		taskPropertyIndex:
		return nil
	}
	return ErrInvalidTaskField{TaskField: fieldName}
}

func getTaskFilterOptsFromCollection(tf *TaskCollection) (opts *taskSearchOptions, err error) {
	if len(tf.SortByArr) > 0 {
		tf.SortBy = append(tf.SortBy, tf.SortByArr...)
	}

	if len(tf.OrderByArr) > 0 {
		tf.OrderBy = append(tf.OrderBy, tf.OrderByArr...)
	}

	var sort = make([]*sortParam, 0, len(tf.SortBy))
	for i, s := range tf.SortBy {
		param := &sortParam{
			sortBy:  s,
			orderBy: orderAscending,
		}
		// This checks if tf.OrderBy has an entry with the same index as the current entry from tf.SortBy
		// Taken from https://stackoverflow.com/a/27252199/10924593
		if len(tf.OrderBy) > i {
			param.orderBy = getSortOrderFromString(tf.OrderBy[i])
		}

		// Param validation
		if err := param.validate(); err != nil {
			return nil, err
		}
		sort = append(sort, param)
	}

	opts = &taskSearchOptions{
		sortby:             sort,
		filterIncludeNulls: tf.FilterIncludeNulls,
		filter:             tf.Filter,
		filterTimezone:     tf.FilterTimezone,
	}

	opts.parsedFilters, err = getTaskFiltersFromFilterString(tf.Filter, tf.FilterTimezone)
	return opts, err
}

// ReadAll gets all tasks for a collection
// @Summary Get tasks in a project
// @Description Returns all tasks for the current project.
// @tags task
// @Accept json
// @Produce json
// @Param id path int true "The project ID."
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search tasks by task text."
// @Param sort_by query string false "The sorting parameter. You can pass this multiple times to get the tasks ordered by multiple different parametes, along with `order_by`. Possible values to sort by are `id`, `title`, `description`, `done`, `done_at`, `due_date`, `created_by_id`, `project_id`, `repeat_after`, `priority`, `start_date`, `end_date`, `hex_color`, `percent_done`, `uid`, `created`, `updated`. Default is `id`."
// @Param order_by query string false "The ordering parameter. Possible values to order by are `asc` or `desc`. Default is `asc`."
// @Param filter query string false "The filter query to match tasks by. Check out https://vikunja.io/docs/filters for a full explanation of the feature."
// @Param filter_timezone query string false "The time zone which should be used for date match (statements like "now" resolve to different actual times)"
// @Param filter_include_nulls query string false "If set to true the result will include filtered fields whose value is set to `null`. Available values are `true` or `false`. Defaults to `false`."
// @Security JWTKeyAuth
// @Success 200 {array} models.Task "The tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id}/tasks [get]
func (tf *TaskCollection) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {

	// If the project id is < -1 this means we're dealing with a saved filter - in that case we get and populate the filter
	// -1 is the favorites project which works as intended
	if tf.ProjectID < -1 {
		sf, err := getSavedFilterSimpleByID(s, getSavedFilterIDFromProjectID(tf.ProjectID))
		if err != nil {
			return nil, 0, 0, err
		}

		// By prepending sort options before the saved ones from the filter, we make sure the supplied sort
		// options via query take precedence over the rest.

		sortby := append(tf.SortBy, tf.SortByArr...)
		sortby = append(sortby, sf.Filters.SortBy...)
		sortby = append(sortby, sf.Filters.SortByArr...)

		orderby := append(tf.OrderBy, tf.OrderByArr...)
		orderby = append(orderby, sf.Filters.OrderBy...)
		orderby = append(orderby, sf.Filters.OrderByArr...)

		sf.Filters.SortBy = sortby
		sf.Filters.SortByArr = nil
		sf.Filters.OrderBy = orderby
		sf.Filters.OrderByArr = nil

		if sf.Filters.FilterTimezone == "" {
			u, err := user.GetUserByID(s, a.GetID())
			if err != nil {
				return nil, 0, 0, err
			}
			sf.Filters.FilterTimezone = u.Timezone
		}

		return sf.getTaskCollection().ReadAll(s, a, search, page, perPage)
	}

	taskopts, err := getTaskFilterOptsFromCollection(tf)
	if err != nil {
		return nil, 0, 0, err
	}

	taskopts.search = search
	taskopts.page = page
	taskopts.perPage = perPage

	shareAuth, is := a.(*LinkSharing)
	if is {
		project, err := GetProjectSimpleByID(s, shareAuth.ProjectID)
		if err != nil {
			return nil, 0, 0, err
		}
		return getTasksForProjects(s, []*Project{project}, a, taskopts)
	}

	// If the project ID is not set, we get all tasks for the user.
	// This allows to use this function in Task.ReadAll with a possibility to deprecate the latter at some point.
	var projects []*Project
	if tf.ProjectID == 0 {
		projects, _, _, err = getRawProjectsForUser(
			s,
			&projectOptions{
				user: &user.User{ID: a.GetID()},
				page: -1,
			},
		)
		if err != nil {
			return nil, 0, 0, err
		}
	} else {
		// Check the project exists and the user has access on it
		project := &Project{ID: tf.ProjectID}
		canRead, _, err := project.CanRead(s, a)
		if err != nil {
			return nil, 0, 0, err
		}
		if !canRead {
			return nil, 0, 0, ErrUserDoesNotHaveAccessToProject{
				ProjectID: tf.ProjectID,
				UserID:    a.GetID(),
			}
		}
		projects = []*Project{{ID: tf.ProjectID}}
	}

	return getTasksForProjects(s, projects, a, taskopts)
}
