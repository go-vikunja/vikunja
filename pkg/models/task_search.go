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
	"fmt"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type SubTableFilter struct {
	Table           string
	BaseFilter      string
	FilterableField string
	AllowNullCheck  bool
}

type SubTableFilters map[string]SubTableFilter

var subTableFilters = SubTableFilters{
	"labels": {
		Table:           "label_tasks",
		BaseFilter:      "tasks.id = task_id",
		FilterableField: "label_id",
		AllowNullCheck:  true,
	},
	"label_id": {
		Table:           "label_tasks",
		BaseFilter:      "tasks.id = task_id",
		FilterableField: "label_id",
		AllowNullCheck:  true,
	},
	"reminders": {
		Table:           "task_reminders",
		BaseFilter:      "tasks.id = task_id",
		FilterableField: "reminder",
		AllowNullCheck:  true,
	},
	"assignees": {
		Table:           "task_assignees",
		BaseFilter:      "tasks.id = task_id",
		FilterableField: "username",
		AllowNullCheck:  true,
	},
	"parent_project": {
		Table:           "projects",
		BaseFilter:      "tasks.project_id = id",
		FilterableField: "parent_project_id",
		AllowNullCheck:  false,
	},
	"parent_project_id": {
		Table:           "projects",
		BaseFilter:      "tasks.project_id = id",
		FilterableField: "parent_project_id",
		AllowNullCheck:  false,
	},
}

var strictComparators = map[taskFilterComparator]bool{
	taskFilterComparatorIn:        true,
	taskFilterComparatorNotIn:     true,
	taskFilterComparatorEquals:    true,
	taskFilterComparatorNotEquals: true,
}

// isRangeComparator returns true for comparators where combining multiple
// conditions into a single EXISTS subquery is semantically correct (i.e. a
// single row can satisfy both conditions simultaneously).
func isRangeComparator(c taskFilterComparator) bool {
	return c == taskFilterComparatorGreater ||
		c == taskFilterComparatorGreateEquals ||
		c == taskFilterComparatorLess ||
		c == taskFilterComparatorLessEquals
}

type taskSearcher interface {
	Search(opts *taskSearchOptions) (tasks []*Task, totalCount int64, err error)
}

type dbTaskSearcher struct {
	s                   *xorm.Session
	a                   web.Auth
	hasFavoritesProject bool
}

func (sf *SubTableFilter) ToBaseSubQuery() *builder.Builder {
	var cond = builder.
		Select("1").
		From(sf.Table).
		Where(builder.Expr(sf.BaseFilter))

	// little hack to add users table for assignees filter
	if sf.Table == "task_assignees" {
		cond.Join("INNER", "users", "users.id = user_id")
	}

	return cond
}

func getOrderByDBStatement(opts *taskSearchOptions) (orderby string, err error) {
	// Since xorm does not use placeholders for order by, it is possible to expose this with sql injection if we're directly
	// passing user input to the db.
	// As a workaround to prevent this, we check for valid column names here prior to passing it to the db.
	for i, param := range opts.sortby {
		// Validate the params
		if err := param.validate(); err != nil {
			return "", err
		}

		var prefix string
		switch param.sortBy {
		case taskPropertyPosition:
			prefix = "task_positions."
		case taskPropertyBucketID:
			prefix = "task_buckets."
		default:
			prefix = "tasks."
		}

		// Mysql sorts columns with null values before ones without null value.
		// Because it does not have support for NULLS FIRST or NULLS LAST we work around this by
		// first sorting for null (or not null) values and then the order we actually want to.
		if db.Type() == schemas.MYSQL {
			orderby += prefix + "`" + param.sortBy + "` IS NULL, "
		}

		orderby += prefix + "`" + param.sortBy + "` " + param.orderBy.String()

		// Postgres and sqlite allow us to control how columns with null values are sorted.
		// To make that consistent with the sort order we have and other dbms, we're adding a separate clause here.
		if db.Type() == schemas.POSTGRES || db.Type() == schemas.SQLITE {
			orderby += " NULLS LAST"
		}

		if (i + 1) < len(opts.sortby) {
			orderby += ", "
		}
	}

	return
}

func convertFiltersToDBFilterCond(rawFilters []*taskFilter, includeNulls bool) (filterCond builder.Cond, err error) {

	var dbFilters = make([]builder.Cond, 0, len(rawFilters))
	// Track join types separately because after merging consecutive sub-table
	// filters, the indexes of dbFilters no longer correspond 1:1 with rawFilters.
	var dbFilterJoins = make([]taskFilterConcatinator, 0, len(rawFilters))

	for i := 0; i < len(rawFilters); i++ {
		f := rawFilters[i]

		if nested, is := f.value.([]*taskFilter); is {
			nestedDBFilters, err := convertFiltersToDBFilterCond(nested, includeNulls)
			if err != nil {
				return nil, err
			}
			dbFilters = append(dbFilters, nestedDBFilters)
			dbFilterJoins = append(dbFilterJoins, f.join)
			continue
		}

		subTableFilterParams, ok := subTableFilters[f.field]
		if ok {
			if f.field == "assignees" && (f.comparator == taskFilterComparatorLike) {
				continue
			}

			// Collect all consecutive AND-joined range filters targeting the same sub-table.
			// Only range comparators (>, >=, <, <=) are merged because they express
			// conditions a single row can satisfy simultaneously (e.g. reminder > X AND
			// reminder < Y). Equality/IN/NOT comparators must remain as separate EXISTS
			// subqueries because each matching value lives in its own row (e.g.
			// labels = 4 && labels = 5 means two different rows must each exist).
			group := []*taskFilter{f}
			if isRangeComparator(f.comparator) {
				for i+1 < len(rawFilters) {
					next := rawFilters[i+1]
					nextSubTable, nextOk := subTableFilters[next.field]
					if !nextOk || nextSubTable.Table != subTableFilterParams.Table || next.join != filterConcatAnd {
						break
					}
					if !isRangeComparator(next.comparator) {
						break
					}
					group = append(group, next)
					i++
				}
			}

			// Build the combined condition for all filters in the group
			var combinedInnerCond builder.Cond
			for _, gf := range group {
				comparator := gf.comparator
				_, isStrict := strictComparators[gf.comparator]
				if isStrict {
					comparator = taskFilterComparatorIn
				}

				innerFilter, err := getFilterCond(&taskFilter{
					field:      subTableFilterParams.FilterableField,
					value:      gf.value,
					comparator: comparator,
					isNumeric:  gf.isNumeric,
				}, false)
				if err != nil {
					return nil, err
				}

				if combinedInnerCond == nil {
					combinedInnerCond = innerFilter
				} else {
					combinedInnerCond = builder.And(combinedInnerCond, innerFilter)
				}
			}

			filterSubQuery := subTableFilterParams.ToBaseSubQuery().And(combinedInnerCond)

			var filter builder.Cond
			if f.comparator == taskFilterComparatorNotEquals || f.comparator == taskFilterComparatorNotIn {
				filter = builder.NotExists(filterSubQuery)
			} else {
				filter = builder.Exists(filterSubQuery)
			}

			if includeNulls && subTableFilterParams.AllowNullCheck {
				filter = builder.Or(filter, builder.NotExists(subTableFilterParams.ToBaseSubQuery()))
			}

			dbFilters = append(dbFilters, filter)
			// Use the join from the first filter in the group: f.join describes how
			// this group connects to the previous element (matches the convention
			// where dbFilterJoins[i+1] combines dbFilters[i] with dbFilters[i+1]).
			dbFilterJoins = append(dbFilterJoins, f.join)
			continue
		}

		if f.field == taskPropertyBucketID {
			f.field = "task_buckets.`bucket_id`"
		} else {
			f.field = "tasks.`" + f.field + "`"
		}
		filter, err := getFilterCond(f, includeNulls)
		if err != nil {
			return nil, err
		}
		dbFilters = append(dbFilters, filter)
		dbFilterJoins = append(dbFilterJoins, f.join)
	}

	if len(dbFilters) > 0 {
		filterCond = dbFilters[0]
		if len(dbFilters) >= 1 {
			for i := range dbFilters {
				if len(dbFilters) > i+1 {
					switch dbFilterJoins[i+1] {
					case filterConcatOr:
						filterCond = builder.Or(filterCond, dbFilters[i+1])
					case filterConcatAnd:
						filterCond = builder.And(filterCond, dbFilters[i+1])
					}
				}
			}
		}
	}

	return filterCond, nil
}

func hasBucketIDInParsedFilter(filters []*taskFilter) bool {
	for _, filter := range filters {
		if subfilters, is := filter.value.([]*taskFilter); is {
			has := hasBucketIDInParsedFilter(subfilters)
			if has {
				return true
			}
		}
		if filter.field == taskPropertyBucketID {
			return true
		}
	}

	return false
}

//nolint:gocyclo
func (d *dbTaskSearcher) Search(opts *taskSearchOptions) (tasks []*Task, totalCount int64, err error) {

	orderby, err := getOrderByDBStatement(opts)
	if err != nil {
		return nil, 0, err
	}

	joinTaskBuckets := hasBucketIDInParsedFilter(opts.parsedFilters)

	filterCond, err := convertFiltersToDBFilterCond(opts.parsedFilters, opts.filterIncludeNulls)
	if err != nil {
		return nil, 0, err
	}

	// Then return all tasks for that projects
	var where builder.Cond
	// textSearchCond holds only the ParadeDB/ILIKE title+description match, kept
	// separate from the index-equality match so the relevance ranking path can
	// score a pure-ParadeDB query (see rankByRelevance below).
	var textSearchCond builder.Cond
	var searchIndex int64

	if opts.search != "" {
		textSearchCond = db.MultiFieldSearchWithTableAlias([]string{"title", "description"}, opts.search, "tasks")
		where = textSearchCond

		searchIndex = getTaskIndexFromSearchString(opts.search)
		if searchIndex > 0 {
			where = builder.Or(where, builder.Eq{"tasks.`index`": searchIndex})
		}
	}

	var projectIDCond builder.Cond
	var favoritesCond builder.Cond
	if len(opts.projectIDs) > 0 {
		projectIDCond = builder.In("tasks.project_id", opts.projectIDs)
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

		favoritesCond = builder.In("tasks.id", favCond)
	}

	scopeCond := builder.Or(projectIDCond, favoritesCond)

	limit, start := getLimitFromPageIndex(opts.page, opts.perPage)
	cond := builder.And(scopeCond, where, filterCond)

	// ParadeDB exposes the BM25 relevance score via pdb.score(<key_field>) for any
	// query containing a ParadeDB operator (the ||| from MultiFieldSearch qualifies).
	// When searching without an explicit user sort, order by relevance so tasks
	// matching all query words rank above tasks matching only some. This is
	// ParadeDB-only: pdb.score is invalid SQL on sqlite/mysql/plain postgres.
	rankByRelevance := db.ParadeDBAvailable() && opts.search != "" && !opts.userProvidedSort

	// ParadeDB's pdb.score() rejects an `id IN (<subquery>)` favorites scope (whether
	// expressed as OR or UNION) as an unsupported query shape, so the relevance arms
	// reach favorites through a LEFT JOIN and scope on the joined column instead,
	// which it can score. Only relevant when favorites are part of the scope.
	rankFavoritesJoin := rankByRelevance && d.hasFavoritesProject
	rankScopeCond := scopeCond
	if rankFavoritesJoin {
		rankScopeCond = builder.Or(projectIDCond, builder.Expr("rank_favorites.entity_id IS NOT NULL"))
	}

	var distinct = "tasks.*"
	if strings.Contains(orderby, "task_positions.") {
		distinct += ", task_positions.position"
	}

	var expandSubtasks = false
	for _, expandable := range opts.expand {
		if expandable == TaskCollectionExpandSubtasks {
			expandSubtasks = true
			break
		}
	}

	// addJoins applies the same LEFT JOINs the count query and every fetch arm
	// rely on (position sort, bucket filter, subtask expansion).
	addJoins := func(query *xorm.Session) *xorm.Session {
		for _, param := range opts.sortby {
			if param.sortBy == taskPropertyPosition {
				query = query.Join("LEFT", "task_positions", "task_positions.task_id = tasks.id AND task_positions.project_view_id = ?", param.projectViewID)
				break
			}
		}
		if joinTaskBuckets {
			joinCond := "task_buckets.task_id = tasks.id"
			if opts.projectViewID > 0 {
				joinCond += " AND task_buckets.project_view_id = ?"
				query = query.Join("LEFT", "task_buckets", joinCond, opts.projectViewID)
			} else {
				query = query.Join("LEFT", "task_buckets", joinCond)
			}
		}
		if expandSubtasks {
			query = query.
				Join("LEFT", "task_relations", "tasks.id = task_relations.task_id and task_relations.relation_kind = 'parenttask'").
				Join("LEFT", "tasks parent_tasks", "task_relations.other_task_id = parent_tasks.id")
		}
		return query
	}

	subtaskParentCond := builder.Or(
		builder.IsNull{"task_relations.id"},
		builder.IsNull{"parent_tasks.id"},
		builder.Expr("parent_tasks.project_id != tasks.project_id"),
	)
	if expandSubtasks {
		cond = builder.And(cond, subtaskParentCond)
	}

	// fetchTasks runs a single fetch arm: it builds the DISTINCT select (raw, so
	// xorm doesn't quote-corrupt the pdb.score function call), applies the joins
	// and the given order. paginate=false fetches every matching row so the caller
	// can merge multiple arms and slice the combined result in Go.
	fetchTasks := func(armCond builder.Cond, selectCols, armOrderby string, paginate bool) ([]*Task, error) {
		query := d.s.Where(armCond)
		if selectCols == distinct {
			query = query.Distinct(selectCols)
		} else {
			// Select() passes the raw column list through untouched while Distinct()
			// (no args) still emits the DISTINCT keyword.
			query = query.Select(selectCols).Distinct()
		}
		if paginate && limit > 0 {
			query = query.Limit(limit, start)
		}
		if rankFavoritesJoin {
			query = query.Join("LEFT", "favorites rank_favorites", "rank_favorites.entity_id = tasks.id AND rank_favorites.user_id = ? AND rank_favorites.kind = ?", d.a.GetID(), FavoriteKindTask)
		}
		query = addJoins(query)

		armTasks := []*Task{}
		if err := query.OrderBy(armOrderby).Find(&armTasks); err != nil {
			sql, vals := query.LastSQL()
			return nil, fmt.Errorf("could not fetch tasks, error was '%w', sql: '%v', values: %v", err, sql, vals)
		}
		return armTasks, nil
	}

	rankCondWith := func(searchCond builder.Cond) builder.Cond {
		c := builder.And(rankScopeCond, searchCond, filterCond)
		if expandSubtasks {
			c = builder.And(c, subtaskParentCond)
		}
		return c
	}

	switch {
	case rankByRelevance && searchIndex > 0:
		// A numeric search matches both the task index and the fuzzy text. pdb.score
		// can only score a pure-ParadeDB query, so a `||| ... OR index = N` group is
		// an unsupported query shape on ParadeDB. Run two supported arms instead and
		// rank exact index matches first, then text matches by relevance.
		indexTasks, err := fetchTasks(rankCondWith(builder.Eq{"tasks.`index`": searchIndex}), distinct, orderby, false)
		if err != nil {
			return nil, 0, err
		}

		textTasks, err := fetchTasks(rankCondWith(textSearchCond), distinct+", pdb.score(tasks.id)", "pdb.score(tasks.id) DESC, "+orderby, false)
		if err != nil {
			return nil, 0, err
		}

		// Exact index matches rank first; dedup a task matching both arms in favour
		// of its index-match position.
		seen := make(map[int64]bool, len(indexTasks)+len(textTasks))
		merged := make([]*Task, 0, len(indexTasks)+len(textTasks))
		for _, t := range indexTasks {
			if !seen[t.ID] {
				seen[t.ID] = true
				merged = append(merged, t)
			}
		}
		for _, t := range textTasks {
			if !seen[t.ID] {
				seen[t.ID] = true
				merged = append(merged, t)
			}
		}

		tasks = paginateInMemory(merged, limit, start)
	case rankByRelevance:
		// Pure text search: a single pdb.score-ordered query over the score-able
		// scope is a supported shape.
		tasks, err = fetchTasks(rankCondWith(textSearchCond), distinct+", pdb.score(tasks.id)", "pdb.score(tasks.id) DESC, "+orderby, true)
		if err != nil {
			return nil, 0, err
		}
	default:
		tasks, err = fetchTasks(cond, distinct, orderby, true)
		if err != nil {
			return nil, 0, err
		}
	}

	// fetch subtasks when expanding
	if expandSubtasks && len(tasks) > 0 {
		subtasks := []*Task{}

		taskIDs := []any{}
		for _, task := range tasks {
			taskIDs = append(taskIDs, task.ID)
		}

		var inPlaceholders = strings.Repeat("?,", len(taskIDs))
		inPlaceholders = inPlaceholders[:len(inPlaceholders)-1]

		var notIn = strings.Repeat("?,", len(taskIDs))
		notIn = notIn[:len(notIn)-1]

		allArgs := make([]any, 0, len(taskIDs)*2)
		allArgs = append(allArgs, taskIDs...)
		allArgs = append(allArgs, taskIDs...)

		err = d.s.SQL(`SELECT * FROM tasks WHERE id IN (WITH RECURSIVE sub_tasks AS (
		SELECT task_id,
			other_task_id,
			relation_kind,
			created_by_id,
			created
		FROM task_relations
		WHERE task_id IN (`+inPlaceholders+`)
		AND relation_kind = '`+string(RelationKindSubtask)+`'

		UNION ALL

		SELECT tr.task_id,
			tr.other_task_id,
			tr.relation_kind,
			tr.created_by_id,
			tr.created
		FROM task_relations tr
		INNER JOIN
		sub_tasks st ON tr.task_id = st.other_task_id
		WHERE tr.relation_kind = '`+string(RelationKindSubtask)+`')
		SELECT other_task_id
		FROM sub_tasks) AND id NOT IN (`+notIn+`)`, allArgs...).Find(&subtasks)
		if err != nil {
			return nil, totalCount, err
		}

		tasks = append(tasks, subtasks...)
	}

	queryCount := d.s.Where(cond)
	if joinTaskBuckets {
		joinCond := "task_buckets.task_id = tasks.id"
		if opts.projectViewID > 0 {
			joinCond += " AND task_buckets.project_view_id = ?"
			queryCount = queryCount.Join("LEFT", "task_buckets", joinCond, opts.projectViewID)
		} else {
			queryCount = queryCount.Join("LEFT", "task_buckets", joinCond)
		}
	}
	if expandSubtasks {
		queryCount = queryCount.
			Join("LEFT", "task_relations", "tasks.id = task_relations.task_id and task_relations.relation_kind = 'parenttask'").
			Join("LEFT", "tasks parent_tasks", "task_relations.other_task_id = parent_tasks.id")
	}
	totalCount, err = queryCount.
		Select("count(DISTINCT tasks.id)").
		Count(&Task{})
	if err != nil {
		sql, vals := queryCount.LastSQL()
		return nil, 0, fmt.Errorf("could not fetch task count, error was '%w', sql: '%v', values: %v", err, sql, vals)
	}
	return
}

// paginateInMemory slices an already-ordered result set. limit == 0 means "no
// limit" (return everything from start onwards), matching getLimitFromPageIndex.
func paginateInMemory(tasks []*Task, limit, start int) []*Task {
	if start >= len(tasks) {
		return []*Task{}
	}
	tasks = tasks[start:]
	if limit > 0 && limit < len(tasks) {
		tasks = tasks[:limit]
	}
	return tasks
}
