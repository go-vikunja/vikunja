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

package services

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"
	"dario.cat/mergo"
	"github.com/ganigeorgiev/fexpr"
	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/copier"
	"github.com/jszwedko/go-datemath"
	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// Task Read All related types and constants
// These are moved from models package to support service-layer implementation

type (
	sortParam struct {
		sortBy        string
		orderBy       sortOrder // asc or desc
		projectViewID int64
	}

	sortOrder string

	taskSearchOptions struct {
		search             string
		page               int
		perPage            int
		sortby             []*sortParam
		parsedFilters      []*taskFilter
		filterIncludeNulls bool
		filter             string
		filterTimezone     string
		isSavedFilter      bool
		projectIDs         []int64
		expand             []models.TaskCollectionExpandable
		projectViewID      int64
	}

	taskFilter struct {
		field        string
		value        interface{}
		comparator   taskFilterComparator
		concatenator taskFilterConcatinator
		isNumeric    bool
	}

	taskFilterComparator   string
	taskFilterConcatinator string
)

const (
	// Sort order constants
	orderInvalid    sortOrder = "invalid"
	orderAscending  sortOrder = "asc"
	orderDescending sortOrder = "desc"

	// Task property constants for sorting and filtering
	taskPropertyID            string = "id"
	taskPropertyTitle         string = "title"
	taskPropertyDescription   string = "description"
	taskPropertyDone          string = "done"
	taskPropertyDoneAt        string = "done_at"
	taskPropertyDueDate       string = "due_date"
	taskPropertyCreatedByID   string = "created_by_id"
	taskPropertyProjectID     string = "project_id"
	taskPropertyRepeatAfter   string = "repeat_after"
	taskPropertyPriority      string = "priority"
	taskPropertyStartDate     string = "start_date"
	taskPropertyEndDate       string = "end_date"
	taskPropertyHexColor      string = "hex_color"
	taskPropertyPercentDone   string = "percent_done"
	taskPropertyUID           string = "uid"
	taskPropertyCreated       string = "created"
	taskPropertyUpdated       string = "updated"
	taskPropertyPosition      string = "position"
	taskPropertyBucketID      string = "bucket_id"
	taskPropertyIndex         string = "index"
	taskPropertyProjectViewID string = "project_view_id"

	// Task filter comparators
	taskFilterComparatorEquals        taskFilterComparator = "="
	taskFilterComparatorNotEquals     taskFilterComparator = "!="
	taskFilterComparatorGreater       taskFilterComparator = ">"
	taskFilterComparatorGreaterEquals taskFilterComparator = ">="
	taskFilterComparatorLess          taskFilterComparator = "<"
	taskFilterComparatorLessEquals    taskFilterComparator = "<="
	taskFilterComparatorLike          taskFilterComparator = "like"
	taskFilterComparatorIn            taskFilterComparator = "in"
	taskFilterComparatorNotIn         taskFilterComparator = "not_in"

	// Task filter concatenators
	taskFilterConcatAnd taskFilterConcatinator = "and"
	taskFilterConcatOr  taskFilterConcatinator = "or"

	taskFilterComparatorInvalid taskFilterComparator = "invalid"
)

// subTableFilter defines how to query related tables (labels, assignees, etc.) via EXISTS subqueries
type subTableFilter struct {
	Table           string // Related table name (e.g., "label_tasks")
	BaseFilter      string // Join condition (e.g., "tasks.id = task_id")
	FilterableField string // Column to filter on (e.g., "label_id")
	AllowNullCheck  bool   // Whether to support filterIncludeNulls
}

// subTableFilters defines configurations for all subtable relationships.
//
// CRITICAL: AllowNullCheck Configuration (T019 Bug Fix)
//
// The AllowNullCheck field controls whether a subtable filter respects the
// FilterIncludeNulls parameter when building filter conditions. This is critical
// for correct filter semantics.
//
// When AllowNullCheck = false (correct for most subtable filters):
//   - "labels = 4" returns ONLY tasks with label 4
//   - Even if FilterIncludeNulls: true, tasks without ANY labels are NOT included
//   - This prevents the "OR NOT EXISTS" clause that caused the T019 bug
//
// When AllowNullCheck = true (use with caution):
//   - "field = X" with FilterIncludeNulls: true will add "OR NOT EXISTS (...)"
//   - This returns tasks with field=X OR tasks without any entries in the subtable
//   - Rarely the desired behavior for relationship filters
//
// T019 Bug (Fixed 2025-10-25):
//   - Symptom: Saved filter "labels = 4" returned tasks WITH label 4 OR WITHOUT any labels
//   - Root Cause: AllowNullCheck: true caused FilterIncludeNulls: true (frontend default)
//     to add "OR NOT EXISTS (SELECT ... FROM label_tasks WHERE task_id = tasks.id)"
//   - Fix: Set AllowNullCheck: false for subtable filters (labels, assignees, reminders)
//   - Validation: TestTaskService_SavedFilter_WithFilterIncludeNulls_True_Integration passes
//
// Filter Semantics Examples (with AllowNullCheck: false):
//   - "labels = 4" → Returns tasks with label 4 only
//   - "labels != 4" with FilterIncludeNulls: true → Returns tasks without label 4 (includes unlabeled)
//   - "labels in [4, 5]" → Returns tasks with label 4 OR label 5 only
//   - "assignees = 1" → Returns tasks assigned to user 1 only
//
// See also: specs/007-fix-saved-filters/T019-DEBUGGING.md for full investigation details
var subTableFilters = map[string]subTableFilter{
	"labels": {
		Table:           "label_tasks",
		BaseFilter:      "tasks.id = task_id",
		FilterableField: "label_id",
		AllowNullCheck:  false, // T019 FIX: "labels = X" should NOT include tasks without labels
	},
	"label_id": {
		Table:           "label_tasks",
		BaseFilter:      "tasks.id = task_id",
		FilterableField: "label_id",
		AllowNullCheck:  false, // T019 FIX: "label_id = X" should NOT include tasks without labels
	},
	"reminders": {
		Table:           "task_reminders",
		BaseFilter:      "tasks.id = task_id",
		FilterableField: "reminder",
		AllowNullCheck:  false, // T019 FIX: "reminders > X" should NOT include tasks without reminders
	},
	"assignees": {
		Table:           "task_assignees",
		BaseFilter:      "tasks.id = task_id",
		FilterableField: "username",
		AllowNullCheck:  false, // T019 FIX: "assignees = X" should NOT include unassigned tasks
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

// strictComparators defines comparators that should be converted to IN/NOT IN for subtable queries
var strictComparators = map[taskFilterComparator]bool{
	taskFilterComparatorIn:        true,
	taskFilterComparatorNotIn:     true,
	taskFilterComparatorEquals:    true,
	taskFilterComparatorNotEquals: true,
}

// toBaseSubQuery creates the base subquery for EXISTS checks on related tables
func (sf *subTableFilter) toBaseSubQuery() *builder.Builder {
	var cond = builder.
		Select("1").
		From(sf.Table).
		Where(builder.Expr(sf.BaseFilter))

	// Special case: assignees filter needs to join users table
	if sf.Table == "task_assignees" {
		cond.Join("INNER", "users", "users.id = user_id")
	}

	return cond
}

// String returns the string representation of a sort order
func (o sortOrder) String() string {
	return string(o)
}

// validate validates a sort parameter
func (sp *sortParam) validate() error {
	switch sp.sortBy {
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
		taskPropertyBucketID,
		taskPropertyIndex,
		taskPropertyProjectViewID:
		// Valid sort parameter
	default:
		return models.ErrInvalidTaskField{
			TaskField: sp.sortBy,
		}
	}

	if sp.orderBy != orderAscending && sp.orderBy != orderDescending {
		return models.ErrInvalidSortOrder{
			OrderBy: models.SortOrder(sp.orderBy),
		}
	}

	return nil
}

// getSortOrderFromString converts a string to sortOrder
func getSortOrderFromString(s string) sortOrder {
	// Normalize the input: trim whitespace and convert to lowercase
	normalized := strings.ToLower(strings.TrimSpace(s))

	switch normalized {
	case "asc", "ascending":
		return orderAscending
	case "desc", "descending":
		return orderDescending
	default:
		// For invalid or empty values, default to ascending for better UX
		// This prevents 500 errors when frontend sends malformed parameters
		return orderAscending
	}
}

// Filter parsing constants
const (
	safariDateAndTime = "2006-01-02 15:04"
	safariDate        = "2006-01-02"

	taskPropertyAssignees = "assignees"
	taskPropertyLabels    = "labels"
	taskPropertyReminders = "reminders"
)

// parseTimeFromUserInput parses various time formats from user input
func (ts *TaskService) parseTimeFromUserInput(timeString string, loc *time.Location) (value time.Time, err error) {
	value, err = time.ParseInLocation(time.RFC3339, timeString, loc)
	if err != nil {
		value, err = time.ParseInLocation(safariDateAndTime, timeString, loc)
	}
	if err != nil {
		value, err = time.ParseInLocation(safariDate, timeString, loc)
	}
	if err != nil {
		// Here we assume a date like 2022-11-1 and try to parse it manually
		parts := strings.Split(timeString, "-")
		if len(parts) < 3 {
			return
		}
		year, err := strconv.Atoi(parts[0])
		if err != nil {
			return value, err
		}
		month, err := strconv.Atoi(parts[1])
		if err != nil {
			return value, err
		}
		day, err := strconv.Atoi(parts[2])
		if err != nil {
			return value, err
		}
		value = time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
		return value.In(config.GetTimeZone()), nil
	}
	return value.In(config.GetTimeZone()), err
}

// validateTaskFieldComparator validates a task filter comparator
func (ts *TaskService) validateTaskFieldComparator(comparator taskFilterComparator) error {
	switch comparator {
	case
		taskFilterComparatorEquals,
		taskFilterComparatorNotEquals,
		taskFilterComparatorGreater,
		taskFilterComparatorGreaterEquals,
		taskFilterComparatorLess,
		taskFilterComparatorLessEquals,
		taskFilterComparatorLike,
		taskFilterComparatorIn,
		taskFilterComparatorNotIn:
		return nil
	default:
		// Since models.taskFilterComparator is not exported, we return a generic error
		return fmt.Errorf("invalid task filter comparator: %s", comparator)
	}
}

// getFilterComparatorFromOp converts fexpr operator to taskFilterComparator
func (ts *TaskService) getFilterComparatorFromOp(op fexpr.SignOp) (taskFilterComparator, error) {
	switch op {
	case fexpr.SignEq:
		return taskFilterComparatorEquals, nil
	case fexpr.SignGt:
		return taskFilterComparatorGreater, nil
	case fexpr.SignGte:
		return taskFilterComparatorGreaterEquals, nil
	case fexpr.SignLt:
		return taskFilterComparatorLess, nil
	case fexpr.SignLte:
		return taskFilterComparatorLessEquals, nil
	case fexpr.SignNeq:
		return taskFilterComparatorNotEquals, nil
	case fexpr.SignLike:
		return taskFilterComparatorLike, nil
	case fexpr.SignAnyEq:
		fallthrough
	case "in":
		return taskFilterComparatorIn, nil
	case fexpr.SignAnyNeq:
		fallthrough
	case "not in":
		return taskFilterComparatorNotIn, nil
	default:
		return taskFilterComparatorEquals, fmt.Errorf("invalid task filter comparator operator: %s", op)
	}
}

// validateTaskField validates that a field name is valid for task filtering
func (ts *TaskService) validateTaskField(fieldName string) error {
	switch fieldName {
	case
		taskPropertyAssignees,
		taskPropertyLabels,
		taskPropertyReminders:
		return nil
	}
	return ts.validateTaskFieldForSorting(fieldName)
}

// validateTaskFieldForSorting validates that a field name is valid for task sorting
func (ts *TaskService) validateTaskFieldForSorting(fieldName string) error {
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
		taskPropertyBucketID,
		taskPropertyIndex,
		taskPropertyProjectViewID,
		"project",           // Alias for project_id
		"parent_project",    // Special field
		"parent_project_id": // Special field
		return nil
	}
	return models.ErrInvalidTaskField{TaskField: fieldName}
}

// getValueForField converts a string value to the appropriate type for a reflect field
func (ts *TaskService) getValueForField(field reflect.StructField, rawValue string, loc *time.Location) (value interface{}, err error) {
	if loc == nil {
		loc = config.GetTimeZone()
	}

	rawValue = strings.TrimSpace(rawValue)

	switch field.Type.Kind() {
	case reflect.Int64:
		value, err = strconv.ParseInt(rawValue, 10, 64)
	case reflect.Float64:
		value, err = strconv.ParseFloat(rawValue, 64)
	case reflect.String:
		value = rawValue
	case reflect.Bool:
		value, err = strconv.ParseBool(rawValue)
	case reflect.Struct:
		if field.Type == schemas.TimeType {
			var t datemath.Expression
			var tt time.Time
			t, err = datemath.Parse(rawValue)
			if err == nil {
				tt = t.Time(datemath.WithLocation(loc)).In(config.GetTimeZone())
			} else {
				tt, err = ts.parseTimeFromUserInput(rawValue, loc)
			}
			if err != nil {
				return
			}
			// Mysql/Mariadb does not support date values where the year < 1. To make this edge-case work,
			// we're setting the year to 1 in that case.
			if db.GetDialect() == builder.MYSQL && tt.Year() < 1 {
				tt = tt.AddDate(1-tt.Year(), 0, 0)
			}
			value = tt
		}
	case reflect.Slice:
		// If this is a slice of pointers we're dealing with some property which is a relation
		// In that case we don't really care about what the actual type is, we just cast the value to an
		// int64 since we need the id - yes, this assumes we only ever have int64 IDs, but this is fine.
		if field.Type.Elem().Kind() == reflect.Ptr {
			value, err = strconv.ParseInt(strings.TrimSpace(rawValue), 10, 64)
			return
		}

		// There are probably better ways to do this - please let me know if you have one.
		if field.Type.Elem().String() == "time.Time" {
			value, err = time.Parse(time.RFC3339, rawValue)
			value = value.(time.Time).In(config.GetTimeZone())
			return
		}
		fallthrough
	default:
		panic(fmt.Errorf("unrecognized filter type %s for field %s, value %s", field.Type.String(), field.Name, value))
	}

	return
}

// getNativeValueForTaskField gets the native value for a task field based on its type
func (ts *TaskService) getNativeValueForTaskField(fieldName string, comparator taskFilterComparator, value string, loc *time.Location) (reflectField *reflect.StructField, nativeValue interface{}, err error) {
	realFieldName := strings.ReplaceAll(strcase.ToCamel(fieldName), "Id", "ID")

	if realFieldName == "Assignees" {
		vals := strings.Split(value, ",")
		valueSlice := append([]string{}, vals...)
		return nil, valueSlice, nil
	}

	field, ok := reflect.TypeOf(&models.Task{}).Elem().FieldByName(realFieldName)
	if !ok {
		return nil, nil, models.ErrInvalidTaskField{TaskField: fieldName}
	}

	if realFieldName == "Reminders" {
		field, ok = reflect.TypeOf(&models.TaskReminder{}).Elem().FieldByName("Reminder")
		if !ok {
			return nil, nil, models.ErrInvalidTaskField{TaskField: fieldName}
		}
	}

	if comparator == taskFilterComparatorIn || comparator == taskFilterComparatorNotIn {
		vals := strings.Split(value, ",")
		valueSlice := []interface{}{}
		for _, val := range vals {
			v, err := ts.getValueForField(field, val, loc)
			if err != nil {
				return nil, nil, err
			}
			valueSlice = append(valueSlice, v)
		}
		return nil, valueSlice, nil
	}

	val, err := ts.getValueForField(field, value, loc)
	return &field, val, err
}

// parseFilterFromExpression parses a single filter expression
func (ts *TaskService) parseFilterFromExpression(f fexpr.ExprGroup, loc *time.Location) (filter *taskFilter, err error) {
	filter = &taskFilter{
		concatenator: taskFilterConcatAnd,
	}
	if f.Join == fexpr.JoinOr {
		filter.concatenator = taskFilterConcatOr
	}

	var value string
	switch v := f.Item.(type) {
	case fexpr.Expr:
		filter.field = v.Left.Literal
		value = v.Right.Literal
		filter.comparator, err = ts.getFilterComparatorFromOp(v.Op)
		if err != nil {
			return
		}
	case []fexpr.ExprGroup:
		values := make([]*taskFilter, 0, len(v))
		for _, expression := range v {
			subfilter, err := ts.parseFilterFromExpression(expression, loc)
			if err != nil {
				return nil, err
			}
			values = append(values, subfilter)
		}
		filter.value = values
		return
	}

	err = ts.validateTaskFieldComparator(filter.comparator)
	if err != nil {
		return
	}

	// Cast the field value to its native type
	var reflectValue *reflect.StructField
	if filter.field == "project" {
		filter.field = "project_id"
	}

	err = ts.validateTaskField(filter.field)
	if err != nil {
		return nil, err
	}

	reflectValue, filter.value, err = ts.getNativeValueForTaskField(filter.field, filter.comparator, value, loc)
	if err != nil {
		return nil, models.ErrInvalidTaskFilterValue{
			Field: filter.field,
			Value: value,
		}
	}
	if reflectValue != nil {
		filter.isNumeric = reflectValue.Type.Kind() == reflect.Int64
	}

	return filter, nil
}

// getTaskFiltersFromFilterString parses a filter string into a list of filter objects
func (ts *TaskService) getTaskFiltersFromFilterString(filter string, filterTimezone string) (filters []*taskFilter, err error) {
	if filter == "" {
		return
	}

	filter = strings.ReplaceAll(filter, " not in ", " "+string(fexpr.SignAnyNeq)+" ")
	filter = strings.ReplaceAll(filter, " in ", " ?= ")
	filter = strings.ReplaceAll(filter, " like ", " ~ ")

	// Regex pattern to match filter expressions
	re := regexp.MustCompile(`(\w+)\s*(>=|<=|!=|~|\?=|\?!=|=|>|<)\s*([^&|()]+)`)

	filter = re.ReplaceAllStringFunc(filter, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}

		field := parts[1]
		comparator := parts[2]
		value := strings.TrimSpace(parts[3])

		// Check if the value is already quoted
		if (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) ||
			(strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) {
			return field + " " + comparator + " " + value
		}

		// Quote the value
		quotedValue := "'" + strings.ReplaceAll(value, "'", "\\'") + "'"
		return field + " " + comparator + " " + quotedValue
	})

	parsedFilter, err := fexpr.Parse(filter)
	if err != nil {
		return nil, &models.ErrInvalidFilterExpression{
			Expression:      filter,
			ExpressionError: err,
		}
	}

	var loc *time.Location
	if filterTimezone != "" {
		loc, err = time.LoadLocation(filterTimezone)
		if err != nil {
			return nil, &models.ErrInvalidTimezone{
				Name:      filterTimezone,
				LoadError: err,
			}
		}
	}

	filters = make([]*taskFilter, 0, len(parsedFilter))
	for _, f := range parsedFilter {
		parsedFilter, err := ts.parseFilterFromExpression(f, loc)
		if err != nil {
			return nil, err
		}
		filters = append(filters, parsedFilter)
	}

	return
}

// getTaskFilterOptsFromCollection converts a TaskCollection to taskSearchOptions
func (ts *TaskService) getTaskFilterOptsFromCollection(tf *models.TaskCollection, projectView *models.ProjectView) (opts *taskSearchOptions, err error) {
	var finalSortBy []string
	var finalOrderBy []string

	if len(tf.SortByArr) > 0 {
		finalSortBy = tf.SortByArr
		finalOrderBy = tf.OrderByArr
	} else if len(tf.SortBy) > 0 {
		finalSortBy = tf.SortBy
		finalOrderBy = tf.OrderBy
	}

	tf.SortBy = finalSortBy
	tf.OrderBy = finalOrderBy

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

		if s == taskPropertyPosition && projectView != nil && projectView.ID < 0 {
			continue
		}

		if s == taskPropertyPosition {
			if projectView != nil {
				param.projectViewID = projectView.ID
			} else if tf.ProjectViewID != 0 {
				param.projectViewID = tf.ProjectViewID
			} else {
				return nil, fmt.Errorf("You must provide a project view ID when sorting by position")
			}
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

	if projectView != nil {
		opts.projectViewID = projectView.ID
	} else if tf.ProjectViewID != 0 {
		opts.projectViewID = tf.ProjectViewID
	}

	// Parse filter string if provided
	if tf.Filter != "" {
		opts.parsedFilters, err = ts.getTaskFiltersFromFilterString(tf.Filter, tf.FilterTimezone)
		if err != nil {
			return nil, err
		}
	}

	return opts, nil
}

// getFilterCond builds a database condition for a single filter
func (ts *TaskService) getFilterCond(f *taskFilter, includeNulls bool) (cond builder.Cond, err error) {
	field := f.field

	switch f.comparator {
	case taskFilterComparatorEquals:
		cond = &builder.Eq{field: f.value}
	case taskFilterComparatorNotEquals:
		cond = &builder.Neq{field: f.value}
	case taskFilterComparatorGreater:
		cond = &builder.Gt{field: f.value}
	case taskFilterComparatorGreaterEquals:
		cond = &builder.Gte{field: f.value}
	case taskFilterComparatorLess:
		cond = &builder.Lt{field: f.value}
	case taskFilterComparatorLessEquals:
		cond = &builder.Lte{field: f.value}
	case taskFilterComparatorLike:
		val, is := f.value.(string)
		if !is {
			return nil, fmt.Errorf("building LIKE filter for field '%s': %w", field, &models.ErrInvalidTaskFilterValue{Field: field, Value: f.value})
		}
		cond = &builder.Like{field, "%" + val + "%"}
	case taskFilterComparatorIn:
		cond = builder.In(field, f.value)
	case taskFilterComparatorNotIn:
		cond = builder.NotIn(field, f.value)
	}

	if includeNulls {
		cond = builder.Or(cond, &builder.IsNull{field})
		if f.isNumeric {
			cond = builder.Or(cond, &builder.IsNull{field}, &builder.Eq{field: 0})
		}
	}

	return
}

// buildSubtableFilterCondition builds a database condition for a subtable filter (labels, assignees, etc.)
// Subtable filters use EXISTS subqueries to handle many-to-many relationships without duplicating results.
func (ts *TaskService) buildSubtableFilterCondition(f *taskFilter, params subTableFilter, includeNulls bool) (builder.Cond, error) {
	// Skip assignees with LIKE operator (not supported)
	if f.field == "assignees" && f.comparator == taskFilterComparatorLike {
		return nil, nil // Signal to skip this filter
	}

	// Convert strict comparators (=, !=, in, not in) to IN for subtable queries
	comparator := f.comparator
	_, isStrict := strictComparators[f.comparator]
	if isStrict {
		comparator = taskFilterComparatorIn

		// For IN operator, the value must be a slice
		// If we're converting from = or !=, wrap the single value in a slice
		// Special handling for assignees: models layer returns []string, convert to []interface{}
		if f.comparator == taskFilterComparatorEquals || f.comparator == taskFilterComparatorNotEquals {
			switch v := f.value.(type) {
			case []string:
				// Assignees filter: models layer already parsed it as []string
				// Convert to []interface{} for XORM builder.In()
				valueSlice := make([]interface{}, len(v))
				for i, str := range v {
					valueSlice[i] = str
				}
				f.value = valueSlice
			case []interface{}:
				// Already a slice, keep as-is (from IN operator parsing)
				// No action needed
			default:
				// Single value, wrap in slice
				f.value = []interface{}{f.value}
			}
		}
	}

	// Build filter condition for the subtable field
	filter, err := ts.getFilterCond(&taskFilter{
		field:      params.FilterableField,
		value:      f.value,
		comparator: comparator,
		isNumeric:  f.isNumeric,
	}, false) // includeNulls=false for subquery
	if err != nil {
		return nil, fmt.Errorf("building subtable filter condition for field '%s': %w", f.field, err)
	}

	// Create EXISTS subquery
	filterSubQuery := params.toBaseSubQuery().And(filter)

	// Use NOT EXISTS for negation operators
	if f.comparator == taskFilterComparatorNotEquals || f.comparator == taskFilterComparatorNotIn {
		filter = builder.NotExists(filterSubQuery)
	} else {
		filter = builder.Exists(filterSubQuery)
	}

	// Add NULL check if requested (tasks with no entries in subtable)
	if includeNulls && params.AllowNullCheck {
		filter = builder.Or(filter, builder.NotExists(params.toBaseSubQuery()))
	}

	return filter, nil
}

// buildRegularFilterCondition builds a database condition for a regular task field (not a subtable)
func (ts *TaskService) buildRegularFilterCondition(f *taskFilter, includeNulls bool) (builder.Cond, error) {
	// Prefix with table name
	if f.field == taskPropertyBucketID {
		f.field = "task_buckets.`bucket_id`"
	} else {
		f.field = "tasks.`" + f.field + "`"
	}

	filter, err := ts.getFilterCond(f, includeNulls)
	if err != nil {
		return nil, fmt.Errorf("building regular filter condition for field '%s': %w", f.field, err)
	}
	return filter, nil
}

// convertFiltersToDBFilterCond converts parsed filter expressions into database query conditions
func (ts *TaskService) convertFiltersToDBFilterCond(rawFilters []*taskFilter, includeNulls bool) (filterCond builder.Cond, err error) {
	var dbFilters = make([]builder.Cond, 0, len(rawFilters))

	// Process each filter
	for _, f := range rawFilters {
		// Handle nested filters (from parentheses in expression)
		if nested, is := f.value.([]*taskFilter); is {
			nestedDBFilters, err := ts.convertFiltersToDBFilterCond(nested, includeNulls)
			if err != nil {
				return nil, fmt.Errorf("converting nested filters: %w", err)
			}
			dbFilters = append(dbFilters, nestedDBFilters)
			continue
		}

		var filter builder.Cond

		// Check if this is a subtable filter (labels, assignees, reminders, etc.)
		if subTableFilterParams, ok := subTableFilters[f.field]; ok {
			filter, err = ts.buildSubtableFilterCondition(f, subTableFilterParams, includeNulls)
			if err != nil {
				return nil, fmt.Errorf("processing subtable filter for field '%s': %w", f.field, err)
			}
			// Skip filter if buildSubtableFilterCondition returned nil (e.g., unsupported assignees LIKE)
			if filter == nil {
				continue
			}
		} else {
			// Regular field filter
			filter, err = ts.buildRegularFilterCondition(f, includeNulls)
			if err != nil {
				return nil, fmt.Errorf("processing regular filter for field '%s': %w", f.field, err)
			}
		}

		dbFilters = append(dbFilters, filter)
	}

	// Combine filters based on their concatenator (AND/OR)
	filterCond = ts.combineFilterConditions(dbFilters, rawFilters)

	return filterCond, nil
}

// combineFilterConditions combines multiple filter conditions using their concatenators (AND/OR)
func (ts *TaskService) combineFilterConditions(dbFilters []builder.Cond, rawFilters []*taskFilter) builder.Cond {
	if len(dbFilters) == 0 {
		return nil
	}

	if len(dbFilters) == 1 {
		return dbFilters[0]
	}

	var filterCond builder.Cond
	for i, f := range dbFilters {
		if len(dbFilters) > i+1 {
			concat := rawFilters[i+1].concatenator
			switch concat {
			case taskFilterConcatOr:
				filterCond = builder.Or(filterCond, f, dbFilters[i+1])
			case taskFilterConcatAnd:
				filterCond = builder.And(filterCond, f, dbFilters[i+1])
			}
		}
	}

	return filterCond
}

// getRelevantProjectsFromCollection determines which projects are relevant for the collection
func (ts *TaskService) getRelevantProjectsFromCollection(s *xorm.Session, a web.Auth, tf *models.TaskCollection) (projects []*models.Project, err error) {
	// Guard against nil session
	if s == nil {
		return nil, fmt.Errorf("database session is required")
	}

	// Check if this is a saved filter (negative project ID)
	isSavedFilter := tf.ProjectID < 0

	if tf.ProjectID == 0 || isSavedFilter {
		// For saved filters or general queries, get all accessible projects
		projectService := NewProjectService(ts.DB)
		projects, _, _, err := projectService.GetAllForUser(s, &user.User{ID: a.GetID()}, "", 0, -1, false)
		return projects, err
	}

	// Check the project exists and the user has access on it
	project := &models.Project{ID: tf.ProjectID}
	canRead, _, err := project.CanRead(s, a)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, models.ErrUserDoesNotHaveAccessToProject{
			ProjectID: tf.ProjectID,
			UserID:    a.GetID(),
		}
	}

	return []*models.Project{{ID: tf.ProjectID}}, nil
}

// handleSavedFilter processes saved filter requests (negative project IDs)
func (ts *TaskService) handleSavedFilter(s *xorm.Session, collection *models.TaskCollection, a web.Auth, search string, page int, perPage int) (interface{}, int, int64, error) {
	// Get the saved filter ID from the project ID
	savedFilterID := models.GetSavedFilterIDFromProjectID(collection.ProjectID)
	if savedFilterID == 0 {
		return nil, 0, 0, fmt.Errorf("invalid saved filter project ID: %d", collection.ProjectID)
	}

	// Load the saved filter
	savedFilter, err := ts.Registry.SavedFilter().GetByIDSimple(s, savedFilterID)
	if err != nil {
		return nil, 0, 0, err
	}

	// Apply the saved filter's settings to the collection
	savedFilterCollection := savedFilter.Filters
	if savedFilterCollection != nil {
	} else {
	}

	// Merge saved filter settings with current collection
	mergedCollection := &models.TaskCollection{
		ProjectID:          0, // Saved filters search across all projects
		Filter:             savedFilterCollection.Filter,
		FilterIncludeNulls: savedFilterCollection.FilterIncludeNulls,
		FilterTimezone:     savedFilterCollection.FilterTimezone,
		SortBy:             collection.SortBy,
		OrderBy:            collection.OrderBy,
		SortByArr:          collection.SortByArr,
		OrderByArr:         collection.OrderByArr,
		ProjectViewID:      0, // Will handle view separately
		Expand:             collection.Expand,
	}

	// If there's an incoming filter from the URL, combine it with the saved filter
	// (This allows users to add additional filters on top of the saved filter)
	if collection.Filter != "" {
		if mergedCollection.Filter != "" {
			mergedCollection.Filter = "(" + collection.Filter + ") && (" + mergedCollection.Filter + ")"
		} else {
			mergedCollection.Filter = collection.Filter
		}
	}

	// If the saved filter has sort order, use it (unless overridden by current collection)
	if len(collection.SortBy) == 0 && len(collection.SortByArr) == 0 {
		if savedFilterCollection.SortBy != nil {
			mergedCollection.SortBy = savedFilterCollection.SortBy
		}
		if savedFilterCollection.OrderBy != nil {
			mergedCollection.OrderBy = savedFilterCollection.OrderBy
		}
	}

	// If there's a view ID, fetch the view using the ORIGINAL project ID (the negative one)
	// because views are associated with the saved filter's virtual project ID
	var view *models.ProjectView
	if collection.ProjectViewID != 0 {
		var viewErr error
		view, viewErr = ts.Registry.ProjectViews().GetByIDAndProject(s, collection.ProjectViewID, collection.ProjectID)
		if viewErr != nil {
			return nil, 0, 0, viewErr
		}

		// Apply view filters to the merged collection
		if view.Filter != nil {
			if view.Filter.Filter != "" {
				if mergedCollection.Filter != "" {
					mergedCollection.Filter = "(" + mergedCollection.Filter + ") && (" + view.Filter.Filter + ")"
				} else {
					mergedCollection.Filter = view.Filter.Filter
				}
			}
			// Note: We don't apply an empty view filter - it would override the saved filter

			if view.Filter.FilterTimezone != "" {
				mergedCollection.FilterTimezone = view.Filter.FilterTimezone
			}

			if view.Filter.FilterIncludeNulls {
				mergedCollection.FilterIncludeNulls = view.Filter.FilterIncludeNulls
			}

			if view.Filter.Search != "" {
				search = view.Filter.Search
			}
		}
	}

	// Convert collection parameters to search options
	opts, err := ts.getTaskFilterOptsFromCollection(mergedCollection, view)
	if err != nil {
		return nil, 0, 0, err
	}

	// Add the id parameter as the last parameter to sortby by default
	if len(opts.sortby) == 0 ||
		len(opts.sortby) > 0 && opts.sortby[len(opts.sortby)-1].sortBy != "id" {
		opts.sortby = append(opts.sortby, &sortParam{
			sortBy:  "id",
			orderBy: orderAscending,
		})
	}

	// Set pagination and search parameters
	opts.search = search
	opts.page = page
	opts.perPage = perPage

	// Get projects the user has access to
	projects, err := ts.getRelevantProjectsFromCollection(s, a, mergedCollection)
	if err != nil {
		return nil, 0, 0, err
	}

	// Get tasks with the saved filter applied
	return ts.getTasksForProjects(s, projects, a, opts, view)
}

// processRegularCollection handles the standard project collection processing
func (ts *TaskService) processRegularCollection(s *xorm.Session, collection *models.TaskCollection, a web.Auth, search string, page int, perPage int) (interface{}, int, int64, error) {
	// This contains the rest of the original GetAllWithFullFiltering logic
	var view *models.ProjectView
	var filteringForBucket bool
	var err error

	if collection.ProjectViewID != 0 {
		view, err = ts.Registry.ProjectViews().GetByIDAndProject(s, collection.ProjectViewID, collection.ProjectID)
		if err != nil {
			return nil, 0, 0, err
		}

		// Apply view filters to collection filters
		if view.Filter != nil {
			if view.Filter.Filter != "" {
				if collection.Filter != "" {
					collection.Filter = "(" + collection.Filter + ") && (" + view.Filter.Filter + ")"
				} else {
					collection.Filter = view.Filter.Filter
				}
			}

			if view.Filter.FilterTimezone != "" {
				collection.FilterTimezone = view.Filter.FilterTimezone
			}

			if view.Filter.FilterIncludeNulls {
				collection.FilterIncludeNulls = view.Filter.FilterIncludeNulls
			}

			if view.Filter.Search != "" {
				search = view.Filter.Search
			}
		}

		// Check for bucket filtering
		if collection.Filter != "" && strings.Contains(collection.Filter, taskPropertyBucketID) {
			filteringForBucket = true
			// For now, skip bucket filter conversion - we'll add this later
		}
	}

	// Step 3: Convert collection parameters to search options
	opts, err := ts.getTaskFilterOptsFromCollection(collection, view)
	if err != nil {
		return nil, 0, 0, err
	}

	// Add the id parameter as the last parameter to sortby by default, but only if it is not already passed as the last parameter
	if len(opts.sortby) == 0 ||
		len(opts.sortby) > 0 && opts.sortby[len(opts.sortby)-1].sortBy != "id" {
		opts.sortby = append(opts.sortby, &sortParam{
			sortBy:  "id",
			orderBy: orderAscending,
		})
	}

	// Step 4: Validate expansion options
	for _, expandValue := range collection.Expand {
		err = expandValue.Validate()
		if err != nil {
			return nil, 0, 0, err
		}
	}

	// Set search options
	opts.search = search
	opts.page = page
	opts.perPage = perPage
	opts.expand = collection.Expand

	// Step 5: Add position sorting for views
	if view != nil {
		var hasOrderByPosition bool
		for _, param := range opts.sortby {
			if param.sortBy == taskPropertyPosition {
				hasOrderByPosition = true
				break
			}
		}
		if !hasOrderByPosition {
			opts.sortby = append(opts.sortby, &sortParam{
				projectViewID: view.ID,
				sortBy:        taskPropertyPosition,
				orderBy:       orderAscending,
			})
		}
	}

	// Step 6: Handle LinkSharing authentication
	shareAuth, is := a.(*models.LinkSharing)
	if is {
		project, err := ts.Registry.Project().GetByIDSimple(s, shareAuth.ProjectID)
		if err != nil {
			return nil, 0, 0, err
		}
		return ts.getTaskOrTasksInBuckets(s, a, []*models.Project{project}, view, opts, filteringForBucket)
	}

	// Step 7: Get relevant projects for the user
	projects, err := ts.getRelevantProjectsFromCollection(s, a, collection)
	if err != nil {
		return nil, 0, 0, err
	}

	// Step 8: Get tasks (or tasks in buckets)
	return ts.getTaskOrTasksInBuckets(s, a, projects, view, opts, filteringForBucket)
}

// GetAllWithFullFiltering implements the complete Task ReadAll functionality
// This method contains all the complex filtering, sorting, and permission logic
// that was previously in models.TaskCollection.ReadAll()
func (ts *TaskService) GetAllWithFullFiltering(s *xorm.Session, collection *models.TaskCollection, a web.Auth, search string, page int, perPage int) (interface{}, int, int64, error) {
	// Step 1: Handle special project IDs
	if collection.ProjectID < 0 {
		// Handle favorites pseudo-project
		if collection.ProjectID == models.FavoritesPseudoProjectID {
			return ts.handleFavorites(s, collection, a, search, page, perPage)
		}
		// Handle saved filters (project ID < -1)
		return ts.handleSavedFilter(s, collection, a, search, page, perPage)
	}

	// Step 2: Handle regular collections
	return ts.processRegularCollection(s, collection, a, search, page, perPage)
}

// handleFavorites processes favorites pseudo-project requests
func (ts *TaskService) handleFavorites(s *xorm.Session, collection *models.TaskCollection, a web.Auth, search string, page int, perPage int) (interface{}, int, int64, error) {
	// Get user from auth
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}

	// Get all favorite task IDs for this user
	favs := []*models.Favorite{}
	err = s.Where(builder.And(
		builder.Eq{"user_id": u.ID},
		builder.Eq{"kind": models.FavoriteKindTask},
	)).Find(&favs)
	if err != nil {
		return nil, 0, 0, err
	}

	// Extract the task IDs
	favoriteTaskIDs := make([]int64, 0, len(favs))
	for _, fav := range favs {
		favoriteTaskIDs = append(favoriteTaskIDs, fav.EntityID)
	}

	// If no favorites, return empty result
	if len(favoriteTaskIDs) == 0 {
		return []*models.Task{}, 0, 0, nil
	}

	// Get the tasks with all the details for these favorite task IDs
	// We need to use the models bridge to ensure we get full task details
	// First, let's get the projects that contain these tasks
	projects, err := ts.getRelevantProjectsFromCollection(s, a, &models.TaskCollection{ProjectID: 0})
	if err != nil {
		return nil, 0, 0, err
	}

	// Convert collection to search options
	opts, err := ts.getTaskFilterOptsFromCollection(collection, nil)
	if err != nil {
		return nil, 0, 0, err
	}

	// Add the id parameter as the last parameter to sortby by default
	if len(opts.sortby) == 0 ||
		len(opts.sortby) > 0 && opts.sortby[len(opts.sortby)-1].sortBy != "id" {
		opts.sortby = append(opts.sortby, &sortParam{
			sortBy:  "id",
			orderBy: orderAscending,
		})
	}

	// Set search options
	opts.search = search
	opts.page = page
	opts.perPage = perPage
	opts.expand = collection.Expand

	// Call a special method to get favorite tasks with full details
	return ts.getFavoriteTasksWithDetails(s, projects, a, favoriteTaskIDs, opts)
}

// getFavoriteTasksWithDetails gets favorite tasks with full details (assignees, labels, etc.)
func (ts *TaskService) getFavoriteTasksWithDetails(s *xorm.Session, projects []*models.Project, a web.Auth, favoriteTaskIDs []int64, opts *taskSearchOptions) (tasks []*models.Task, resultCount int, totalItems int64, err error) {
	if len(favoriteTaskIDs) == 0 {
		return []*models.Task{}, 0, 0, nil
	}

	// We need to call the models bridge function but filter the results to only include favorites
	// First get all tasks using the bridge
	allTasks, _, _, err := models.CallGetTasksForProjects(
		s,
		projects,
		a,
		opts.search,
		0,  // Get all pages for now
		-1, // No limit for now
		convertSortParamsToStrings(opts.sortby),
		convertSortParamsToOrderStrings(opts.sortby),
		opts.filterIncludeNulls,
		opts.filter,
		opts.filterTimezone,
		opts.expand,
	)
	if err != nil {
		return nil, 0, 0, err
	}

	// Filter to only include favorites
	favoritesMap := make(map[int64]bool)
	for _, id := range favoriteTaskIDs {
		favoritesMap[id] = true
	}

	var favoriteTasks []*models.Task
	for _, task := range allTasks {
		if favoritesMap[task.ID] {
			favoriteTasks = append(favoriteTasks, task)
		}
	}

	// Apply pagination to the filtered results
	totalItems = int64(len(favoriteTasks))

	// Handle pagination
	if opts.perPage <= 0 {
		// No pagination - return all results
		return favoriteTasks, len(favoriteTasks), totalItems, nil
	}

	page := opts.page
	if page <= 0 {
		page = 1 // Default to page 1
	}

	start := (page - 1) * opts.perPage
	end := start + opts.perPage

	if start >= len(favoriteTasks) {
		return []*models.Task{}, 0, totalItems, nil
	}

	if end > len(favoriteTasks) {
		end = len(favoriteTasks)
	}

	favoriteTasks = favoriteTasks[start:end]
	return favoriteTasks, len(favoriteTasks), totalItems, nil
}

// getTaskIndexFromSearchString extracts a task index number from a search string
// For example, "#17" in the search string "number #17" will return 17
func getTaskIndexFromSearchString(s string) (index int64) {
	re := regexp.MustCompile("#([0-9]+)")
	in := re.FindString(s)

	stringIndex := strings.ReplaceAll(in, "#", "")
	index, _ = strconv.ParseInt(stringIndex, 10, 64)
	return
}

// getTaskOrTasksInBuckets determines whether to return tasks or buckets
func (ts *TaskService) getTaskOrTasksInBuckets(s *xorm.Session, a web.Auth, projects []*models.Project, view *models.ProjectView, opts *taskSearchOptions, filteringForBucket bool) (tasks interface{}, resultCount int, totalItems int64, err error) {
	if filteringForBucket {
		return ts.getTasksForProjects(s, projects, a, opts, view)
	}

	if view != nil && !strings.Contains(opts.filter, taskPropertyBucketID) {
		if view.BucketConfigurationMode != models.BucketConfigurationModeNone {
			// For now, delegate bucket handling to models - this is complex functionality
			// TODO: Move bucket logic to service layer
			return []*models.Bucket{}, 0, 0, nil // Simplified for now
		}
	}

	return ts.getTasksForProjects(s, projects, a, opts, view)
}

// getTasksForProjects gets tasks for the specified projects with full details
func (ts *TaskService) getTasksForProjects(s *xorm.Session, projects []*models.Project, a web.Auth, opts *taskSearchOptions, view *models.ProjectView) (tasks []*models.Task, resultCount int, totalItems int64, err error) {
	// Use the service layer's query building to properly apply filters
	// This fixes the saved filters bug where filters were being ignored

	// Extract project IDs
	projectIDs := make([]int64, 0, len(projects))
	for _, p := range projects {
		if p.ID != models.FavoritesPseudoProject.ID {
			projectIDs = append(projectIDs, p.ID)
		}
	}

	if len(projectIDs) == 0 {
		return []*models.Task{}, 0, 0, nil
	}

	// Set project IDs in opts if not already set
	if len(opts.projectIDs) == 0 {
		opts.projectIDs = projectIDs
	}

	// Build all conditions using builder.And() like the models layer does
	var whereCond builder.Cond

	// Project ID condition
	projectIDCond := builder.In("tasks.project_id", opts.projectIDs)
	whereCond = projectIDCond

	// Search condition
	if opts.search != "" {
		searchCond := db.MultiFieldSearchWithTableAlias([]string{"title", "description"}, opts.search, "tasks")

		// Check if search contains a task index (e.g., "#17")
		searchIndex := getTaskIndexFromSearchString(opts.search)
		if searchIndex > 0 {
			searchCond = builder.Or(searchCond, builder.Eq{"`index`": searchIndex})
		}

		whereCond = builder.And(whereCond, searchCond)
	}
	// Apply custom filters if present
	if opts.parsedFilters != nil && len(opts.parsedFilters) > 0 {
		filterCond, err := ts.convertFiltersToDBFilterCond(opts.parsedFilters, opts.filterIncludeNulls)
		if err != nil {
			return nil, 0, 0, err
		}
		whereCond = builder.And(whereCond, filterCond)
	}

	// Get total count using the same conditions
	totalItems, err = s.Table("tasks").Where(whereCond).Count(&models.Task{})
	if err != nil {
		return nil, 0, 0, err
	}

	// Determine if we need DISTINCT with additional fields for sorting by position
	// This prevents duplicate results when LEFT JOINing task_positions or task_buckets
	var distinct = "tasks.*"
	var hasSortByPosition = false
	for _, param := range opts.sortby {
		if param.sortBy == "position" {
			distinct += ", task_positions.position"
			hasSortByPosition = true
			break
		}
	}

	// Build complete query in one go
	baseQuery := s.Table("tasks").Where(whereCond).Distinct(distinct)

	// Add JOINs for position or bucket_id sorting
	if hasSortByPosition {
		for _, param := range opts.sortby {
			if param.sortBy == "position" {
				baseQuery = baseQuery.Join("LEFT", "task_positions", "task_positions.task_id = tasks.id AND task_positions.project_view_id = ?", param.projectViewID)
				break
			}
		}
	}

	// Check if we need to join task_buckets for bucket_id filtering or sorting
	joinTaskBuckets := false
	if opts.parsedFilters != nil {
		for _, filter := range opts.parsedFilters {
			if filter.field == "bucket_id" {
				joinTaskBuckets = true
				break
			}
		}
	}
	for _, param := range opts.sortby {
		if param.sortBy == "bucket_id" {
			joinTaskBuckets = true
			break
		}
	}

	if joinTaskBuckets {
		joinCond := "task_buckets.task_id = tasks.id"
		if opts.projectViewID > 0 {
			baseQuery = baseQuery.Join("LEFT", "task_buckets", joinCond+" AND task_buckets.project_view_id = ?", opts.projectViewID)
		} else {
			baseQuery = baseQuery.Join("LEFT", "task_buckets", joinCond)
		}
	}

	// Build order by clause using proper field prefixes
	var orderby string
	for i, param := range opts.sortby {
		if i > 0 {
			orderby += ", "
		}

		var prefix string
		switch param.sortBy {
		case "position":
			prefix = "task_positions."
		case "bucket_id":
			prefix = "task_buckets."
		default:
			prefix = "tasks."
		}

		// MySQL sorts null values first - add IS NULL check to make it consistent
		if db.Type() == schemas.MYSQL {
			orderby += prefix + "`" + param.sortBy + "` IS NULL, "
		}

		orderby += prefix + "`" + param.sortBy + "` " + param.orderBy.String()

		// Postgres and SQLite allow NULLS LAST for consistent sorting
		if db.Type() == schemas.POSTGRES || db.Type() == schemas.SQLITE {
			orderby += " NULLS LAST"
		}
	}

	if orderby != "" {
		baseQuery = baseQuery.OrderBy(orderby)
	}

	// Apply pagination
	if opts.page > 0 && opts.perPage > 0 {
		baseQuery = baseQuery.Limit(opts.perPage, (opts.page-1)*opts.perPage)
	}

	// Execute query to get raw tasks
	tasks = []*models.Task{}
	err = baseQuery.Find(&tasks)
	if err != nil {
		return nil, 0, 0, err
	}

	// Create a map of tasks for adding more info
	taskMap := make(map[int64]*models.Task, len(tasks))
	for i, t := range tasks {
		taskMap[t.ID] = tasks[i] // Use tasks[i] to ensure we get the pointer from the slice
	}

	// Add additional details (labels, assignees, attachments, etc.)
	err = models.AddMoreInfoToTasks(s, taskMap, a, view, opts.expand)
	if err != nil {
		return nil, 0, 0, err
	}

	return tasks, len(tasks), totalItems, nil
}

// getRawTasksForProjects gets the basic task data without extra details
func (ts *TaskService) getRawTasksForProjects(s *xorm.Session, projects []*models.Project, a web.Auth, opts *taskSearchOptions) (tasks []*models.Task, resultCount int, totalItems int64, err error) {
	// For now, delegate back to the models package's getRawTasksForProjects function
	// This ensures all existing filtering, sorting, and search logic continues to work
	// while we're in the process of moving it to the service layer
	// TODO: Move all filtering logic to service layer completely

	// Use the bridge function that calls getRawTasksForProjects directly (not getTasksForProjects)
	return models.CallGetRawTasksForProjects(
		s,
		projects,
		a,
		opts.search,
		opts.page,
		opts.perPage,
		convertSortParamsToStrings(opts.sortby),
		convertSortParamsToOrderStrings(opts.sortby),
		opts.filterIncludeNulls,
		opts.filter,
		opts.filterTimezone,
		opts.expand,
	)
}

// convertSortParamsToStrings converts sortParam structs to strings for TaskCollection
func convertSortParamsToStrings(sortParams []*sortParam) []string {
	if len(sortParams) == 0 {
		return nil
	}

	result := make([]string, len(sortParams))
	for i, param := range sortParams {
		result[i] = param.sortBy
	}
	return result
}

// convertSortParamsToOrderStrings converts sortParam order to strings for TaskCollection
func convertSortParamsToOrderStrings(sortParams []*sortParam) []string {
	if len(sortParams) == 0 {
		return nil
	}

	result := make([]string, len(sortParams))
	for i, param := range sortParams {
		if param.orderBy == orderDescending {
			result[i] = "desc"
		} else {
			result[i] = "asc"
		}
	}
	return result
}

// TaskService represents a service for managing tasks.
type TaskService struct {
	DB       *xorm.Engine
	Registry *ServiceRegistry
}

// NewTaskService creates a new TaskService.
// Deprecated: Use ServiceRegistry.Task() instead.
func NewTaskService(db *xorm.Engine) *TaskService {
	registry := NewServiceRegistry(db)
	return registry.Task()
}

// Wire models.AddMoreInfoToTasksFunc to the service implementation via dependency inversion
// InitTaskService sets up dependency injection for task-related model functions.
// This function must be called during test initialization to ensure models can call services.
func InitTaskService() {
	models.AddMoreInfoToTasksFunc = func(s *xorm.Session, taskMap map[int64]*models.Task, a web.Auth, view *models.ProjectView, expand []models.TaskCollectionExpandable) error {
		return NewTaskService(nil).AddDetailsToTasks(s, taskMap, a, view, expand)
	}

	models.GetUsersOrLinkSharesFromIDsFunc = func(s *xorm.Session, ids []int64) (map[int64]*user.User, error) {
		return NewTaskService(nil).getUsersOrLinkSharesFromIDs(s, ids)
	}

	models.TaskCreateFunc = func(s *xorm.Session, task *models.Task, u *user.User, updateAssignees bool, setBucket bool) error {
		_, err := NewTaskService(s.Engine()).CreateWithOptions(s, task, u, updateAssignees, setBucket, false)
		return err
	}

	// Wire TaskCollection.ReadAll to our new service method
	models.TaskCollectionReadAllFunc = func(s *xorm.Session, tf *models.TaskCollection, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
		return NewTaskService(s.Engine()).GetAllWithFullFiltering(s, tf, a, search, page, perPage)
	}

	// Wire GetTaskByIDSimple to service layer (T-PERM-014A Phase 3)
	models.GetTaskByIDSimpleFunc = func(s *xorm.Session, taskID int64) (*models.Task, error) {
		ts := NewTaskService(s.Engine())
		return ts.GetByIDSimple(s, taskID)
	}

	// Set up permission delegation (T-PERM-007)
	models.CheckTaskReadFunc = func(s *xorm.Session, taskID int64, a web.Auth) (bool, int, error) {
		ts := NewTaskService(s.Engine())
		return ts.CanRead(s, taskID, a)
	}
	models.CheckTaskWriteFunc = func(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
		ts := NewTaskService(s.Engine())
		return ts.CanWrite(s, taskID, a)
	}
	models.CheckTaskUpdateFunc = func(s *xorm.Session, taskID int64, task *models.Task, a web.Auth) (bool, error) {
		ts := NewTaskService(s.Engine())
		return ts.CanUpdate(s, taskID, task, a)
	}
	models.CheckTaskDeleteFunc = func(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
		ts := NewTaskService(s.Engine())
		return ts.CanDelete(s, taskID, a)
	}
	models.CheckTaskCreateFunc = func(s *xorm.Session, task *models.Task, a web.Auth) (bool, error) {
		ts := NewTaskService(s.Engine())
		return ts.CanCreate(s, task, a)
	}

	// TaskAssignee permission delegation
	// Note: No dedicated TaskAssignee service yet, so we use project permissions
	models.CheckTaskAssigneeCreateFunc = func(s *xorm.Session, assignee *models.TaskAssginee, a web.Auth) (bool, error) {
		project, err := models.GetProjectSimpleByTaskID(s, assignee.TaskID)
		if err != nil {
			return false, err
		}
		return project.CanUpdate(s, a)
	}
	models.CheckTaskAssigneeDeleteFunc = func(s *xorm.Session, assignee *models.TaskAssginee, a web.Auth) (bool, error) {
		project, err := models.GetProjectSimpleByTaskID(s, assignee.TaskID)
		if err != nil {
			return false, err
		}
		return project.CanUpdate(s, a)
	}

	// TaskRelation permission delegation
	// Note: No dedicated TaskRelation service yet, using task permission checks
	models.CheckTaskRelationCreateFunc = func(s *xorm.Session, relation *models.TaskRelation, a web.Auth) (bool, error) {
		// Check if the relation kind is valid
		if !relation.RelationKind.IsValid() {
			return false, models.ErrInvalidRelationKind{Kind: relation.RelationKind}
		}

		// Need write access to the base task and at least read access to the other task
		baseTask := &models.Task{ID: relation.TaskID}
		has, err := baseTask.CanUpdate(s, a)
		if err != nil || !has {
			return false, err
		}

		otherTask := &models.Task{ID: relation.OtherTaskID}
		has, _, err = otherTask.CanRead(s, a)
		if err != nil {
			return false, err
		}
		return has, nil
	}
	models.CheckTaskRelationDeleteFunc = func(s *xorm.Session, relation *models.TaskRelation, a web.Auth) (bool, error) {
		// A user can delete a relation if they can update the base task
		baseTask := &models.Task{ID: relation.TaskID}
		return baseTask.CanUpdate(s, a)
	}
}

// GetByIDSimple retrieves a task by ID without permission checks or extra data
// This is a simple lookup helper used by permission methods
// MIGRATION: Added in T-PERM-004 (migrated from models.GetTaskByIDSimple)
func (ts *TaskService) GetByIDSimple(s *xorm.Session, taskID int64) (*models.Task, error) {
	if taskID < 1 {
		return nil, models.ErrTaskDoesNotExist{ID: taskID}
	}

	task := &models.Task{}
	exists, err := s.Where("id = ?", taskID).Get(task)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, models.ErrTaskDoesNotExist{ID: taskID}
	}
	return task, nil
}

// GetByIDs retrieves multiple tasks by their IDs without additional details
func (ts *TaskService) GetByIDs(s *xorm.Session, ids []int64) ([]*models.Task, error) {
	if len(ids) == 0 {
		return []*models.Task{}, nil
	}

	tasks := []*models.Task{}
	err := s.In("id", ids).Find(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetByID gets a single task by its ID, checking permissions.
func (ts *TaskService) GetByID(s *xorm.Session, taskID int64, u *user.User) (*models.Task, error) {
	// Use a simple model function to get the raw data
	task := new(models.Task)
	has, err := s.ID(taskID).Get(task)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, models.ErrTaskDoesNotExist{ID: taskID}
	}

	// Permission Check: The TaskService asks the ProjectService for a decision.
	projectService := NewProjectService(ts.DB)
	can, err := projectService.HasPermission(s, task.ProjectID, u, models.PermissionRead)
	if err != nil {
		return nil, fmt.Errorf("checking project read permission: %w", err)
	}
	if !can {
		return nil, ErrAccessDenied
	}

	// Add details to the task
	taskMap := map[int64]*models.Task{task.ID: task}
	err = ts.AddDetailsToTasks(s, taskMap, u, nil, nil)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetByIDWithExpansion gets a single task by its ID with support for expansion parameters
// and returns the maximum permission the user has on the task's project.
func (ts *TaskService) GetByIDWithExpansion(s *xorm.Session, taskID int64, u *user.User, expand []models.TaskCollectionExpandable) (*models.Task, int, error) {
	// Load the task with all fields at the service layer
	task := &models.Task{}
	exists, err := s.Where("id = ?", taskID).Get(task)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, models.ErrTaskDoesNotExist{ID: taskID}
	}

	// Permission Check: The TaskService asks the ProjectService for a decision.
	projectService := NewProjectService(ts.DB)
	permissionMap, err := projectService.checkPermissionsForProjects(s, u, []int64{task.ProjectID})
	if err != nil {
		return nil, 0, fmt.Errorf("checking project permissions: %w", err)
	}
	permission, ok := permissionMap[task.ProjectID]
	if !ok || permission == nil {
		return nil, 0, ErrAccessDenied
	}
	maxPermission := permission.MaxPermission
	if maxPermission < int(models.PermissionRead) {
		return nil, 0, ErrAccessDenied
	}

	// Add details to the task with expansion support
	taskMap := map[int64]*models.Task{task.ID: task}
	err = ts.AddDetailsToTasks(s, taskMap, u, nil, expand)
	if err != nil {
		return nil, 0, err
	}

	// Load subscription data for single task requests (matches original behavior)
	subscription, err := models.GetSubscriptionForUser(s, models.SubscriptionEntityTask, task.ID, u)
	if err != nil && !models.IsErrProjectDoesNotExist(err) {
		return nil, 0, err
	}
	if subscription != nil {
		task.Subscription = &subscription.Subscription
	}

	return task, maxPermission, nil
}

// GetAllByProject gets all tasks for a project with pagination and filtering
func (ts *TaskService) GetAllByProject(s *xorm.Session, projectID int64, u *user.User, page int, perPage int, search string) ([]*models.Task, int, int64, error) {
	// Handle saved filters (negative project IDs)
	if projectID < 0 {
		// Create a TaskCollection to use the existing saved filter handling
		collection := &models.TaskCollection{
			ProjectID: projectID,
			Search:    search,
		}
		result, resultCount, totalItems, err := ts.GetAllWithFullFiltering(s, collection, u, search, page, perPage)
		if err != nil {
			return nil, 0, 0, err
		}
		// Convert result to []*models.Task
		if tasks, ok := result.([]*models.Task); ok {
			return tasks, resultCount, totalItems, nil
		}
		// If not a simple task array, return empty (shouldn't happen for saved filters)
		return []*models.Task{}, 0, 0, nil
	}

	// Permission Check: Use ProjectService for proper inter-service communication
	projectService := NewProjectService(ts.DB)
	canRead, err := projectService.HasPermission(s, projectID, u, models.PermissionRead)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrAccessDenied
	}

	// Calculate offset for pagination
	offset := (page - 1) * perPage

	// Query tasks directly from the database
	var tasks []*models.Task

	// Add search filter if provided
	searchCondition := builder.NewCond()
	if search != "" {
		searchCondition = builder.Or(
			builder.Like{"title", "%" + search + "%"},
			builder.Like{"description", "%" + search + "%"},
		)
	}

	// Get total count for pagination (use separate query to avoid session corruption)
	countQuery := s.Where("project_id = ?", projectID)
	if search != "" {
		countQuery = countQuery.And(searchCondition)
	}
	totalCount, err := countQuery.Count(&models.Task{})
	if err != nil {
		return nil, 0, 0, err
	}

	// Create fresh query for finding tasks to avoid any session corruption
	findQuery := s.Where("project_id = ?", projectID)
	if search != "" {
		findQuery = findQuery.And(searchCondition)
	}

	// Get the actual tasks with pagination
	err = findQuery.
		OrderBy("id ASC").
		Limit(perPage, offset).
		Find(&tasks)
	if err != nil {
		return nil, 0, 0, err
	}

	// Add details to all tasks (CreatedBy, Labels, Attachments, etc.)
	if len(tasks) > 0 {
		taskMap := make(map[int64]*models.Task)
		for _, task := range tasks {
			taskMap[task.ID] = task
		}
		err = ts.AddDetailsToTasks(s, taskMap, u, nil, nil)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	return tasks, len(tasks), totalCount, nil
}

// GetAllWithFilters gets all tasks with complex filtering, sorting and expansion options
// This method replicates the functionality of models.TaskCollection.ReadAll() at the service layer
func (ts *TaskService) GetAllWithFilters(s *xorm.Session, collection *models.TaskCollection, a web.Auth, search string, page int, perPage int) ([]*models.Task, int, int64, error) {
	// Use our new full filtering implementation
	result, resultCount, totalItems, err := ts.GetAllWithFullFiltering(s, collection, a, search, page, perPage)
	if err != nil {
		return nil, 0, 0, err
	}

	tasks, ok := result.([]*models.Task)
	if !ok {
		return nil, 0, 0, fmt.Errorf("unexpected result type from GetAllWithFullFiltering")
	}

	return tasks, resultCount, totalItems, nil
}

// Update updates a task with full business logic.
func (ts *TaskService) Update(s *xorm.Session, task *models.Task, u *user.User) (*models.Task, error) {
	updatedTask, err := ts.updateSingleTask(s, task, u, nil)
	return updatedTask, err
}

// UpdateWithFields updates a task with only specific fields.
func (ts *TaskService) UpdateWithFields(s *xorm.Session, task *models.Task, u *user.User, fields []string) (*models.Task, error) {
	return ts.updateSingleTask(s, task, u, fields)
}

//nolint:gocyclo
func (ts *TaskService) updateSingleTask(s *xorm.Session, t *models.Task, u *user.User, fields []string) (*models.Task, error) {
	// Check if the task exists and get the old values FIRST (before permission check)
	// This is necessary because t.ProjectID might be 0
	otPtr, err := ts.GetByIDSimple(s, t.ID)
	if err != nil {
		return nil, err
	}
	ot := *otPtr

	// Now check permissions using the old task (which has the correct ProjectID)
	can, err := ts.Can(s, &ot, u).Write()
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrAccessDenied
	}

	if t.ProjectID == 0 {
		t.ProjectID = ot.ProjectID
	}

	// Get the stored reminders
	reminders, err := models.GetRemindersForTasks(s, []int64{t.ID})
	if err != nil {
		return nil, err
	}

	// Old task has the stored reminders
	ot.Reminders = reminders

	// Update the assignees
	// Pass the user as web.Auth for model methods that need it
	if err := ot.UpdateTaskAssignees(s, t.Assignees, u); err != nil {
		return nil, err
	}

	// Update the labels if provided
	if t.Labels != nil {
		if err := ts.syncTaskLabels(s, &ot, t.Labels, u); err != nil {
			return nil, err
		}
		// Copy the updated labels back to the task being returned
		t.Labels = ot.Labels
	}

	// All columns to update in a separate variable to be able to add to them
	colsToUpdate := []string{
		"title",
		"description",
		"done",
		"due_date",
		"repeat_after",
		"priority",
		"start_date",
		"end_date",
		"hex_color",
		"percent_done",
		"project_id",
		"bucket_id",
		"repeat_mode",
		"cover_image_attachment_id",
	}

	// Validate fields if provided
	if len(fields) > 0 {
		allowed := map[string]bool{}
		for _, c := range colsToUpdate {
			allowed[c] = true
		}
		cols := []string{}
		fieldSet := map[string]bool{}
		for _, f := range fields {
			if !allowed[f] {
				return nil, models.ErrInvalidTaskColumn{Column: f}
			}
			cols = append(cols, f)
			fieldSet[f] = true
		}
		colsToUpdate = cols

		if !fieldSet["title"] {
			t.Title = ot.Title
		}
		if !fieldSet["description"] {
			t.Description = ot.Description
		}
		if !fieldSet["done"] {
			t.Done = ot.Done
			t.DoneAt = ot.DoneAt
		}
		if !fieldSet["due_date"] {
			t.DueDate = ot.DueDate
		}
		if !fieldSet["repeat_after"] {
			t.RepeatAfter = ot.RepeatAfter
		}
		if !fieldSet["priority"] {
			t.Priority = ot.Priority
		}
		if !fieldSet["start_date"] {
			t.StartDate = ot.StartDate
		}
		if !fieldSet["end_date"] {
			t.EndDate = ot.EndDate
		}
		if !fieldSet["hex_color"] {
			t.HexColor = ot.HexColor
		}
		if !fieldSet["percent_done"] {
			t.PercentDone = ot.PercentDone
		}
		if !fieldSet["project_id"] {
			t.ProjectID = ot.ProjectID
		}
		if !fieldSet["bucket_id"] {
			t.BucketID = ot.BucketID
		}
		if !fieldSet["repeat_mode"] {
			t.RepeatMode = ot.RepeatMode
		}
		if !fieldSet["cover_image_attachment_id"] {
			t.CoverImageAttachmentID = ot.CoverImageAttachmentID
		}
	}

	// If the task is being moved between projects, make sure to move the bucket + index as well
	if t.ProjectID != 0 && ot.ProjectID != t.ProjectID {
		t.Index, err = models.CalculateNextTaskIndex(s, t.ProjectID)
		if err != nil {
			return nil, err
		}
		t.BucketID = 0
		colsToUpdate = append(colsToUpdate, "index")
	}

	views := []*models.ProjectView{}
	if (!t.IsRepeating() && t.Done != ot.Done) || t.ProjectID != ot.ProjectID {
		err = s.
			Where("project_id = ? AND view_kind = ? AND bucket_configuration_mode = ?",
				t.ProjectID, models.ProjectViewKindKanban, models.BucketConfigurationModeManual).
			Find(&views)
		if err != nil {
			return nil, err
		}
	}

	// When a task was moved between projects, ensure it is in the correct bucket
	if t.ProjectID != ot.ProjectID {
		_, err = s.Where("task_id = ?", t.ID).Delete(&models.TaskBucket{})
		if err != nil {
			return nil, err
		}
		_, err = s.Where("task_id = ?", t.ID).Delete(&models.TaskPosition{})
		if err != nil {
			return nil, err
		}

		for _, view := range views {
			var bucketID = view.DoneBucketID
			if bucketID == 0 || !t.Done {
				bucketID, err = models.GetDefaultBucketID(s, view)
				if err != nil {
					return nil, err
				}
			}

			tb := &models.TaskBucket{
				BucketID:      bucketID,
				TaskID:        t.ID,
				ProjectViewID: view.ID,
				ProjectID:     t.ProjectID,
			}
			err = tb.Update(s, u)
			if err != nil {
				return nil, err
			}

			tp, err := models.CalculateNewPositionForTask(s, u, t, view)
			if err != nil {
				return nil, err
			}

			err = tp.Update(s, u)
			if err != nil {
				return nil, err
			}
		}
	}

	// When a task changed its done status, make sure it is in the correct bucket
	if t.ProjectID == ot.ProjectID && !t.IsRepeating() && t.Done != ot.Done {
		err = t.MoveTaskToDoneBuckets(s, u, views)
		if err != nil {
			return nil, err
		}
	}

	// When a repeating task is marked as done, we update all deadlines and reminders and set it as undone
	models.UpdateDone(&ot, t)
	colsToUpdate = append(colsToUpdate, "done_at")

	// Update the reminders
	if err := ot.UpdateReminders(s, t); err != nil {
		return nil, err
	}

	// If a task attachment is being set as cover image, check if the attachment actually belongs to the task
	if t.CoverImageAttachmentID != 0 {
		is, err := s.Exist(&models.TaskAttachment{
			TaskID: t.ID,
			ID:     t.CoverImageAttachmentID,
		})
		if err != nil {
			return nil, err
		}
		if !is {
			return nil, &models.ErrAttachmentDoesNotBelongToTask{
				AttachmentID: t.CoverImageAttachmentID,
				TaskID:       t.ID,
			}
		}
	}

	// Handle favorite status changes
	wasFavorite, err := ts.Registry.Favorite().IsFavorite(s, t.ID, u, models.FavoriteKindTask)
	if err != nil {
		return nil, err
	}
	if t.IsFavorite && !wasFavorite {
		if err := ts.Registry.Favorite().AddToFavorite(s, t.ID, u, models.FavoriteKindTask); err != nil {
			return nil, err
		}
	}

	if !t.IsFavorite && wasFavorite {
		if err := ts.Registry.Favorite().RemoveFromFavorite(s, t.ID, u, models.FavoriteKindTask); err != nil {
			return nil, err
		}
	}

	// Merge the old task with the new task
	// mergo ignores nil values, so we need to handle them manually below
	if err := mergo.Merge(&ot, t, mergo.WithOverride); err != nil {
		return nil, err
	}

	t.HexColor = utils.NormalizeHex(t.HexColor)

	// Mergo does ignore nil values. Because of that, we need to check all parameters and set the updated to
	// nil/their nil value in the struct which is inserted.

	// Done
	if !t.Done {
		ot.Done = false
	}
	// Priority
	if t.Priority == 0 {
		ot.Priority = 0
	}
	// Description
	if t.Description == "" {
		ot.Description = ""
	}
	// Due date
	if t.DueDate.IsZero() {
		ot.DueDate = time.Time{}
	}
	// Repeat after
	if t.RepeatAfter == 0 {
		ot.RepeatAfter = 0
	}
	// Start date
	if t.StartDate.IsZero() {
		ot.StartDate = time.Time{}
	}
	// End date
	if t.EndDate.IsZero() {
		ot.EndDate = time.Time{}
	}
	// Color
	if t.HexColor == "" {
		ot.HexColor = ""
	}
	// Percent Done
	if t.PercentDone == 0 {
		ot.PercentDone = 0
	}
	// Repeat from current date
	if t.RepeatMode == models.TaskRepeatModeDefault {
		ot.RepeatMode = models.TaskRepeatModeDefault
	}
	// Is Favorite
	if !t.IsFavorite {
		ot.IsFavorite = false
	}
	// Attachment cover image
	if t.CoverImageAttachmentID == 0 {
		ot.CoverImageAttachmentID = 0
	}

	_, err = s.ID(t.ID).
		Cols(colsToUpdate...).
		Update(ot)
	*t = ot
	if err != nil {
		return nil, err
	}

	// Get the task updated timestamp in a new struct - if we'd just try to put it into t which we already have, it
	// would still contain the old updated date.
	nt := &models.Task{}
	_, err = s.ID(t.ID).Get(nt)
	if err != nil {
		return nil, err
	}
	t.Updated = nt.Updated

	err = events.Dispatch(&models.TaskUpdatedEvent{
		Task: t,
		Doer: u,
	})
	if err != nil {
		return nil, err
	}

	return t, models.UpdateProjectLastUpdated(s, &models.Project{ID: t.ProjectID})
}

// Delete deletes a task.
func (ts *TaskService) Delete(s *xorm.Session, task *models.Task, a web.Auth) error {
	// Check permissions using web.Auth to support both regular users and LinkSharing
	can, err := ts.canWriteTaskWithAuth(s, task.ID, a)
	if err != nil {
		return err
	}
	if !can {
		return ErrAccessDenied
	}

	tPtr, err := ts.GetByIDSimple(s, task.ID)
	if err != nil {
		return err
	}
	t := *tPtr

	// duplicate the task for the event
	fullTask := &models.Task{ID: task.ID}
	err = fullTask.ReadOne(s, a)
	if err != nil {
		return err
	}

	// Delete assignees
	if _, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskAssginee{}); err != nil {
		return err
	}

	// Delete Favorites using the service
	err = ts.Registry.Favorite().RemoveFromFavorite(s, task.ID, a, models.FavoriteKindTask)
	if err != nil {
		return err
	}

	// Delete label associations
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.LabelTask{})
	if err != nil {
		return err
	}

	// Delete task attachments
	attachments, err := ts.getTaskAttachmentsByTaskIDs(s, []int64{task.ID})
	if err != nil {
		return err
	}
	for _, attachment := range attachments {
		// Using the attachment delete method here because that takes care of removing all files properly
		err = attachment.Delete(s, a)
		if err != nil && !models.IsErrTaskAttachmentDoesNotExist(err) {
			return err
		}
	}

	// Delete all comments
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskComment{})
	if err != nil {
		return err
	}

	// Delete all relations
	_, err = s.Where("task_id = ? OR other_task_id = ?", task.ID, task.ID).Delete(&models.TaskRelation{})
	if err != nil {
		return err
	}

	// Delete all reminders
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskReminder{})
	if err != nil {
		return err
	}

	// Delete all positions
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskPosition{})
	if err != nil {
		return err
	}

	// Delete all bucket relations
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskBucket{})
	if err != nil {
		return err
	}

	// Actually delete the task
	_, err = s.ID(task.ID).Delete(&models.Task{})
	if err != nil {
		return err
	}

	doer, _ := user.GetFromAuth(a)
	err = events.Dispatch(&models.TaskDeletedEvent{
		Task: fullTask,
		Doer: doer,
	})
	if err != nil {
		return err
	}

	err = ts.updateProjectLastUpdated(s, t.ProjectID)
	return err
}

// deleteWithoutPermissionCheck deletes a task without checking permissions.
//
// WARNING: This method bypasses ALL permission checks and should ONLY be called in specific,
// well-controlled scenarios where permissions have already been verified at a higher level.
//
// Current valid usage:
//   - ProjectService.Delete(): Permission checks are performed at the project level before
//     cascading to child tasks. Checking permissions again on each task would fail because
//     parent projects may be in the process of being deleted.
//
// SECURITY: This method is intentionally private (unexported) to prevent misuse outside
// the services package. Any new usage must be carefully reviewed for security implications.
func (ts *TaskService) deleteWithoutPermissionCheck(s *xorm.Session, taskID int64, a web.Auth) error {
	tPtr, err := ts.GetByIDSimple(s, taskID)
	if err != nil {
		return err
	}
	t := *tPtr

	// duplicate the task for the event
	fullTask := &models.Task{ID: taskID}
	err = fullTask.ReadOne(s, a)
	if err != nil {
		return err
	}

	// Delete assignees
	if _, err = s.Where("task_id = ?", taskID).Delete(&models.TaskAssginee{}); err != nil {
		return err
	}

	// Delete Favorites using the service
	err = ts.Registry.Favorite().RemoveFromFavorite(s, taskID, a, models.FavoriteKindTask)
	if err != nil {
		return err
	}

	// Delete label associations
	_, err = s.Where("task_id = ?", taskID).Delete(&models.LabelTask{})
	if err != nil {
		return err
	}

	// Delete task attachments
	attachments, err := ts.getTaskAttachmentsByTaskIDs(s, []int64{taskID})
	if err != nil {
		return err
	}
	for _, attachment := range attachments {
		// Using the attachment delete method here because that takes care of removing all files properly
		err = attachment.Delete(s, a)
		if err != nil && !models.IsErrTaskAttachmentDoesNotExist(err) {
			return err
		}
	}

	// Delete all comments
	_, err = s.Where("task_id = ?", taskID).Delete(&models.TaskComment{})
	if err != nil {
		return err
	}

	// Delete all relations
	_, err = s.Where("task_id = ? OR other_task_id = ?", taskID, taskID).Delete(&models.TaskRelation{})
	if err != nil {
		return err
	}

	// Delete all reminders
	_, err = s.Where("task_id = ?", taskID).Delete(&models.TaskReminder{})
	if err != nil {
		return err
	}

	// Delete all positions
	_, err = s.Where("task_id = ?", taskID).Delete(&models.TaskPosition{})
	if err != nil {
		return err
	}

	// Delete all bucket relations
	_, err = s.Where("task_id = ?", taskID).Delete(&models.TaskBucket{})
	if err != nil {
		return err
	}

	// Actually delete the task
	_, err = s.ID(taskID).Delete(&models.Task{})
	if err != nil {
		return err
	}

	doer, _ := user.GetFromAuth(a)
	err = events.Dispatch(&models.TaskDeletedEvent{
		Task: fullTask,
		Doer: doer,
	})
	if err != nil {
		return err
	}

	err = ts.updateProjectLastUpdated(s, t.ProjectID)
	return err
}

// TaskPermissions represents the permissions for a task.
type TaskPermissions struct {
	s    *xorm.Session
	task *models.Task
	user *user.User
	ts   *TaskService
}

// Can returns a new TaskPermissions struct.
func (ts *TaskService) Can(s *xorm.Session, task *models.Task, u *user.User) *TaskPermissions {
	return &TaskPermissions{s: s, task: task, user: u, ts: ts}
}

// Read checks if the user can read the task.
// This implements the "Move Logic, Don't Expose It" principle by moving permission logic from models to services.
func (tp *TaskPermissions) Read() (bool, error) {
	if tp.user == nil {
		return false, nil
	}

	// Use ProjectService for permission checking instead of calling model methods
	projectService := NewProjectService(tp.ts.DB)
	return projectService.HasPermission(tp.s, tp.task.ProjectID, tp.user, models.PermissionRead)
}

// Write checks if the user can write to the task.
// This implements the "Move Logic, Don't Expose It" principle by moving permission logic from models to services.
func (tp *TaskPermissions) Write() (bool, error) {
	if tp.user == nil {
		return false, nil
	}

	// Use ProjectService for permission checking instead of calling model methods
	projectService := NewProjectService(tp.ts.DB)
	return projectService.HasPermission(tp.s, tp.task.ProjectID, tp.user, models.PermissionWrite)
}

func (ts *TaskService) addDetailsToTasks(s *xorm.Session, tasks []*models.Task, u *user.User) error {
	if len(tasks) == 0 {
		return nil
	}

	taskMap := make(map[int64]*models.Task, len(tasks))
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	// Use the standard AddDetailsToTasks method
	return ts.AddDetailsToTasks(s, taskMap, u, nil, nil)
}

// AddDetailsToTasks adds more info to tasks, like assignees, labels, etc.
// This is the service layer implementation of what was previously models.AddMoreInfoToTasks.
// Empty collections are kept as null for standards compliance
func (ts *TaskService) AddDetailsToTasks(s *xorm.Session, taskMap map[int64]*models.Task, a web.Auth, view *models.ProjectView, expand []models.TaskCollectionExpandable) error {
	if len(taskMap) == 0 {
		return nil
	}

	// Initialize array/map fields for consistent API behavior
	// Keep empty collections as null for standards compliance
	for _, task := range taskMap {
		if task.RelatedTasks == nil {
			task.RelatedTasks = make(models.RelatedTaskMap)
		}
	}

	// Collect identifiers for batched lookups
	taskIDs := make([]int64, 0, len(taskMap))
	creatorIDSet := make(map[int64]struct{}, len(taskMap))
	projectIDSet := make(map[int64]struct{}, len(taskMap))
	for _, task := range taskMap {
		taskIDs = append(taskIDs, task.ID)
		if task.CreatedByID != 0 {
			creatorIDSet[task.CreatedByID] = struct{}{}
		}
		projectIDSet[task.ProjectID] = struct{}{}
	}

	// Convert project id set to slice for retrieval
	projectIDs := make([]int64, 0, len(projectIDSet))
	for id := range projectIDSet {
		projectIDs = append(projectIDs, id)
	}

	// Add assignees
	if err := ts.addAssigneesToTasks(s, taskIDs, taskMap); err != nil {
		return err
	}

	// Add labels
	if err := ts.addLabelsToTasks(s, taskIDs, taskMap); err != nil {
		return err
	}

	// Add attachments
	if err := ts.addAttachmentsToTasks(s, taskIDs, taskMap); err != nil {
		return err
	}

	// Get task reminders
	taskReminders, err := ts.getTaskReminderMap(s, taskIDs)
	if err != nil {
		return err
	}

	// Get favorites if auth is provided
	var taskFavorites map[int64]bool
	if a != nil {
		taskFavorites, err = ts.getFavorites(s, taskIDs, a, models.FavoriteKindTask)
		if err != nil {
			return err
		}
	}

	// Get all projects for identifiers
	projects, err := ts.Registry.Project().GetMapByIDs(s, projectIDs)
	if err != nil {
		return err
	}

	// Determine fallback creator assignments for legacy tasks without CreatedByID
	legacyCreators := make(map[int64]int64)
	for _, task := range taskMap {
		if task.CreatedByID != 0 {
			continue
		}
		project := projects[task.ProjectID]
		if project == nil || project.OwnerID == 0 {
			continue
		}
		legacyCreators[task.ID] = project.OwnerID
		if _, seen := creatorIDSet[project.OwnerID]; !seen {
			creatorIDSet[project.OwnerID] = struct{}{}
		}
	}

	// Resolve all required users (task creators + fallbacks)
	userIDs := make([]int64, 0, len(creatorIDSet))
	for id := range creatorIDSet {
		userIDs = append(userIDs, id)
	}

	users := map[int64]*user.User{}
	if len(userIDs) > 0 {
		users, err = ts.getUsersOrLinkSharesFromIDs(s, userIDs)
		if err != nil {
			return err
		}
	}

	// Add all objects to their tasks
	for _, task := range taskMap {
		if createdBy, has := users[task.CreatedByID]; has {
			task.CreatedBy = createdBy
		} else if fallbackID, ok := legacyCreators[task.ID]; ok {
			if fallbackUser, hasUser := users[fallbackID]; hasUser {
				task.CreatedBy = fallbackUser
				task.CreatedByID = fallbackID
			}
		}

		if remindersList := taskReminders[task.ID]; remindersList != nil {
			task.Reminders = remindersList
		}

		if project, exists := projects[task.ProjectID]; exists && project != nil {
			if project.Identifier == "" {
				task.Identifier = "#" + strconv.FormatInt(task.Index, 10)
			} else {
				task.Identifier = project.Identifier + "-" + strconv.FormatInt(task.Index, 10)
			}
		}

		if taskFavorites != nil {
			task.IsFavorite = taskFavorites[task.ID]
		}
	}

	// Handle expansion parameters using proper service layer methods
	if expand != nil && len(expand) > 0 {
		for _, expandable := range expand {
			switch expandable {
			case models.TaskCollectionExpandBuckets:
				err = ts.addBucketsToTasks(s, a, taskIDs, taskMap)
				if err != nil {
					return err
				}
			case models.TaskCollectionExpandReactions:
				err = ts.addReactionsToTasks(s, taskIDs, taskMap)
				if err != nil {
					return err
				}
			case models.TaskCollectionExpandComments:
				err = ts.addCommentsToTasks(s, taskIDs, taskMap)
				if err != nil {
					return err
				}
			}
		}
	}

	// Add related tasks
	err = ts.addRelatedTasksToTasks(s, taskIDs, taskMap, a)
	if err != nil {
		return err
	}

	// Normalize slice fields to empty arrays so the frontend can safely iterate without null checks.
	for _, task := range taskMap {
		if task.Assignees == nil {
			task.Assignees = []*user.User{}
		}
		if task.Labels == nil {
			task.Labels = []*models.Label{}
		}
		if task.Attachments == nil {
			task.Attachments = []*models.TaskAttachment{}
		}
		if task.Reminders == nil {
			task.Reminders = []*models.TaskReminder{}
		}
		if task.Comments == nil {
			task.Comments = []*models.TaskComment{}
		}
		if task.RelatedTasks == nil {
			task.RelatedTasks = make(models.RelatedTaskMap)
		}
		if task.Buckets == nil {
			task.Buckets = []*models.Bucket{}
		}
		if task.Reactions == nil {
			task.Reactions = models.ReactionMap{}
		}
	}

	return nil
}

// Helper methods moved from models package

func (ts *TaskService) addAssigneesToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	// Initialize empty Assignees slice for all tasks to ensure JSON returns [] instead of null
	for _, task := range taskMap {
		if task.Assignees == nil {
			task.Assignees = []*user.User{}
		}
	}

	taskAssignees := []*models.TaskAssigneeWithUser{}
	err := s.Table("task_assignees").
		Select("task_id, users.*").
		In("task_id", taskIDs).
		Join("INNER", "users", "task_assignees.user_id = users.id").
		Find(&taskAssignees)
	if err != nil {
		return err
	}

	// Put the assignees in the task map
	for i, a := range taskAssignees {
		if a != nil {
			a.Email = "" // Obfuscate the email

			// Check if assignee already exists to avoid duplicates
			alreadyExists := false
			for _, existingAssignee := range taskMap[a.TaskID].Assignees {
				if existingAssignee.ID == taskAssignees[i].ID {
					alreadyExists = true
					break
				}
			}

			if !alreadyExists {
				taskMap[a.TaskID].Assignees = append(taskMap[a.TaskID].Assignees, &taskAssignees[i].User)
			}
		}
	}

	return nil
}

func (ts *TaskService) addLabelsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	// Initialize empty Labels slice for all tasks to ensure JSON returns [] instead of null
	for _, task := range taskMap {
		if task.Labels == nil {
			task.Labels = []*models.Label{}
		}
	}

	labelService := NewLabelService(ts.DB)
	labels, _, _, err := labelService.GetLabelsByTaskIDs(s, &GetLabelsByTaskIDsOptions{
		TaskIDs: taskIDs,
		Page:    -1,
	})
	if err != nil {
		return err
	}

	for i, l := range labels {
		if l != nil {
			// Check if this label is already in the task's Labels slice
			alreadyExists := false
			if taskMap[l.TaskID].Labels != nil {
				for _, existingLabel := range taskMap[l.TaskID].Labels {
					if existingLabel.ID == l.ID {
						alreadyExists = true
						break
					}
				}
			}

			if !alreadyExists {
				taskMap[l.TaskID].Labels = append(taskMap[l.TaskID].Labels, &labels[i].Label)
			}
		}
	}

	return nil
}

func (ts *TaskService) addAttachmentsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	// Initialize empty Attachments slice for all tasks to ensure JSON returns [] instead of null
	for _, task := range taskMap {
		if task.Attachments == nil {
			task.Attachments = []*models.TaskAttachment{}
		}
	}

	attachments, err := ts.getTaskAttachmentsByTaskIDs(s, taskIDs)
	if err != nil {
		return err
	}

	for _, a := range attachments {
		// Check if attachment already exists to avoid duplicates
		alreadyExists := false
		for _, existingAttachment := range taskMap[a.TaskID].Attachments {
			if existingAttachment.ID == a.ID {
				alreadyExists = true
				break
			}
		}

		if !alreadyExists {
			taskMap[a.TaskID].Attachments = append(taskMap[a.TaskID].Attachments, a)
		}
	}

	return nil
}

func (ts *TaskService) getTaskReminderMap(s *xorm.Session, taskIDs []int64) (map[int64][]*models.TaskReminder, error) {
	reminders := []*models.TaskReminder{}
	err := s.In("task_id", taskIDs).
		OrderBy("reminder asc").
		Find(&reminders)
	if err != nil {
		return nil, err
	}

	reminderMap := make(map[int64][]*models.TaskReminder)
	for _, reminder := range reminders {
		reminderMap[reminder.TaskID] = append(reminderMap[reminder.TaskID], reminder)
	}

	return reminderMap, nil
}

func (ts *TaskService) getFavorites(s *xorm.Session, entityIDs []int64, a web.Auth, kind models.FavoriteKind) (map[int64]bool, error) {
	favorites := make(map[int64]bool)
	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return favorites, nil
	}

	favs := []*models.Favorite{}
	err = s.Where(builder.And(
		builder.Eq{"user_id": u.ID},
		builder.Eq{"kind": kind},
		builder.In("entity_id", entityIDs),
	)).
		Find(&favs)

	for _, fav := range favs {
		favorites[fav.EntityID] = true
	}
	return favorites, err
}

func (ts *TaskService) addRelatedTasksToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task, a web.Auth) error {
	relatedTasks := []*models.TaskRelation{}
	err := s.In("task_id", taskIDs).Find(&relatedTasks)
	if err != nil {
		return err
	}

	// Collect all related task IDs, so we can get all related task headers in one go
	var relatedTaskIDs []int64
	for _, rt := range relatedTasks {
		relatedTaskIDs = append(relatedTaskIDs, rt.OtherTaskID)
	}

	if len(relatedTaskIDs) == 0 {
		return nil
	}

	fullRelatedTasks := make(map[int64]*models.Task)
	err = s.In("id", relatedTaskIDs).Find(&fullRelatedTasks)
	if err != nil {
		return err
	}

	taskFavorites, err := ts.getFavorites(s, relatedTaskIDs, a, models.FavoriteKindTask)
	if err != nil {
		return err
	}

	// Go through all task relations and put them into the task objects
	for _, rt := range relatedTasks {
		_, has := fullRelatedTasks[rt.OtherTaskID]
		if !has {
			continue
		}
		fullRelatedTasks[rt.OtherTaskID].IsFavorite = taskFavorites[rt.OtherTaskID]

		// We're duplicating the other task to avoid cycles as these can't be represented properly in json
		// and would thus fail with an error.
		otherTask := &models.Task{}
		err = copier.Copy(otherTask, fullRelatedTasks[rt.OtherTaskID])
		if err != nil {
			continue
		}
		// Clear RelatedTasks map to prevent cycles and match null behavior in JSON
		otherTask.RelatedTasks = nil
		// Note: Other slice/map fields stay nil to match original behavior
		taskMap[rt.TaskID].RelatedTasks[rt.RelationKind] = append(taskMap[rt.TaskID].RelatedTasks[rt.RelationKind], otherTask)
	}

	return nil
}

func (ts *TaskService) canWriteTask(s *xorm.Session, taskID int64, u *user.User) (bool, error) {
	project, err := models.GetProjectSimpleByTaskID(s, taskID)
	if err != nil {
		if models.IsErrProjectDoesNotExist(err) {
			return false, nil
		}
		return false, err
	}

	// Check project permissions using ProjectService
	projectService := NewProjectService(ts.DB)
	return projectService.HasPermission(s, project.ID, u, models.PermissionWrite)
}

// canWriteTaskWithAuth checks if the auth object (User or LinkSharing) can write to a task
// This version accepts web.Auth to support both regular users and LinkSharing authentication
func (ts *TaskService) canWriteTaskWithAuth(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	taskPtr, err := ts.GetByIDSimple(s, taskID)
	if err != nil {
		if models.IsErrTaskDoesNotExist(err) {
			return false, nil
		}
		return false, err
	}
	task := *taskPtr

	// Use the model's CanWrite which properly handles both User and LinkSharing auth
	taskForPermCheck := &models.Task{ID: taskID, ProjectID: task.ProjectID}
	return taskForPermCheck.CanWrite(s, a)
}

// getTaskAttachmentsByTaskIDs gets task attachments with full details
func (ts *TaskService) getTaskAttachmentsByTaskIDs(s *xorm.Session, taskIDs []int64) (attachments []*models.TaskAttachment, err error) {
	attachments = []*models.TaskAttachment{}
	err = s.
		In("task_id", taskIDs).
		Find(&attachments)
	if err != nil {
		return
	}

	if len(attachments) == 0 {
		return
	}

	fileIDs := []int64{}
	userIDs := []int64{}
	for _, a := range attachments {
		userIDs = append(userIDs, a.CreatedByID)
		fileIDs = append(fileIDs, a.FileID)
	}

	// Get all files
	fs := make(map[int64]*files.File)
	err = s.In("id", fileIDs).Find(&fs)
	if err != nil {
		return
	}

	users, err := ts.getUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return nil, err
	}

	// Obfuscate all user emails
	for _, u := range users {
		u.Email = ""
	}

	for _, a := range attachments {
		if createdBy, has := users[a.CreatedByID]; has {
			a.CreatedBy = createdBy
		}
		a.File = fs[a.FileID]
	}

	return
}

// updateProjectLastUpdated updates the last updated timestamp of a project
func (ts *TaskService) updateProjectLastUpdated(s *xorm.Session, projectID int64) error {
	project := &models.Project{
		ID:      projectID,
		Updated: time.Now(),
	}
	_, err := s.ID(projectID).Cols("updated").Update(project)
	return err
}

// getUsersOrLinkSharesFromIDs gets users and link shares from their IDs.
func (ts *TaskService) getUsersOrLinkSharesFromIDs(s *xorm.Session, ids []int64) (users map[int64]*user.User, err error) {
	users = make(map[int64]*user.User)
	var userIDs []int64
	var linkShareIDs []int64
	for _, id := range ids {
		if id < 0 {
			linkShareIDs = append(linkShareIDs, id*-1)
			continue
		}

		userIDs = append(userIDs, id)
	}

	if len(userIDs) > 0 {
		users, err = user.GetUsersByIDs(s, userIDs)
		if err != nil {
			return
		}
	}

	if len(linkShareIDs) == 0 {
		return
	}

	shares, err := ts.Registry.LinkShare().GetByIDs(s, linkShareIDs)
	if err != nil {
		return nil, err
	}

	for _, share := range shares {
		users[share.ID*-1] = ts.toUser(share)
	}

	return
}

func (ts *TaskService) toUser(share *models.LinkSharing) *user.User {
	suffix := "Link Share"
	if share.Name != "" {
		suffix = " (" + suffix + ")"
	}

	username := "link-share-" + strconv.FormatInt(share.ID, 10)

	return &user.User{
		ID:       ts.getUserID(share),
		Name:     share.Name + suffix,
		Username: username,
		Created:  share.Created,
		Updated:  share.Updated,
	}
}

func (ts *TaskService) getUserID(share *models.LinkSharing) int64 {
	return share.ID * -1
}

type taskCreationOptions struct {
	skipPermissionCheck bool
	updateAssignees     bool
	setBucket           bool
}

// Create creates a new task with permission checks and full service-layer business logic.
func (ts *TaskService) Create(s *xorm.Session, task *models.Task, u *user.User) (*models.Task, error) {
	return ts.CreateWithOptions(s, task, u, true, true, false)
}

// CreateWithoutPermissionCheck creates a new task without performing permission checks.
// This is intended for internal use where permissions have already been validated externally.
func (ts *TaskService) CreateWithoutPermissionCheck(s *xorm.Session, task *models.Task, u *user.User) (*models.Task, error) {
	return ts.CreateWithOptions(s, task, u, true, true, true)
}

// CreateWithOptions provides fine-grained control over task creation behavior while reusing
// the core service-layer implementation. Callers can disable assignee updates or bucket placement
// when duplicating tasks or performing specialized operations.
func (ts *TaskService) CreateWithOptions(s *xorm.Session, task *models.Task, u *user.User, updateAssignees bool, setBucket bool, skipPermissionCheck bool) (*models.Task, error) {
	opts := taskCreationOptions{
		skipPermissionCheck: skipPermissionCheck,
		updateAssignees:     updateAssignees,
		setBucket:           setBucket,
	}
	return ts.createTask(s, task, u, opts)
}

// createTask contains the core business logic for task creation.
func (ts *TaskService) createTask(s *xorm.Session, task *models.Task, actor *user.User, opts taskCreationOptions) (*models.Task, error) {
	if task == nil {
		return nil, fmt.Errorf("task must not be nil")
	}
	if actor == nil {
		return nil, ErrAccessDenied
	}

	if task.Title == "" {
		return nil, models.ErrTaskCannotBeEmpty{}
	}

	project, err := ts.Registry.Project().GetByIDSimple(s, task.ProjectID)
	if err != nil {
		return nil, err
	}

	if !opts.skipPermissionCheck {
		projectService := NewProjectService(ts.DB)
		canWrite, err := projectService.HasPermission(s, task.ProjectID, actor, models.PermissionWrite)
		if err != nil {
			return nil, fmt.Errorf("checking project write permission: %w", err)
		}
		if !canWrite {
			return nil, ErrAccessDenied
		}
	}

	createdBy, err := models.GetUserOrLinkShareUser(s, actor)
	if err != nil {
		return nil, err
	}
	task.CreatedByID = createdBy.ID
	task.CreatedBy = createdBy

	if task.UID == "" {
		task.UID = uuid.NewString()
	}

	if err := ts.ensureTaskIndex(s, task); err != nil {
		return nil, err
	}

	task.HexColor = utils.NormalizeHex(task.HexColor)

	if _, err := s.Insert(task); err != nil {
		return nil, err
	}

	var providedBucket *models.Bucket
	if opts.setBucket && task.BucketID != 0 {
		providedBucket, err = ts.Registry.Kanban().GetBucketByID(s, task.BucketID)
		if err != nil {
			return nil, err
		}
		if _, err = ts.Registry.Kanban().checkBucketLimit(s, createdBy, task, providedBucket); err != nil {
			return nil, err
		}
	}

	if opts.setBucket {
		if err := ts.assignTaskToViews(s, task, createdBy, providedBucket); err != nil {
			return nil, err
		}
	}

	if opts.updateAssignees {
		if err := ts.syncTaskAssignees(s, task, task.Assignees, createdBy); err != nil {
			return nil, err
		}
	}

	// Sync labels if provided
	if len(task.Labels) > 0 {
		if err := ts.syncTaskLabels(s, task, task.Labels, createdBy); err != nil {
			return nil, err
		}
	}

	if err := ts.syncTaskReminders(s, task); err != nil {
		return nil, err
	}

	ts.setTaskIdentifier(task, project)

	if task.IsFavorite {
		if err := ts.Registry.Favorite().AddToFavorite(s, task.ID, createdBy, models.FavoriteKindTask); err != nil {
			return nil, err
		}
	}

	if err := events.Dispatch(&models.TaskCreatedEvent{Task: task, Doer: createdBy}); err != nil {
		return nil, err
	}

	if err := ts.updateProjectLastUpdated(s, task.ProjectID); err != nil {
		return nil, err
	}

	return task, nil
}

func (ts *TaskService) assignTaskToViews(s *xorm.Session, task *models.Task, auth web.Auth, providedBucket *models.Bucket) error {
	views, err := ts.getViewsForProject(s, task.ProjectID)
	if err != nil {
		return err
	}

	positions := make([]*models.TaskPosition, 0, len(views))
	taskBuckets := make([]*models.TaskBucket, 0, len(views))
	moveToDone := false

	for _, view := range views {
		if view.ViewKind == models.ProjectViewKindKanban && view.BucketConfigurationMode == models.BucketConfigurationModeManual && !moveToDone {
			bucketID := view.DoneBucketID
			if !task.Done || view.DoneBucketID == 0 {
				if providedBucket != nil && view.ID == providedBucket.ProjectViewID {
					bucketID = providedBucket.ID
				} else {
					bucketID, err = ts.Registry.Kanban().getDefaultBucketID(s, view)
					if err != nil {
						return err
					}
				}
			}

			if view.DoneBucketID != 0 && view.DoneBucketID == task.BucketID && !task.Done {
				task.Done = true
				if _, err = s.Where("id = ?", task.ID).Cols("done").Update(task); err != nil {
					return err
				}

				if err = ts.moveTaskToDoneBuckets(s, task, auth, views); err != nil {
					return err
				}

				moveToDone = true
				continue
			}

			taskBuckets = append(taskBuckets, &models.TaskBucket{
				BucketID:      bucketID,
				TaskID:        task.ID,
				ProjectViewID: view.ID,
				ProjectID:     task.ProjectID,
			})
		}

		position, err := ts.calculateNewPositionForTask(s, auth, task, view)
		if err != nil {
			return err
		}
		positions = append(positions, position)
	}

	if moveToDone {
		taskBuckets = []*models.TaskBucket{}
	}

	if len(positions) > 0 {
		if _, err = s.Insert(&positions); err != nil {
			return err
		}
	}

	if len(taskBuckets) > 0 {
		if _, err = s.Insert(&taskBuckets); err != nil {
			return err
		}
	}

	return nil
}

func (ts *TaskService) getViewsForProject(s *xorm.Session, projectID int64) ([]*models.ProjectView, error) {
	views := make([]*models.ProjectView, 0)
	err := s.Where("project_id = ?", projectID).OrderBy("position asc").Find(&views)
	return views, err
}

func (ts *TaskService) calculateNewPositionForTask(s *xorm.Session, auth web.Auth, task *models.Task, view *models.ProjectView) (*models.TaskPosition, error) {
	if task.Position == 0 {
		lowestPosition := &models.TaskPosition{}
		exists, err := s.Where("project_view_id = ?", view.ID).OrderBy("position asc").Get(lowestPosition)
		if err != nil {
			return nil, err
		}
		if exists {
			if lowestPosition.Position == 0 {
				if err = models.RecalculateTaskPositions(s, view, auth); err != nil {
					return nil, err
				}

				lowestPosition = &models.TaskPosition{}
				if _, err = s.Where("project_view_id = ?", view.ID).OrderBy("position asc").Get(lowestPosition); err != nil {
					return nil, err
				}
			}

			task.Position = lowestPosition.Position / 2
		}
	}

	return &models.TaskPosition{
		TaskID:        task.ID,
		ProjectViewID: view.ID,
		Position:      ts.calculateDefaultPosition(task.Index, task.Position),
	}, nil
}

func (ts *TaskService) calculateDefaultPosition(entityID int64, position float64) float64 {
	if position == 0 {
		return float64(entityID) * 1000
	}
	return position
}

func (ts *TaskService) moveTaskToDoneBuckets(s *xorm.Session, task *models.Task, auth web.Auth, views []*models.ProjectView) error {
	for _, view := range views {
		currentTaskBucket := &models.TaskBucket{}
		if _, err := s.Where("task_id = ? AND project_view_id = ?", task.ID, view.ID).Get(currentTaskBucket); err != nil {
			return err
		}

		bucketID := currentTaskBucket.BucketID

		if task.Done && view.DoneBucketID == 0 {
			continue
		}

		if !task.Done && bucketID != view.DoneBucketID {
			continue
		}

		if task.Done && view.DoneBucketID != 0 {
			bucketID = view.DoneBucketID
		}

		if !task.Done && bucketID == view.DoneBucketID {
			var err error
			bucketID, err = ts.Registry.Kanban().getDefaultBucketID(s, view)
			if err != nil {
				return err
			}
		}

		tb := &models.TaskBucket{
			BucketID:      bucketID,
			TaskID:        task.ID,
			ProjectViewID: view.ID,
			ProjectID:     task.ProjectID,
		}
		if err := tb.Update(s, auth); err != nil {
			return err
		}

		tp := models.TaskPosition{
			TaskID:        task.ID,
			ProjectViewID: view.ID,
			Position:      ts.calculateDefaultPosition(task.Index, task.Position),
		}
		if err := tp.Update(s, auth); err != nil {
			return err
		}
	}
	return nil
}

func (ts *TaskService) syncTaskAssignees(s *xorm.Session, task *models.Task, desiredAssignees []*user.User, createdBy web.Auth) error {
	currentAssignees, err := ts.getRawTaskAssigneesForTask(s, task.ID)
	if err != nil {
		return err
	}

	currentAssigneeMap := make(map[int64]struct{}, len(currentAssignees))
	for _, entry := range currentAssignees {
		currentAssigneeMap[entry.ID] = struct{}{}
	}

	desiredAssigneeMap := make(map[int64]*user.User)
	for _, assignee := range desiredAssignees {
		if assignee == nil || assignee.ID == 0 {
			continue
		}
		if _, exists := desiredAssigneeMap[assignee.ID]; !exists {
			desiredAssigneeMap[assignee.ID] = assignee
		}
	}

	// Delete assignees that are no longer desired
	assigneesToDelete := make([]int64, 0)
	for id := range currentAssigneeMap {
		if _, keep := desiredAssigneeMap[id]; !keep {
			assigneesToDelete = append(assigneesToDelete, id)
		}
	}

	if len(assigneesToDelete) > 0 {
		if _, err = s.In("user_id", assigneesToDelete).And("task_id = ?", task.ID).Delete(&models.TaskAssginee{}); err != nil {
			return err
		}
	}

	// Add new assignees
	for id := range desiredAssigneeMap {
		if _, already := currentAssigneeMap[id]; already {
			continue
		}

		assignee := &models.TaskAssginee{TaskID: task.ID, UserID: id}
		if err := assignee.Create(s, createdBy); err != nil {
			if !models.IsErrUserAlreadyAssigned(err) {
				return err
			}
		}
	}

	// Refresh assignee list on the task to include full user data
	taskMap := map[int64]*models.Task{task.ID: task}
	if err := ts.addAssigneesToTasks(s, []int64{task.ID}, taskMap); err != nil {
		return err
	}

	if len(task.Assignees) == 0 {
		task.Assignees = nil
	}

	return ts.updateProjectLastUpdated(s, task.ProjectID)
}

// syncTaskLabels synchronizes the task's labels with the desired state.
// It adds new labels and removes labels that are no longer desired.
func (ts *TaskService) syncTaskLabels(s *xorm.Session, task *models.Task, desiredLabels []*models.Label, createdBy web.Auth) error {
	// Get current labels
	currentLabelTasks := make([]*models.LabelTask, 0)
	err := s.Where("task_id = ?", task.ID).Find(&currentLabelTasks)
	if err != nil {
		return err
	}

	currentLabelMap := make(map[int64]struct{}, len(currentLabelTasks))
	for _, lt := range currentLabelTasks {
		currentLabelMap[lt.LabelID] = struct{}{}
	}

	desiredLabelMap := make(map[int64]*models.Label)
	for _, label := range desiredLabels {
		if label == nil || label.ID == 0 {
			continue
		}
		if _, exists := desiredLabelMap[label.ID]; !exists {
			desiredLabelMap[label.ID] = label
		}
	}

	// Delete labels that are no longer desired
	labelsToDelete := make([]int64, 0)
	for id := range currentLabelMap {
		if _, keep := desiredLabelMap[id]; !keep {
			labelsToDelete = append(labelsToDelete, id)
		}
	}

	if len(labelsToDelete) > 0 {
		if _, err = s.In("label_id", labelsToDelete).And("task_id = ?", task.ID).Delete(&models.LabelTask{}); err != nil {
			return err
		}
	}

	// Add new labels
	for id := range desiredLabelMap {
		if _, already := currentLabelMap[id]; already {
			continue
		}

		// Validate that the user has access to the label using LabelService
		hasAccess, err := ts.Registry.Label().HasAccessToLabel(s, id, createdBy)
		if err != nil {
			return err
		}
		if !hasAccess {
			u, _ := createdBy.(*user.User)
			if u != nil {
				return models.ErrUserHasNoAccessToLabel{LabelID: id, UserID: u.ID}
			}
			return ErrAccessDenied
		}

		// Insert the label-task association
		_, err = s.Insert(&models.LabelTask{
			LabelID: id,
			TaskID:  task.ID,
		})
		if err != nil {
			return err
		}
	}

	// Refresh label list on the task to include full label data
	taskMap := map[int64]*models.Task{task.ID: task}
	if err := ts.addLabelsToTasks(s, []int64{task.ID}, taskMap); err != nil {
		return err
	}

	if len(task.Labels) == 0 {
		task.Labels = nil
	}

	return ts.updateProjectLastUpdated(s, task.ProjectID)
}

func (ts *TaskService) getRawTaskAssigneesForTask(s *xorm.Session, taskID int64) ([]*models.TaskAssigneeWithUser, error) {
	assignees := make([]*models.TaskAssigneeWithUser, 0)
	err := s.Table("task_assignees").
		Select("task_assignees.task_id, users.*").
		Join("INNER", "users", "task_assignees.user_id = users.id").
		Where("task_assignees.task_id = ?", taskID).
		Find(&assignees)
	return assignees, err
}

func (ts *TaskService) syncTaskReminders(s *xorm.Session, task *models.Task) error {
	if _, err := s.Where("task_id = ?", task.ID).Delete(&models.TaskReminder{}); err != nil {
		return err
	}

	if err := ts.normalizeRelativeReminderDates(task); err != nil {
		return err
	}

	reminderMap := make(map[int64]*models.TaskReminder, len(task.Reminders))
	for _, reminder := range task.Reminders {
		reminderMap[reminder.Reminder.UTC().Unix()] = reminder
	}

	task.Reminders = make([]*models.TaskReminder, 0, len(reminderMap))
	for _, reminder := range reminderMap {
		entry := &models.TaskReminder{
			TaskID:         task.ID,
			Reminder:       reminder.Reminder,
			RelativePeriod: reminder.RelativePeriod,
			RelativeTo:     reminder.RelativeTo,
		}
		if _, err := s.Insert(entry); err != nil {
			return err
		}
		task.Reminders = append(task.Reminders, entry)
	}

	sort.Slice(task.Reminders, func(i, j int) bool {
		return task.Reminders[i].Reminder.Before(task.Reminders[j].Reminder)
	})

	if len(task.Reminders) == 0 {
		task.Reminders = nil
	}

	return ts.updateProjectLastUpdated(s, task.ProjectID)
}

func (ts *TaskService) normalizeRelativeReminderDates(task *models.Task) error {
	for _, reminder := range task.Reminders {
		relativeDuration := time.Duration(reminder.RelativePeriod) * time.Second
		if reminder.RelativeTo != "" {
			reminder.Reminder = time.Time{}
		}

		switch reminder.RelativeTo {
		case models.ReminderRelationDueDate:
			if !task.DueDate.IsZero() {
				reminder.Reminder = task.DueDate.Add(relativeDuration)
			}
		case models.ReminderRelationStartDate:
			if !task.StartDate.IsZero() {
				reminder.Reminder = task.StartDate.Add(relativeDuration)
			}
		case models.ReminderRelationEndDate:
			if !task.EndDate.IsZero() {
				reminder.Reminder = task.EndDate.Add(relativeDuration)
			}
		default:
			if reminder.RelativePeriod != 0 {
				return models.ErrReminderRelativeToMissing{TaskID: task.ID}
			}
		}
	}
	return nil
}

func (ts *TaskService) ensureTaskIndex(s *xorm.Session, task *models.Task) error {
	if task.Index == 0 {
		nextIndex, err := ts.calculateNextTaskIndex(s, task.ProjectID)
		if err != nil {
			return err
		}
		task.Index = nextIndex
		return nil
	}

	exists, err := s.Where("project_id = ? AND `index` = ?", task.ProjectID, task.Index).Exist(&models.Task{})
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	nextIndex, err := ts.calculateNextTaskIndex(s, task.ProjectID)
	if err != nil {
		return err
	}
	task.Index = nextIndex
	return nil
}

func (ts *TaskService) calculateNextTaskIndex(s *xorm.Session, projectID int64) (int64, error) {
	latestTask := &models.Task{}
	_, err := s.Where("project_id = ?", projectID).OrderBy("`index` desc").Get(latestTask)
	if err != nil {
		return 0, err
	}

	return latestTask.Index + 1, nil
}

func (ts *TaskService) setTaskIdentifier(task *models.Task, project *models.Project) {
	if project == nil || project.Identifier == "" {
		task.Identifier = "#" + strconv.FormatInt(task.Index, 10)
		return
	}

	task.Identifier = project.Identifier + "-" + strconv.FormatInt(task.Index, 10)
}

// getRawFavoriteTasks gets favorite tasks with filtering and sorting
func (ts *TaskService) getRawFavoriteTasks(s *xorm.Session, favoriteTaskIDs []int64, opts *taskSearchOptions) (tasks []*models.Task, resultCount int, totalItems int64, err error) {
	if len(favoriteTaskIDs) == 0 {
		return nil, 0, 0, nil
	}

	// Create a copy of opts for favorites
	favoriteOpts := *opts
	favoriteOpts.projectIDs = nil // Clear project IDs for favorites

	// Build the query using favorite task IDs
	query := s.In("id", favoriteTaskIDs)

	// Apply filters, sorting, and search
	query, _, err = ts.applyFiltersToQuery(query, &favoriteOpts)
	if err != nil {
		return nil, 0, 0, err
	}

	// Apply sorting
	query = ts.applySortingToQuery(query, favoriteOpts.sortby)

	// Get total count first (before pagination)
	totalItems, err = s.In("id", favoriteTaskIDs).Count(&models.Task{})
	if err != nil {
		return nil, 0, 0, err
	}

	// Apply pagination
	query = query.Limit(opts.perPage, (opts.page-1)*opts.perPage)

	// Execute query
	err = query.Find(&tasks)
	if err != nil {
		return nil, 0, 0, err
	}

	return tasks, len(tasks), totalItems, nil
}

// buildAndExecuteTaskQuery builds and executes the main task query with all filters
func (ts *TaskService) buildAndExecuteTaskQuery(s *xorm.Session, opts *taskSearchOptions) (tasks []*models.Task, resultCount int, totalItems int64, err error) {
	// Start with project filtering
	query := s.In("project_id", opts.projectIDs)

	// Apply filters, sorting, and search
	query, _, err = ts.applyFiltersToQuery(query, opts)
	if err != nil {
		return nil, 0, 0, err
	}

	// Apply sorting
	query = ts.applySortingToQuery(query, opts.sortby)

	// Get total count first (before pagination)
	totalItems, err = s.In("project_id", opts.projectIDs).Count(&models.Task{})
	if err != nil {
		return nil, 0, 0, err
	}

	// Apply pagination
	query = query.Limit(opts.perPage, (opts.page-1)*opts.perPage)

	// Execute query
	err = query.Find(&tasks)
	if err != nil {
		return nil, 0, 0, err
	}

	return tasks, len(tasks), totalItems, nil
}

// applyFiltersToQuery applies all filters to the query
func (ts *TaskService) applyFiltersToQuery(query *xorm.Session, opts *taskSearchOptions) (*xorm.Session, *xorm.Session, error) {
	// For now, delegate complex filtering to the model
	// TODO: Move all filter logic to service layer

	// Apply search filter
	if opts.search != "" {
		searchWhere := "title LIKE ?"
		searchPattern := "%" + opts.search + "%"
		query = query.Where(searchWhere, searchPattern)
	}

	// Apply custom filters if present
	if opts.parsedFilters != nil && len(opts.parsedFilters) > 0 {
		filterCond, err := ts.convertFiltersToDBFilterCond(opts.parsedFilters, opts.filterIncludeNulls)
		if err != nil {
			return nil, nil, err
		}
		query = query.And(filterCond)
	}

	// Use the same query for count (xorm doesn't have Clone)
	totalQuery := query
	return query, totalQuery, nil
}

// applySortingToQuery applies sorting to the query
func (ts *TaskService) applySortingToQuery(query *xorm.Session, sortParams []*sortParam) *xorm.Session {
	for _, param := range sortParams {
		var orderBy string
		if param.orderBy == orderDescending {
			orderBy = param.sortBy + " DESC"
		} else {
			orderBy = param.sortBy + " ASC"
		}
		query = query.OrderBy(orderBy)
	}
	return query
}

// addBucketsToTasks adds bucket information to tasks using the KanbanService
func (ts *TaskService) addBucketsToTasks(s *xorm.Session, a web.Auth, taskIDs []int64, taskMap map[int64]*models.Task) error {
	u, err := models.GetUserOrLinkShareUser(s, a)
	if err != nil {
		return err
	}

	return ts.Registry.Kanban().AddBucketsToTasks(s, taskIDs, taskMap, u)
}

// addReactionsToTasks adds reaction data to tasks using the ReactionsService
func (ts *TaskService) addReactionsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	return ts.Registry.Reactions().AddReactionsToTasks(s, taskIDs, taskMap)
}

// addCommentsToTasks adds comment data to tasks using the CommentService
func (ts *TaskService) addCommentsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	fmt.Printf("DEBUG: addCommentsToTasks called with taskIDs: %v\n", taskIDs)
	fmt.Printf("DEBUG: Calling CommentService.AddCommentsToTasks\n")
	return ts.Registry.Comment().AddCommentsToTasks(s, taskIDs, taskMap)
}

// ===== Permission Methods =====
// Migrated from models layer as part of T-PERM-007
// All task permissions delegate to project permissions since tasks belong to projects

// CanRead checks if a user has read access to a task.
// Returns (canRead, maxPermissionLevel, error).
// MIGRATION: Migrated from models.Task.CanRead (T-PERM-007)
func (ts *TaskService) CanRead(s *xorm.Session, taskID int64, a web.Auth) (bool, int, error) {
	// Get task to find project
	task, err := ts.GetByIDSimple(s, taskID)
	if err != nil {
		return false, 0, err
	}

	// Delegate to ProjectService
	ps := NewProjectService(s.Engine())
	return ps.CanRead(s, task.ProjectID, a)
}

// CanWrite checks if a user has write access to a task.
// MIGRATION: Migrated from models.Task.CanWrite (T-PERM-007)
func (ts *TaskService) CanWrite(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	return ts.canDoTask(s, taskID, nil, a)
}

// CanUpdate checks if a user can update a task.
// MIGRATION: Migrated from models.Task.CanUpdate (T-PERM-007)
func (ts *TaskService) CanUpdate(s *xorm.Session, taskID int64, task *models.Task, a web.Auth) (bool, error) {
	return ts.canDoTask(s, taskID, task, a)
}

// CanDelete checks if a user can delete a task.
// MIGRATION: Migrated from models.Task.CanDelete (T-PERM-007)
func (ts *TaskService) CanDelete(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	return ts.canDoTask(s, taskID, nil, a)
}

// CanCreate checks if a user can create a task in the specified project.
// MIGRATION: Migrated from models.Task.CanCreate (T-PERM-007)
func (ts *TaskService) CanCreate(s *xorm.Session, task *models.Task, a web.Auth) (bool, error) {
	// A user can create a task if they have write access to its project
	ps := NewProjectService(s.Engine())
	return ps.CanWrite(s, task.ProjectID, a)
}

// canDoTask is a helper function to check if a user can do operations on a task.
// It handles the case where a task is being moved to a different project.
// MIGRATION: Migrated from models.Task.canDoTask (T-PERM-007)
func (ts *TaskService) canDoTask(s *xorm.Session, taskID int64, updatedTask *models.Task, a web.Auth) (bool, error) {
	// Get the original task
	originalTask, err := ts.GetByIDSimple(s, taskID)
	if err != nil {
		return false, err
	}

	ps := NewProjectService(s.Engine())

	// Check if we're moving the task to a different project
	// If so, verify permissions on the new project
	if updatedTask != nil && updatedTask.ProjectID != 0 && updatedTask.ProjectID != originalTask.ProjectID {
		canWriteToNewProject, err := ps.CanWrite(s, updatedTask.ProjectID, a)
		if err != nil {
			return false, err
		}
		if !canWriteToNewProject {
			return false, models.ErrGenericForbidden{}
		}
	}

	// A user can do a task if they have write access to its (original) project
	return ps.CanWrite(s, originalTask.ProjectID, a)
}

// ===== Task Relation Permission Methods =====
// Migrated from models layer as part of T-PERM-010

// CanCreateAssignee checks if a user can add a new assignee to a task.
// MIGRATION: Migrated from models.TaskAssginee.CanCreate (T-PERM-010)
func (ts *TaskService) CanCreateAssignee(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	// User must be able to write to the task
	return ts.CanWrite(s, taskID, a)
}

// CanDeleteAssignee checks if a user can delete an assignee from a task.
// MIGRATION: Migrated from models.TaskAssginee.CanDelete (T-PERM-010)
func (ts *TaskService) CanDeleteAssignee(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	// User must be able to write to the task
	return ts.CanWrite(s, taskID, a)
}

// CanCreateRelation checks if a user can create a new relation between two tasks.
// MIGRATION: Migrated from models.TaskRelation.CanCreate (T-PERM-010)
func (ts *TaskService) CanCreateRelation(s *xorm.Session, taskID int64, otherTaskID int64, relationKind models.RelationKind, a web.Auth) (bool, error) {
	// Check if the relation kind is valid
	if !relationKind.IsValid() {
		return false, models.ErrInvalidRelationKind{Kind: relationKind}
	}

	// Needs write access to the base task
	canWrite, err := ts.CanUpdate(s, taskID, nil, a)
	if err != nil || !canWrite {
		return false, err
	}

	// And at least read access to the other task
	canRead, _, err := ts.CanRead(s, otherTaskID, a)
	if err != nil {
		return false, err
	}
	return canRead, nil
}

// CanDeleteRelation checks if a user can delete a task relation.
// MIGRATION: Migrated from models.TaskRelation.CanDelete (T-PERM-010)
func (ts *TaskService) CanDeleteRelation(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	// A user can delete a relation if they can update the base task
	return ts.CanUpdate(s, taskID, nil, a)
}

// CanUpdatePosition checks if a user can update a task's position.
// MIGRATION: Migrated from models.TaskPosition.CanUpdate (T-PERM-010)
func (ts *TaskService) CanUpdatePosition(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	// User must be able to update the task
	return ts.CanUpdate(s, taskID, nil, a)
}
