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

package models

import (
	"regexp"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// TaskCollection is a struct used to hold filter details and not clutter the Task struct with information not related to actual tasks.
type TaskCollection struct {
	ProjectID     int64 `param:"project" json:"-"`
	ProjectViewID int64 `param:"view" json:"-"`

	Search string `query:"s" json:"s"`

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

	// If set to `subtasks`, Vikunja will fetch only tasks which do not have subtasks and then in a
	// second step, will fetch all of these subtasks. This may result in more tasks than the
	// pagination limit being returned, but all subtasks will be present in the response.
	// If set to `buckets`, the buckets of each task will be present in the response.
	// If set to `reactions`, the reactions of each task will be present in the response.
	// If set to `comments`, the first 50 comments of each task will be present in the response.
	// You can set this multiple times with different values.
	Expand []TaskCollectionExpandable `query:"expand" json:"-"`

	isSavedFilter bool

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

type TaskCollectionExpandable string

const TaskCollectionExpandSubtasks TaskCollectionExpandable = `subtasks`
const TaskCollectionExpandBuckets TaskCollectionExpandable = `buckets`
const TaskCollectionExpandReactions TaskCollectionExpandable = `reactions`
const TaskCollectionExpandComments TaskCollectionExpandable = `comments`

// Validate validates if the TaskCollectionExpandable value is valid.
func (t TaskCollectionExpandable) Validate() error {
	switch t {
	case TaskCollectionExpandSubtasks:
		return nil
	case TaskCollectionExpandBuckets:
		return nil
	case TaskCollectionExpandReactions:
		return nil
	case TaskCollectionExpandComments:
		return nil
	}

	return InvalidFieldErrorWithMessage([]string{"expand"}, "Expand must be one of the following values: subtasks, buckets, reactions")
}

func validateTaskField(fieldName string) error {
	switch fieldName {
	case
		taskPropertyAssignees,
		taskPropertyLabels,
		taskPropertyReminders:
		return nil
	}

	return ValidateTaskFieldForSorting(fieldName)
}

func getTaskFilterOptsFromCollection(tf *TaskCollection, projectView *ProjectView) (opts *taskSearchOptions, err error) {
	if len(tf.SortByArr) > 0 {
		tf.SortBy = append(tf.SortBy, tf.SortByArr...)
	}

	if len(tf.OrderByArr) > 0 {
		tf.OrderBy = append(tf.OrderBy, tf.OrderByArr...)
	}

	opts = &taskSearchOptions{
		sortby:             tf.SortBy,
		filterIncludeNulls: tf.FilterIncludeNulls,
		filter:             tf.Filter,
		filterTimezone:     tf.FilterTimezone,
	}

	if projectView != nil {
		opts.projectViewID = projectView.ID
	}

	opts.parsedFilters, err = getTaskFiltersFromFilterString(tf.Filter, tf.FilterTimezone)
	return opts, err
}

func getTaskOrTasksInBuckets(s *xorm.Session, u *user.User, projects []*Project, view *ProjectView, opts *taskSearchOptions, filteringForBucket bool) (tasks interface{}, resultCount int, totalItems int64, err error) {
	if filteringForBucket {
		return getTasksForProjects(s, projects, u, opts, view)
	}

	if view != nil && !strings.Contains(opts.filter, taskPropertyBucketID) {
		if view.BucketConfigurationMode != BucketConfigurationModeNone {
			tasksInBuckets, err := GetTasksInBucketsForView(s, view, projects, opts, u)
			return tasksInBuckets, len(tasksInBuckets), int64(len(tasksInBuckets)), err
		}
	}

	return getTasksForProjects(s, projects, u, opts, view)
}

func getRelevantProjectsFromCollection(s *xorm.Session, u *user.User, tf *TaskCollection) (projects []*Project, err error) {
	if tf.ProjectID == 0 || tf.isSavedFilter {
		projects, _, _, err = getRawProjectsForUser(
			s,
			&ProjectOptions{
				User: u,
				Page: -1,
			},
		)
		return projects, err
	}

	// Check the project exists and the user has access on it
	project := &Project{ID: tf.ProjectID}
	canRead, _, err := project.CanRead(s, u)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, ErrUserDoesNotHaveAccessToProject{
			ProjectID: tf.ProjectID,
			UserID:    u.ID,
		}
	}

	return []*Project{{ID: tf.ProjectID}}, nil
}

func getFilterValueForBucketFilter(filter string, view *ProjectView) (newFilter string, err error) {
	if view.BucketConfigurationMode != BucketConfigurationModeFilter {
		return filter, nil
	}

	re := regexp.MustCompile(`bucket_id\s*=\s*(\d+)`)

	match := re.FindStringSubmatch(filter)
	if len(match) < 2 {
		return filter, nil
	}

	bucketID, err := strconv.Atoi(match[1])
	if err != nil {
		return "", err
	}

	for id, bucket := range view.BucketConfiguration {
		if id == bucketID {
			return re.ReplaceAllString(filter, `(`+bucket.Filter.Filter+`)`), nil
		}
	}

	return filter, nil
}

// ReadAll gets all tasks for a collection
// @Summary Get tasks in a project
// @Description Returns all tasks for the selected project. When the requested view is a kanban view, a list of buckets containing the tasks will be returned. Otherwise, a list of tasks will be returned.
// @tags task
// @Accept json
// @Produce json
// @Param id path int true "The project ID."
// @Param view path int true "The project view ID."
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search tasks by task text."
// @Param sort_by query string false "The sorting parameter. You can pass this multiple times to get the tasks ordered by multiple different parametes, along with `order_by`. Possible values to sort by are `id`, `title`, `description`, `done`, `done_at`, `due_date`, `created_by_id`, `project_id`, `repeat_after`, `priority`, `start_date`, `end_date`, `hex_color`, `percent_done`, `uid`, `created`, `updated`. Default is `id`."
// @Param order_by query string false "The ordering parameter. Possible values to order by are `asc` or `desc`. Default is `asc`."
// @Param filter query string false "The filter query to match tasks by. Check out https://vikunja.io/docs/filters for a full explanation of the feature."
// @Param filter_timezone query string false "The time zone which should be used for date match (statements like "now" resolve to different actual times)"
// @Param filter_include_nulls query string false "If set to true the result will include filtered fields whose value is set to `null`. Available values are `true` or `false`. Defaults to `false`."
// @Param expand query array false "If set to `subtasks`, Vikunja will fetch only tasks which do not have subtasks and then in a second step, will fetch all of these subtasks. This may result in more tasks than the pagination limit being returned, but all subtasks will be present in the response. If set to `buckets`, the buckets of each task will be present in the response. If set to `reactions`, the reactions of each task will be present in the response. If set to `comments`, the first 50 comments of each task will be present in the response. You can set this multiple times with different values."
// @Security JWTKeyAuth
// @Success 200 {array} models.Task "The tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id}/views/{view}/tasks [get]
func (tf *TaskCollection) ReadAllTasks(s *xorm.Session, u *user.User, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	if tf.ProjectID != 0 {
		p, err := GetProjectSimpleByID(s, tf.ProjectID)
		if err != nil {
			return nil, 0, 0, err
		}
		tasks, resultCount, totalItems, err := getTasksForProjects(s, []*Project{p}, u, &taskSearchOptions{
			search:  search,
			page:    page,
			perPage: perPage,
		}, nil)
		if err != nil {
			return nil, 0, 0, err
		}
		return tasks, resultCount, totalItems, nil
	}
	return tf.ReadAllTasks(s, u, search, page, perPage)
}

func (tf *TaskCollection) ReadAll(s *xorm.Session, u *user.User, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	// If the project id is < -1 this means we're dealing with a saved filter - in that case we get and populate the filter
	// -1 is the favorites project which works as intended
	if !tf.isSavedFilter && tf.ProjectID < -1 {
		sf, err := GetSavedFilterSimpleByID(s, GetSavedFilterIDFromProjectID(tf.ProjectID))
		if err != nil {
			return nil, 0, 0, err
		}

		canRead, _, err := sf.CanRead(s, u)
		if err != nil {
			return nil, 0, 0, err
		}
		if !canRead {
			return nil, 0, 0, ErrGenericForbidden{}
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
			sf.Filters.FilterTimezone = u.Timezone
		}

		tc := sf.getTaskCollection()
		tc.ProjectViewID = tf.ProjectViewID
		tc.ProjectID = tf.ProjectID
		tc.isSavedFilter = true

		if tf.Filter != "" {
			if tc.Filter != "" {
				tc.Filter = "(" + tf.Filter + ") && (" + tc.Filter + ")"
			} else {
				tc.Filter = tf.Filter
			}
		}

		return tc.ReadAll(s, u, search, page, perPage)
	}

	var view *ProjectView
	var filteringForBucket bool
	if tf.ProjectViewID != 0 {
		view, err = GetProjectViewByIDAndProject(s, tf.ProjectViewID, tf.ProjectID)
		if err != nil {
			return nil, 0, 0, err
		}

		if view.Filter != nil {
			if view.Filter.Filter != "" {
				if tf.Filter != "" {
					tf.Filter = "(" + tf.Filter + ") && (" + view.Filter.Filter + ")"
				} else {
					tf.Filter = view.Filter.Filter
				}
			}

			if view.Filter.FilterTimezone != "" {
				tf.FilterTimezone = view.Filter.FilterTimezone
			}

			if view.Filter.FilterIncludeNulls {
				tf.FilterIncludeNulls = view.Filter.FilterIncludeNulls
			}

			if view.Filter.Search != "" {
				search = view.Filter.Search
			}
		}

		if strings.Contains(tf.Filter, taskPropertyBucketID) {
			filteringForBucket = true
			tf.Filter, err = getFilterValueForBucketFilter(tf.Filter, view)
			if err != nil {
				return nil, 0, 0, err
			}
		}
	}

	opts, err := getTaskFilterOptsFromCollection(tf, view)
	if err != nil {
		return nil, 0, 0, err
	}

	for _, expandValue := range tf.Expand {
		err = expandValue.Validate()
		if err != nil {
			return nil, 0, 0, err
		}
	}

	opts.search = search
	opts.page = page
	opts.perPage = perPage
	opts.expand = tf.Expand
	opts.isSavedFilter = tf.isSavedFilter

	if view != nil {
		var hasOrderByPosition bool
		for _, param := range opts.sortby {
			if param == "position" {
				hasOrderByPosition = true
				break
			}
		}
		if !hasOrderByPosition {
			opts.sortby = append(opts.sortby, "position")
		}
	}

	if u == nil {
		// This should not happen, because we check for link shares before, but better be safe than sorry
		return nil, 0, 0, ErrGenericForbidden{}
	}

	projects, err := getRelevantProjectsFromCollection(s, u, tf)
	if err != nil {
		return nil, 0, 0, err
	}

	return getTaskOrTasksInBuckets(s, u, projects, view, opts, filteringForBucket)
}
