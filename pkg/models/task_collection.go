// Copyright 2018-2020 Vikunja and contriubtors. All rights reserved.
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
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
)

// TaskCollection is a struct used to hold filter details and not clutter the Task struct with information not related to actual tasks.
type TaskCollection struct {
	ListID int64   `param:"list" json:"-"`
	Lists  []*List `json:"-"`

	// The query parameter to sort by. This is for ex. done, priority, etc.
	SortBy    []string `query:"sort_by" json:"sort_by"`
	SortByArr []string `query:"sort_by[]" json:"-"`
	// The query parameter to order the items by. This can be either asc or desc, with asc being the default.
	OrderBy    []string `query:"order_by" json:"order_by"`
	OrderByArr []string `query:"order_by[]" json:"-"`

	// The field name of the field to filter by
	FilterBy    []string `query:"filter_by" json:"filter_by"`
	FilterByArr []string `query:"filter_by[]" json:"-"`
	// The value of the field name to filter by
	FilterValue    []string `query:"filter_value" json:"filter_value"`
	FilterValueArr []string `query:"filter_value[]" json:"-"`
	// The comparator for field and value
	FilterComparator    []string `query:"filter_comparator" json:"filter_comparator"`
	FilterComparatorArr []string `query:"filter_comparator[]" json:"-"`
	// The way all filter conditions are concatenated together, can be either "and" or "or".,
	FilterConcat string `query:"filter_concat" json:"filter_concat"`
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
		taskPropertyListID,
		taskPropertyRepeatAfter,
		taskPropertyPriority,
		taskPropertyStartDate,
		taskPropertyEndDate,
		taskPropertyHexColor,
		taskPropertyPercentDone,
		taskPropertyUID,
		taskPropertyCreated,
		taskPropertyUpdated,
		taskPropertyPosition:
		return nil
	}
	return ErrInvalidTaskField{TaskField: fieldName}

}

// ReadAll gets all tasks for a collection
// @Summary Get tasks in a list
// @Description Returns all tasks for the current list.
// @tags task
// @Accept json
// @Produce json
// @Param listID path int true "The list ID."
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search tasks by task text."
// @Param sort_by query string false "The sorting parameter. You can pass this multiple times to get the tasks ordered by multiple different parametes, along with `order_by`. Possible values to sort by are `id`, `title`, `description`, `done`, `done_at`, `due_date`, `created_by_id`, `list_id`, `repeat_after`, `priority`, `start_date`, `end_date`, `hex_color`, `percent_done`, `uid`, `created`, `updated`. Default is `id`."
// @Param order_by query string false "The ordering parameter. Possible values to order by are `asc` or `desc`. Default is `asc`."
// @Param filter_by query string false "The name of the field to filter by. Accepts an array for multiple filters which will be chanied together, all supplied filter must match."
// @Param filter_value query string false "The value to filter for."
// @Param filter_comparator query string false "The comparator to use for a filter. Available values are `equals`, `greater`, `greater_equals`, `less` and `less_equals`. Defaults to `equals`"
// @Param filter_concat query string false "The concatinator to use for filters. Available values are `and` or `or`. Defaults to `or`."
// @Param filter_include_nulls query string false "If set to true the result will include filtered fields whose value is set to `null`. Available values are `true` or `false`. Defaults to `false`."
// @Security JWTKeyAuth
// @Success 200 {array} models.Task "The tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/tasks [get]
func (tf *TaskCollection) ReadAll(a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {

	// If the list id is < -1 this means we're dealing with a saved filter - in that case we get and populate the filter
	// -1 is the favorites list which works as intended
	if tf.ListID < -1 {
		s, err := getSavedFilterSimpleByID(getSavedFilterIDFromListID(tf.ListID))
		if err != nil {
			return nil, 0, 0, err
		}

		return s.getTaskCollection().ReadAll(a, search, page, perPage)
	}

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
			return nil, 0, 0, err
		}
		sort = append(sort, param)
	}

	taskopts := &taskOptions{
		search:             search,
		page:               page,
		perPage:            perPage,
		sortby:             sort,
		filterConcat:       taskFilterConcatinator(tf.FilterConcat),
		filterIncludeNulls: tf.FilterIncludeNulls,
	}

	taskopts.filters, err = getTaskFiltersByCollections(tf)
	if err != nil {
		return
	}

	shareAuth, is := a.(*LinkSharing)
	if is {
		list := &List{ID: shareAuth.ListID}
		err := list.GetSimpleByID()
		if err != nil {
			return nil, 0, 0, err
		}
		return getTasksForLists([]*List{list}, a, taskopts)
	}

	// If the list ID is not set, we get all tasks for the user.
	// This allows to use this function in Task.ReadAll with a possibility to deprecate the latter at some point.
	if tf.ListID == 0 {
		tf.Lists, _, _, err = getRawListsForUser(&listOptions{
			user: &user.User{ID: a.GetID()},
			page: -1,
		})
		if err != nil {
			return nil, 0, 0, err
		}
	} else {
		// Check the list exists and the user has acess on it
		list := &List{ID: tf.ListID}
		canRead, _, err := list.CanRead(a)
		if err != nil {
			return nil, 0, 0, err
		}
		if !canRead {
			return nil, 0, 0, ErrUserDoesNotHaveAccessToList{ListID: tf.ListID}
		}
		tf.Lists = []*List{{ID: tf.ListID}}
	}

	return getTasksForLists(tf.Lists, a, taskopts)
}
