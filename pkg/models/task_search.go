// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2023 Vikunja and contributors. All rights reserved.
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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/web"
	"github.com/typesense/typesense-go/typesense/api"
	"github.com/typesense/typesense-go/typesense/api/pointer"
	"strconv"
	"strings"

	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type taskSearcher interface {
	Search(opts *taskSearchOptions) (tasks []*Task, totalCount int64, err error)
}

type dbTaskSearcher struct {
	s                   *xorm.Session
	a                   web.Auth
	hasFavoritesProject bool
}

func (d *dbTaskSearcher) Search(opts *taskSearchOptions) (tasks []*Task, totalCount int64, err error) {
	// Since xorm does not use placeholders for order by, it is possible to expose this with sql injection if we're directly
	// passing user input to the db.
	// As a workaround to prevent this, we check for valid column names here prior to passing it to the db.
	var orderby string
	for i, param := range opts.sortby {
		// Validate the params
		if err := param.validate(); err != nil {
			return nil, totalCount, err
		}

		// Mysql sorts columns with null values before ones without null value.
		// Because it does not have support for NULLS FIRST or NULLS LAST we work around this by
		// first sorting for null (or not null) values and then the order we actually want to.
		if db.Type() == schemas.MYSQL {
			orderby += "`" + param.sortBy + "` IS NULL, "
		}

		orderby += "`" + param.sortBy + "` " + param.orderBy.String()

		// Postgres and sqlite allow us to control how columns with null values are sorted.
		// To make that consistent with the sort order we have and other dbms, we're adding a separate clause here.
		if db.Type() == schemas.POSTGRES || db.Type() == schemas.SQLITE {
			orderby += " NULLS LAST"
		}

		if (i + 1) < len(opts.sortby) {
			orderby += ", "
		}
	}

	// Some filters need a special treatment since they are in a separate table
	reminderFilters := []builder.Cond{}
	assigneeFilters := []builder.Cond{}
	labelFilters := []builder.Cond{}
	projectFilters := []builder.Cond{}

	var filters = make([]builder.Cond, 0, len(opts.filters))
	// To still find tasks with nil values, we exclude 0s when comparing with >/< values.
	for _, f := range opts.filters {
		if f.field == "reminders" {
			f.field = "reminder" // This is the name in the db
			filter, err := getFilterCond(f, opts.filterIncludeNulls)
			if err != nil {
				return nil, totalCount, err
			}
			reminderFilters = append(reminderFilters, filter)
			continue
		}

		if f.field == "assignees" {
			if f.comparator == taskFilterComparatorLike {
				return nil, totalCount, err
			}
			f.field = "username"
			filter, err := getFilterCond(f, opts.filterIncludeNulls)
			if err != nil {
				return nil, totalCount, err
			}
			assigneeFilters = append(assigneeFilters, filter)
			continue
		}

		if f.field == "labels" || f.field == "label_id" {
			f.field = "label_id"
			filter, err := getFilterCond(f, opts.filterIncludeNulls)
			if err != nil {
				return nil, totalCount, err
			}
			labelFilters = append(labelFilters, filter)
			continue
		}

		if f.field == "parent_project" || f.field == "parent_project_id" {
			f.field = "parent_project_id"
			filter, err := getFilterCond(f, opts.filterIncludeNulls)
			if err != nil {
				return nil, totalCount, err
			}
			projectFilters = append(projectFilters, filter)
			continue
		}

		filter, err := getFilterCond(f, opts.filterIncludeNulls)
		if err != nil {
			return nil, totalCount, err
		}
		filters = append(filters, filter)
	}

	// Then return all tasks for that projects
	var where builder.Cond

	if opts.search != "" {
		where =
			builder.Or(
				db.ILIKE("title", opts.search),
				db.ILIKE("description", opts.search),
			)

		searchIndex := getTaskIndexFromSearchString(opts.search)
		if searchIndex > 0 {
			where = builder.Or(where, builder.Eq{"`index`": searchIndex})
		}
	}

	var projectIDCond builder.Cond
	var favoritesCond builder.Cond
	if len(opts.projectIDs) > 0 {
		projectIDCond = builder.In("project_id", opts.projectIDs)
	}

	if d.hasFavoritesProject {
		// All favorite tasks for that user
		favCond := builder.
			Select("entity_id").
			From("favorites").
			Where(
				builder.And(
					builder.Eq{"user_id": d.a.GetID()},
					builder.Eq{"kind": FavoriteKindTask},
				))

		favoritesCond = builder.In("id", favCond)
	}

	if len(reminderFilters) > 0 {
		filters = append(filters, getFilterCondForSeparateTable("task_reminders", opts.filterConcat, reminderFilters))
	}

	if len(assigneeFilters) > 0 {
		assigneeFilter := []builder.Cond{
			builder.In("user_id",
				builder.Select("id").
					From("users").
					Where(builder.Or(assigneeFilters...)),
			)}
		filters = append(filters, getFilterCondForSeparateTable("task_assignees", opts.filterConcat, assigneeFilter))
	}

	if len(labelFilters) > 0 {
		filters = append(filters, getFilterCondForSeparateTable("label_tasks", opts.filterConcat, labelFilters))
	}

	if len(projectFilters) > 0 {
		var filtercond builder.Cond
		if opts.filterConcat == filterConcatOr {
			filtercond = builder.Or(projectFilters...)
		}
		if opts.filterConcat == filterConcatAnd {
			filtercond = builder.And(projectFilters...)
		}

		cond := builder.In(
			"project_id",
			builder.
				Select("id").
				From("projects").
				Where(filtercond),
		)
		filters = append(filters, cond)
	}

	var filterCond builder.Cond
	if len(filters) > 0 {
		if opts.filterConcat == filterConcatOr {
			filterCond = builder.Or(filters...)
		}
		if opts.filterConcat == filterConcatAnd {
			filterCond = builder.And(filters...)
		}
	}

	limit, start := getLimitFromPageIndex(opts.page, opts.perPage)
	cond := builder.And(builder.Or(projectIDCond, favoritesCond), where, filterCond)

	query := d.s.Where(cond)
	if limit > 0 {
		query = query.Limit(limit, start)
	}

	tasks = []*Task{}
	err = query.OrderBy(orderby).Find(&tasks)
	if err != nil {
		return nil, totalCount, err
	}

	queryCount := d.s.Where(cond)
	totalCount, err = queryCount.
		Count(&Task{})
	if err != nil {
		return nil, totalCount, err

	}

	return
}

type typesenseTaskSearcher struct {
	s *xorm.Session
}

func (t *typesenseTaskSearcher) Search(opts *taskSearchOptions) (tasks []*Task, totalCount int64, err error) {
	projectIDStrings := []string{}
	for _, id := range opts.projectIDs {
		projectIDStrings = append(projectIDStrings, strconv.FormatInt(id, 10))
	}

	params := &api.SearchCollectionParams{
		Q:                opts.search,
		QueryBy:          "title, description, comments.comment",
		Page:             pointer.Int(opts.page),
		PerPage:          pointer.Int(opts.perPage),
		ExhaustiveSearch: pointer.True(),
		FilterBy:         pointer.String("project_id: [" + strings.Join(projectIDStrings, ", ") + "]"),
	}

	result, err := typesenseClient.Collection("tasks").
		Documents().
		Search(params)
	if err != nil {
		return
	}

	taskIDs := []int64{}
	for _, h := range *result.Hits {
		hit := *h.Document
		taskID, err := strconv.ParseInt(hit["id"].(string), 10, 64)
		if err != nil {
			return nil, 0, err
		}
		taskIDs = append(taskIDs, taskID)
	}

	tasks = []*Task{}

	err = t.s.In("id", taskIDs).Find(&tasks)
	return tasks, int64(*result.Found), err
}
