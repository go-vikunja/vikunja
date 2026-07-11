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
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"

	"github.com/ganigeorgiev/fexpr"
	"github.com/jszwedko/go-datemath"
	"xorm.io/builder"
)

// entriesForProjectCond matches time entries belonging to a project given a
// predicate over a project_id column: standalone entries whose own project_id
// matches, plus task-attached entries whose task currently lives in a matching
// project. Tasks move between projects, so the project is resolved via the task
// at query time rather than denormalized. Used for both permission scoping and
// the project_id filter.
func entriesForProjectCond(projectIDCond builder.Cond) builder.Cond {
	return builder.Or(
		projectIDCond,
		builder.In("task_id",
			builder.Select("id").From("tasks").Where(builder.And(projectIDCond, taskNotDeletedCond("tasks"))),
		),
	)
}

// timeEntryFilterCond parses a task-style filter string into a condition over
// the time_entries table, or nil for an empty filter. Filterable fields:
// user_id, task_id, project_id (ints / in-lists), start_time, end_time (dates,
// datemath, or the literal null for running timers). comment is deliberately
// not filterable — text matching belongs to search.
func timeEntryFilterCond(filter, filterTimezone string) (builder.Cond, error) {
	if filter == "" {
		return nil, nil
	}

	parsed, err := fexpr.Parse(preprocessFilterString(filter))
	if err != nil {
		return nil, &ErrInvalidFilterExpression{Expression: filter, ExpressionError: err}
	}

	loc := config.GetTimeZone()
	if filterTimezone != "" {
		loc, err = time.LoadLocation(filterTimezone)
		if err != nil {
			return nil, &ErrInvalidTimezone{Name: filterTimezone, LoadError: err}
		}
	}

	return buildTimeEntryFilterCond(parsed, loc)
}

func buildTimeEntryFilterCond(groups []fexpr.ExprGroup, loc *time.Location) (builder.Cond, error) {
	conds := make([]builder.Cond, 0, len(groups))
	joins := make([]taskFilterConcatinator, 0, len(groups))

	for _, g := range groups {
		join := filterConcatAnd
		if g.Join == fexpr.JoinOr {
			join = filterConcatOr
		}

		var (
			cond builder.Cond
			err  error
		)
		switch item := g.Item.(type) {
		case []fexpr.ExprGroup: // a parenthesized sub-expression
			cond, err = buildTimeEntryFilterCond(item, loc)
		case fexpr.Expr:
			var comparator taskFilterComparator
			comparator, err = getFilterComparatorFromOp(item.Op)
			if err == nil {
				cond, err = resolveTimeEntryFilter(item.Left.Literal, comparator, item.Right.Literal, loc)
			}
		}
		if err != nil {
			return nil, err
		}
		conds = append(conds, cond)
		joins = append(joins, join)
	}

	if len(conds) == 0 {
		return nil, nil
	}
	result := conds[0]
	for i := 1; i < len(conds); i++ {
		if joins[i] == filterConcatOr {
			result = builder.Or(result, conds[i])
			continue
		}
		result = builder.And(result, conds[i])
	}
	return result, nil
}

func resolveTimeEntryFilter(field string, comparator taskFilterComparator, raw string, loc *time.Location) (builder.Cond, error) {
	switch field {
	case "user_id", "task_id":
		value, err := timeEntryIntFilterValue(raw, comparator)
		if err != nil {
			return nil, ErrInvalidTimeEntryFilterValue{Field: field, Value: raw}
		}
		return getFilterCond(&taskFilter{field: field, value: value, comparator: comparator, isNumeric: true}, false)

	case "project", "project_id":
		value, err := timeEntryIntFilterValue(raw, comparator)
		if err != nil {
			return nil, ErrInvalidTimeEntryFilterValue{Field: "project_id", Value: raw}
		}
		// Build membership positively (standalone-in-project OR task-in-project)
		// and negate the whole set for != / not in. Negating project_id alone would
		// wrongly match task-attached entries, whose own project_id is 0.
		positive, negate := comparator, false
		if comparator == taskFilterComparatorNotEquals {
			positive, negate = taskFilterComparatorEquals, true
		}
		if comparator == taskFilterComparatorNotIn {
			positive, negate = taskFilterComparatorIn, true
		}
		inner, err := getFilterCond(&taskFilter{field: "project_id", value: value, comparator: positive, isNumeric: true}, false)
		if err != nil {
			return nil, err
		}
		cond := entriesForProjectCond(inner)
		if negate {
			cond = builder.Not{cond}
		}
		return cond, nil

	case "start_time", "end_time":
		if raw == "null" {
			return nullTimeFilterCond(field, comparator)
		}
		value, err := timeEntryTimeFilterValue(raw, loc)
		if err != nil {
			return nil, ErrInvalidTimeEntryFilterValue{Field: field, Value: raw}
		}
		return getFilterCond(&taskFilter{field: field, value: value, comparator: comparator}, false)

	default:
		return nil, ErrInvalidTimeEntryFilterField{Field: field}
	}
}

// nullTimeFilterCond handles `end_time = null` (running timers) and its negation.
func nullTimeFilterCond(field string, comparator taskFilterComparator) (builder.Cond, error) {
	if comparator == taskFilterComparatorEquals {
		return &builder.IsNull{field}, nil
	}
	if comparator == taskFilterComparatorNotEquals {
		return &builder.NotNull{field}, nil
	}
	return nil, ErrInvalidTimeEntryFilterValue{Field: field, Value: "null"}
}

func timeEntryIntFilterValue(raw string, comparator taskFilterComparator) (any, error) {
	if comparator == taskFilterComparatorIn || comparator == taskFilterComparatorNotIn {
		parts := strings.Split(raw, ",")
		values := make([]int64, 0, len(parts))
		for _, part := range parts {
			v, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
			if err != nil {
				return nil, err
			}
			values = append(values, v)
		}
		return values, nil
	}
	return strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
}

// timeEntryTimeFilterValue mirrors the task filter's date handling: datemath
// (now, now-7d) first, then explicit date formats.
func timeEntryTimeFilterValue(raw string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = config.GetTimeZone()
	}
	if expr, err := safeDatemathParse(raw); err == nil {
		t := expr.Time(datemath.WithLocation(loc)).In(config.GetTimeZone())
		return adjustDateForMysql(t), nil
	}
	return parseTimeFromUserInput(raw, loc)
}
