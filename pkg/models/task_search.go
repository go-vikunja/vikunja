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

func (sf *SubTableFilter) ToBaseSubQuery(taskAlias string) *builder.Builder {
	baseFilter := sf.BaseFilter
	if taskAlias != "tasks" {
		baseFilter = strings.ReplaceAll(baseFilter, "tasks.", taskAlias+".")
	}

	var cond = builder.
		Select("1").
		From(sf.Table).
		Where(builder.Expr(baseFilter))

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
	parts := make([]string, 0, len(opts.sortby))
	for _, param := range opts.sortby {
		// Validate the params
		if err := param.validate(); err != nil {
			return "", err
		}

		if param.sortBy == taskPropertyRelevance {
			// pdb.score is only valid SQL when the ParadeDB extension is installed.
			// Search strips the param when the query cannot be scored, this guards
			// any other caller. Most-relevant-first is the only useful direction,
			// the requested order is ignored.
			if db.ParadeDBAvailable() {
				parts = append(parts, "pdb.score(tasks.id) DESC")
			}
			continue
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

		part := prefix + "`" + param.sortBy + "` " + param.orderBy.String()

		// Mysql sorts columns with null values before ones without null value.
		// Because it does not have support for NULLS FIRST or NULLS LAST we work around this by
		// first sorting for null (or not null) values and then the order we actually want to.
		if db.Type() == schemas.MYSQL {
			part = prefix + "`" + param.sortBy + "` IS NULL, " + part
		}

		// Postgres and sqlite allow us to control how columns with null values are sorted.
		// To make that consistent with the sort order we have and other dbms, we're adding a separate clause here.
		if db.Type() == schemas.POSTGRES || db.Type() == schemas.SQLITE {
			part += " NULLS LAST"
		}

		parts = append(parts, part)
	}

	return strings.Join(parts, ", "), nil
}

func convertFiltersToDBFilterCond(rawFilters []*taskFilter, includeNulls bool) (filterCond builder.Cond, err error) {
	return convertFiltersToDBFilterCondWithAlias(rawFilters, includeNulls, "tasks")
}

// convertFiltersToDBFilterCondWithAlias builds the filter condition against the
// given task table alias. Passing "parent_tasks" lets the subtask-expansion root
// condition ask "does the parent satisfy the filter" (see #2646).
func convertFiltersToDBFilterCondWithAlias(rawFilters []*taskFilter, includeNulls bool, taskAlias string) (filterCond builder.Cond, err error) {

	var dbFilters = make([]builder.Cond, 0, len(rawFilters))
	// Track join types separately because after merging consecutive sub-table
	// filters, the indexes of dbFilters no longer correspond 1:1 with rawFilters.
	var dbFilterJoins = make([]taskFilterConcatinator, 0, len(rawFilters))

	for i := 0; i < len(rawFilters); i++ {
		f := rawFilters[i]

		if nested, is := f.value.([]*taskFilter); is {
			nestedDBFilters, err := convertFiltersToDBFilterCondWithAlias(nested, includeNulls, taskAlias)
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

			filterSubQuery := subTableFilterParams.ToBaseSubQuery(taskAlias).And(combinedInnerCond)

			var filter builder.Cond
			if f.comparator == taskFilterComparatorNotEquals || f.comparator == taskFilterComparatorNotIn {
				filter = builder.NotExists(filterSubQuery)
			} else {
				filter = builder.Exists(filterSubQuery)
			}

			if includeNulls && subTableFilterParams.AllowNullCheck {
				filter = builder.Or(filter, builder.NotExists(subTableFilterParams.ToBaseSubQuery(taskAlias)))
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
			f.field = taskAlias + ".`" + f.field + "`"
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

// cloneTaskFilters deep-copies the parsed filters so the parent-scoped filter
// build does not mutate the shared field names (convertFiltersToDBFilterCond
// rewrites f.field in place, which must not leak back into the main query).
func cloneTaskFilters(filters []*taskFilter) []*taskFilter {
	cloned := make([]*taskFilter, len(filters))
	for i, f := range filters {
		c := *f
		if nested, is := f.value.([]*taskFilter); is {
			c.value = cloneTaskFilters(nested)
		}
		cloned[i] = &c
	}
	return cloned
}

// stripBucketIDFilters returns a copy of filters with every bucket_id condition
// removed (recursing into nested groups and dropping groups left empty). The
// parent-scoped root condition cannot evaluate a bucket_id filter against the
// parent: convertFiltersToDBFilterCondWithAlias hard-codes the task_buckets.bucket_id
// column, and the only task_buckets join is keyed on the child (task_buckets.task_id
// = tasks.id). Keeping it would bind the parent filter to the child's bucket and
// misclassify roots, so a bucket_id filter simply does not constrain the parent.
func stripBucketIDFilters(filters []*taskFilter) []*taskFilter {
	stripped := make([]*taskFilter, 0, len(filters))
	for _, f := range filters {
		if nested, is := f.value.([]*taskFilter); is {
			child := stripBucketIDFilters(nested)
			if len(child) == 0 {
				continue
			}
			c := *f
			c.value = child
			stripped = append(stripped, &c)
			continue
		}
		if f.field == taskPropertyBucketID {
			continue
		}
		stripped = append(stripped, f)
	}
	return stripped
}

// buildSubtaskRootCondition decides which tasks count as "roots" when expanding
// subtasks: a task is a root unless its parent is itself part of this result set.
//
// A task is excluded from roots only when ALL of the following hold:
//   - it has a parenttask relation, AND
//   - the parent task exists, AND
//   - the parent is within the queried result scope, AND
//   - the parent satisfies the active filter.
func (d *dbTaskSearcher) buildSubtaskRootCondition(opts *taskSearchOptions) (builder.Cond, error) {
	// The base result set is (projectIDCond OR favoritesCond); mirror both so the
	// parent is considered "in scope" exactly when it could appear as a result row.
	scopes := make([]builder.Cond, 0, 2)
	if len(opts.projectIDs) > 0 {
		scopes = append(scopes, builder.In("parent_tasks.project_id", opts.projectIDs))
	}
	if d.hasFavoritesProject {
		favCond := builder.
			Select("entity_id").
			From("favorites").
			Where(builder.And(
				builder.Eq{"user_id": d.a.GetID()},
				builder.Eq{"kind": FavoriteKindTask},
			))
		scopes = append(scopes, builder.In("parent_tasks.id", favCond))
	}

	parentInScope := builder.Cond(builder.Expr("1 = 1"))
	if len(scopes) > 0 {
		parentInScope = builder.Or(scopes...)
	}

	parentMatchesFilter := builder.Cond(builder.Expr("1 = 1"))
	if len(opts.parsedFilters) > 0 {
		parentFilters := stripBucketIDFilters(cloneTaskFilters(opts.parsedFilters))
		filterCond, err := convertFiltersToDBFilterCondWithAlias(parentFilters, opts.filterIncludeNulls, "parent_tasks")
		if err != nil {
			return nil, err
		}
		if filterCond != nil {
			parentMatchesFilter = filterCond
		}
	}

	// A soft-deleted parent no longer counts, so its children become roots
	parentIsRoot := builder.And(
		builder.NotNull{"task_relations.id"},
		builder.NotNull{"parent_tasks.id"},
		taskNotDeletedCond("parent_tasks"),
		parentInScope,
		parentMatchesFilter,
	)

	return builder.Not{parentIsRoot}, nil
}

//nolint:gocyclo
func (d *dbTaskSearcher) Search(opts *taskSearchOptions) (tasks []*Task, totalCount int64, err error) {

	joinTaskBuckets := hasBucketIDInParsedFilter(opts.parsedFilters)

	var expandSubtasks = false
	for _, expandable := range opts.expand {
		if expandable == TaskCollectionExpandSubtasks {
			expandSubtasks = true
			break
		}
	}

	// The root condition asks whether a task's parent is part of this result set,
	// which means re-building the filter against the parent_tasks alias. Compute it
	// before convertFiltersToDBFilterCond mutates the shared filter field names.
	var subtaskRootCond builder.Cond
	if expandSubtasks {
		subtaskRootCond, err = d.buildSubtaskRootCondition(opts)
		if err != nil {
			return nil, 0, err
		}
	}

	filterCond, err := convertFiltersToDBFilterCond(opts.parsedFilters, opts.filterIncludeNulls)
	if err != nil {
		return nil, 0, err
	}

	// Then return all tasks for that projects
	var where builder.Cond

	searchIndex := getTaskIndexFromSearchString(opts.search)
	if opts.search != "" {
		// With the fuzzy cast the relevance score is a constant sum, not BM25: each
		// query word matching a field adds 1.0 (exact/prefix) or 0.5 (one edit away),
		// times the field's boost. Boosting titles 1.5x keeps "more matched words
		// wins": two description words (2.0) still beat a single title word (1.5),
		// the boost only decides between tasks matching the same number of words.
		where = db.MultiFieldSearchWithBoosts([]string{"title", "description"}, []float64{1.5, 1}, opts.search, "tasks")

		if searchIndex > 0 {
			where = builder.Or(where, builder.Eq{"tasks.`index`": searchIndex})
		}
	}

	relevanceSortRequested := false
	for _, param := range opts.sortby {
		if param.sortBy == taskPropertyRelevance {
			relevanceSortRequested = true
			break
		}
	}

	// ParadeDB exposes a relevance score via pdb.score(tasks.id) for a query
	// containing a ParadeDB operator (the ||| from MultiFieldSearchWithBoosts
	// above qualifies; the comment there describes how the score adds up). When
	// searching without an explicit user sort — or when the client explicitly
	// sorts by relevance — order by that score so tasks matching all query words
	// rank above tasks matching only some.
	//
	// Limited to pure-text searches: numeric searches add an `OR index = N` branch,
	// which pdb.score rejects as an unsupported query shape. pdb.score is also
	// invalid SQL on sqlite/mysql/plain postgres, hence the ParadeDBAvailable() gate.
	wantsRelevanceRanking := db.ParadeDBAvailable() &&
		opts.search != "" &&
		searchIndex == 0 &&
		(!opts.userProvidedSort || relevanceSortRequested)

	var projectIDCond builder.Cond
	var favoritesCond builder.Cond
	if len(opts.projectIDs) > 0 {
		projectIDCond = builder.In("tasks.project_id", opts.projectIDs)
	}

	if d.hasFavoritesProject {
		addFavoritesCond := true
		if wantsRelevanceRanking && len(opts.projectIDs) > 0 {
			// pdb.score also rejects the favorites arm (`OR tasks.id IN (<subquery>)`).
			// On an all-projects scope that arm is usually redundant — every favorited
			// task already lives in one of the user's projects — so drop it and keep
			// relevance ranking. Only favorites outside the scope (e.g. in projects the
			// user lost access to) need the arm and keep the default, unranked ordering.
			var hasOutOfScopeFavorites bool
			hasOutOfScopeFavorites, err = d.s.
				Table("favorites").
				Join("INNER", "tasks", "tasks.id = favorites.entity_id").
				Where(builder.And(
					builder.Eq{"favorites.user_id": d.a.GetID()},
					builder.Eq{"favorites.kind": FavoriteKindTask},
					builder.NotIn("tasks.project_id", opts.projectIDs),
					taskNotDeletedCond("tasks"),
				)).
				Exist()
			if err != nil {
				return nil, 0, err
			}
			addFavoritesCond = hasOutOfScopeFavorites
		}

		if addFavoritesCond {
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
	}

	limit, start := getLimitFromPageIndex(opts.page, opts.perPage)
	cond := builder.And(builder.Or(projectIDCond, favoritesCond), where, filterCond)

	// When the favorites arm is still part of the query (Favorites view, or
	// out-of-scope favorites exist), its shape is unsupported — stay unranked.
	rankByRelevance := wantsRelevanceRanking && favoritesCond == nil

	if rankByRelevance && !relevanceSortRequested {
		opts.sortby = append([]*sortParam{{sortBy: taskPropertyRelevance, orderBy: orderDescending}}, opts.sortby...)
	}
	if !rankByRelevance && relevanceSortRequested {
		kept := make([]*sortParam, 0, len(opts.sortby))
		for _, param := range opts.sortby {
			if param.sortBy != taskPropertyRelevance {
				kept = append(kept, param)
			}
		}
		opts.sortby = kept
	}

	orderby, err := getOrderByDBStatement(opts)
	if err != nil {
		return nil, 0, err
	}

	var distinct = "tasks.*"
	if strings.Contains(orderby, "task_positions.") {
		distinct += ", task_positions.position"
	}

	if expandSubtasks {
		cond = builder.And(cond, subtaskRootCond)
	}

	query := d.s.Where(cond)
	if rankByRelevance {
		// Select() passes the raw column list through untouched while Distinct()
		// (no args) still emits DISTINCT. Distinct("tasks.*, pdb.score(tasks.id)")
		// would quote-corrupt the function call into "pdb"."score(tasks"."id)".
		query = query.Select(distinct + ", pdb.score(tasks.id)").Distinct()
	} else {
		query = query.Distinct(distinct)
	}
	if limit > 0 {
		query = query.Limit(limit, start)
	}

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

	tasks = []*Task{}
	err = query.
		OrderBy(orderby).
		Find(&tasks)
	if err != nil {
		sql, vals := query.LastSQL()
		return nil, 0, fmt.Errorf("could not fetch tasks, error was '%w', sql: '%v', values: %v", err, sql, vals)
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
		FROM sub_tasks) AND id NOT IN (`+notIn+`) AND deleted_at IS NULL`, allArgs...).Find(&subtasks)
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
