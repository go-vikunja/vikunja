// Copyright 2020 Vikunja and contriubtors. All rights reserved.
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
	"fmt"
	"reflect"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"github.com/iancoleman/strcase"
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
)

type taskFilter struct {
	field      string
	value      interface{} // Needs to be an interface to be able to hold the field's native value
	comparator taskFilterComparator
}

func getTaskFiltersByCollections(c *TaskCollection) (filters []*taskFilter, err error) {

	if len(c.FilterByArr) > 0 {
		c.FilterBy = append(c.FilterBy, c.FilterByArr...)
	}

	if len(c.FilterValueArr) > 0 {
		c.FilterValue = append(c.FilterValue, c.FilterValueArr...)
	}

	if len(c.FilterComparatorArr) > 0 {
		c.FilterComparator = append(c.FilterComparator, c.FilterComparatorArr...)
	}

	if c.FilterConcat != "" && c.FilterConcat != filterConcatAnd && c.FilterConcat != filterConcatOr {
		return nil, ErrInvalidTaskFilterConcatinator{
			Concatinator: taskFilterConcatinator(c.FilterConcat),
		}
	}

	filters = make([]*taskFilter, 0, len(c.FilterBy))
	for i, f := range c.FilterBy {
		filter := &taskFilter{
			field:      f,
			comparator: taskFilterComparatorEquals,
		}

		if len(c.FilterComparator) > i {
			filter.comparator, err = getFilterComparatorFromString(c.FilterComparator[i])
			if err != nil {
				return
			}
		}

		err = validateTaskFieldComparator(filter.comparator)
		if err != nil {
			return
		}

		// Cast the field value to its native type
		if len(c.FilterValue) > i {
			filter.value, err = getNativeValueForTaskField(filter.field, c.FilterValue[i])
			if err != nil {
				return nil, ErrInvalidTaskFilterValue{
					Value: filter.field,
					Field: c.FilterValue[i],
				}
			}
		}

		filters = append(filters, filter)
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
		taskFilterComparatorLike:
		return nil
	case taskFilterComparatorInvalid:
		fallthrough
	default:
		return ErrInvalidTaskFilterComparator{Comparator: comparator}
	}
}

func getFilterComparatorFromString(comparator string) (taskFilterComparator, error) {
	switch comparator {
	case "equals":
		return taskFilterComparatorEquals, nil
	case "greater":
		return taskFilterComparatorGreater, nil
	case "greater_equals":
		return taskFilterComparatorGreateEquals, nil
	case "less":
		return taskFilterComparatorLess, nil
	case "less_equals":
		return taskFilterComparatorLessEquals, nil
	case "not_equals":
		return taskFilterComparatorNotEquals, nil
	case "like":
		return taskFilterComparatorLike, nil
	default:
		return taskFilterComparatorInvalid, ErrInvalidTaskFilterComparator{Comparator: taskFilterComparator(comparator)}
	}
}

func getNativeValueForTaskField(fieldName, value string) (nativeValue interface{}, err error) {
	field, ok := reflect.TypeOf(&Task{}).Elem().FieldByName(strcase.ToCamel(fieldName))
	if !ok {
		return nil, ErrInvalidTaskField{TaskField: fieldName}
	}

	switch field.Type.Kind() {
	case reflect.Int64:
		nativeValue, err = strconv.ParseInt(value, 10, 64)
	case reflect.Float64:
		nativeValue, err = strconv.ParseFloat(value, 64)
	case reflect.String:
		nativeValue = value
	case reflect.Bool:
		nativeValue, err = strconv.ParseBool(value)
	case reflect.Struct:
		if field.Type == schemas.TimeType {
			nativeValue, err = time.Parse(time.RFC3339, value)
			nativeValue = nativeValue.(time.Time).In(config.GetTimeZone())
		}
	case reflect.Slice:
		t := reflect.SliceOf(schemas.TimeType)
		if t != nil {
			nativeValue, err = time.Parse(time.RFC3339, value)
			nativeValue = nativeValue.(time.Time).In(config.GetTimeZone())
			return
		}
		fallthrough
	default:
		panic(fmt.Errorf("unrecognized filter type %s for field %s, value %s", field.Type.String(), fieldName, value))
	}

	return
}
