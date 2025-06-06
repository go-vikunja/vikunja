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
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"

	"github.com/ganigeorgiev/fexpr"
	"github.com/iancoleman/strcase"
	"github.com/jszwedko/go-datemath"
	"xorm.io/builder"
	"xorm.io/xorm/schemas"
)

type taskFilterComparator string

const (
	taskFilterComparatorInvalid taskFilterComparator = "invalid"

	taskFilterComparatorEquals       taskFilterComparator = "="
	taskFilterComparatorGreater      taskFilterComparator = ">"
	taskFilterComparatorGreateEquals taskFilterComparator = ">="
	taskFilterComparatorLess         taskFilterComparator = "<"
	taskFilterComparatorLessEquals   taskFilterComparator = "<="
	taskFilterComparatorNotEquals    taskFilterComparator = "!="
	taskFilterComparatorLike         taskFilterComparator = "like"
	taskFilterComparatorIn           taskFilterComparator = "in"
	taskFilterComparatorNotIn        taskFilterComparator = "not in"
)

// Guess what you get back if you ask Safari for a rfc 3339 formatted date?
const safariDateAndTime = "2006-01-02 15:04"
const safariDate = "2006-01-02"

type taskFilter struct {
	field      string
	value      interface{} // Needs to be an interface to be able to hold the field's native value
	comparator taskFilterComparator
	isNumeric  bool
	join       taskFilterConcatinator
}

func parseTimeFromUserInput(timeString string, loc *time.Location) (value time.Time, err error) {
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

func parseFilterFromExpression(f fexpr.ExprGroup, loc *time.Location) (filter *taskFilter, err error) {
	filter = &taskFilter{
		join: filterConcatAnd,
	}
	if f.Join == fexpr.JoinOr {
		filter.join = filterConcatOr
	}

	var value string
	switch v := f.Item.(type) {
	case fexpr.Expr:
		filter.field = v.Left.Literal
		value = v.Right.Literal
		filter.comparator, err = getFilterComparatorFromOp(v.Op)
		if err != nil {
			return
		}
	case []fexpr.ExprGroup:
		values := make([]*taskFilter, 0, len(v))
		for _, expression := range v {
			subfilter, err := parseFilterFromExpression(expression, loc)
			if err != nil {
				return nil, err
			}
			values = append(values, subfilter)
		}
		filter.value = values
		return
	}

	err = validateTaskFieldComparator(filter.comparator)
	if err != nil {
		return
	}

	// Cast the field value to its native type
	var reflectValue *reflect.StructField
	if filter.field == "project" {
		filter.field = "project_id"
	}

	err = validateTaskField(filter.field)
	if err != nil {
		return nil, err
	}

	reflectValue, filter.value, err = getNativeValueForTaskField(filter.field, filter.comparator, value, loc)
	if err != nil {
		return nil, ErrInvalidTaskFilterValue{
			Field: filter.field,
			Value: value,
		}
	}
	if reflectValue != nil {
		filter.isNumeric = reflectValue.Type.Kind() == reflect.Int64
	}

	return filter, nil
}

func getTaskFiltersFromFilterString(filter string, filterTimezone string) (filters []*taskFilter, err error) {

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
		return nil, &ErrInvalidFilterExpression{
			Expression:      filter,
			ExpressionError: err,
		}
	}

	var loc *time.Location
	if filterTimezone != "" {
		loc, err = time.LoadLocation(filterTimezone)
		if err != nil {
			return nil, &ErrInvalidTimezone{
				Name:      filterTimezone,
				LoadError: err,
			}
		}
	}

	filters = make([]*taskFilter, 0, len(parsedFilter))
	for _, f := range parsedFilter {
		parsedFilter, err := parseFilterFromExpression(f, loc)
		if err != nil {
			return nil, err
		}
		filters = append(filters, parsedFilter)
	}

	return
}

func validateTaskFieldComparator(comparator taskFilterComparator) error {
	switch comparator {
	case
		taskFilterComparatorEquals,
		taskFilterComparatorGreater,
		taskFilterComparatorGreateEquals,
		taskFilterComparatorLess,
		taskFilterComparatorLessEquals,
		taskFilterComparatorNotEquals,
		taskFilterComparatorLike,
		taskFilterComparatorIn,
		taskFilterComparatorNotIn:
		return nil
	case taskFilterComparatorInvalid:
		fallthrough
	default:
		return ErrInvalidTaskFilterComparator{Comparator: comparator}
	}
}

func getFilterComparatorFromOp(op fexpr.SignOp) (taskFilterComparator, error) {
	switch op {
	case fexpr.SignEq:
		return taskFilterComparatorEquals, nil
	case fexpr.SignGt:
		return taskFilterComparatorGreater, nil
	case fexpr.SignGte:
		return taskFilterComparatorGreateEquals, nil
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
		return taskFilterComparatorInvalid, ErrInvalidTaskFilterComparator{Comparator: taskFilterComparator(op)}
	}
}

func getValueForField(field reflect.StructField, rawValue string, loc *time.Location) (value interface{}, err error) {

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
				tt, err = parseTimeFromUserInput(rawValue, loc)
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

func getNativeValueForTaskField(fieldName string, comparator taskFilterComparator, value string, loc *time.Location) (reflectField *reflect.StructField, nativeValue interface{}, err error) {

	realFieldName := strings.ReplaceAll(strcase.ToCamel(fieldName), "Id", "ID")

	if realFieldName == "Assignees" {
		vals := strings.Split(value, ",")
		valueSlice := append([]string{}, vals...)
		return nil, valueSlice, nil
	}

	field, ok := reflect.TypeOf(&Task{}).Elem().FieldByName(realFieldName)
	if !ok {
		return nil, nil, ErrInvalidTaskField{TaskField: fieldName}
	}

	if realFieldName == "Reminders" {
		field, ok = reflect.TypeOf(&TaskReminder{}).Elem().FieldByName("Reminder")
		if !ok {
			return nil, nil, ErrInvalidTaskField{TaskField: fieldName}
		}
	}

	if comparator == taskFilterComparatorIn || comparator == taskFilterComparatorNotIn {
		vals := strings.Split(value, ",")
		valueSlice := []interface{}{}
		for _, val := range vals {
			v, err := getValueForField(field, val, loc)
			if err != nil {
				return nil, nil, err
			}
			valueSlice = append(valueSlice, v)
		}
		return nil, valueSlice, nil
	}

	val, err := getValueForField(field, value, loc)
	return &field, val, err
}
