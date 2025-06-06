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

package migration

import (
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type taskCollectionFilter20231121191822 struct {
	SortBy  []string `query:"sort_by" json:"sort_by"`
	OrderBy []string `query:"order_by" json:"order_by"`

	FilterBy         []string `query:"filter_by" json:"filter_by,omitempty"`
	FilterValue      []string `query:"filter_value" json:"filter_value,omitempty"`
	FilterComparator []string `query:"filter_comparator" json:"filter_comparator,omitempty"`
	FilterConcat     string   `query:"filter_concat" json:"filter_concat,omitempty"`

	Filter             string `query:"filter" json:"filter"`
	FilterIncludeNulls bool   `query:"filter_include_nulls" json:"filter_include_nulls"`
}

type savedFilter20231121191822 struct {
	ID      int64                               `xorm:"autoincr not null unique pk" json:"id" param:"filter"`
	Filters *taskCollectionFilter20231121191822 `xorm:"JSON not null" json:"filters" valid:"required"`
}

func (savedFilter20231121191822) TableName() string {
	return "saved_filters"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20231121191822",
		Description: "Migrate saved filter structure",
		Migrate: func(tx *xorm.Engine) (err error) {
			allFilters := []*savedFilter20231121191822{}
			err = tx.Find(&allFilters)
			if err != nil {
				return
			}

			for _, filter := range allFilters {
				var filterStrings []string
				for i, f := range filter.Filters.FilterBy {
					var comparator string
					switch filter.Filters.FilterComparator[i] {
					case "equals":
						comparator = "="
					case "greater":
						comparator = ">"
					case "greater_equals":
						comparator = ">="
					case "less":
						comparator = "<"
					case "less_equals":
						comparator = "<="
					case "not_equals":
						comparator = "!="
					case "like":
						comparator = "~"
					case "in":
						comparator = "?="
					}
					filterStrings = append(filterStrings, f+" "+comparator+" "+filter.Filters.FilterValue[i])
				}

				filter.Filters.FilterConcat = " || "
				if filter.Filters.FilterConcat == "and" {
					filter.Filters.FilterConcat = " && "
				}
				filter.Filters.Filter = strings.Join(filterStrings, filter.Filters.FilterConcat)

				filter.Filters.FilterBy = nil
				filter.Filters.FilterComparator = nil
				filter.Filters.FilterValue = nil
				filter.Filters.FilterConcat = ""

				_, err = tx.Where("id = ?", filter.ID).Update(filter)
				if err != nil {
					return
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
