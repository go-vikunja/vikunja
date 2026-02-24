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
	"bytes"
	"encoding/json"
	"fmt"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type bucketConfigFix20260224122023 struct {
	Title  string          `json:"title"`
	Filter json.RawMessage `json:"filter"`
}

type bucketConfigFixNew20260224122023 struct {
	Title  string                         `json:"title"`
	Filter *bucketFilterFix20260224122023 `json:"filter"`
}

type bucketFilterFix20260224122023 struct {
	Search             string   `json:"s,omitempty"`
	SortBy             []string `json:"sort_by,omitempty"`
	OrderBy            []string `json:"order_by,omitempty"`
	Filter             string   `json:"filter,omitempty"`
	FilterIncludeNulls bool     `json:"filter_include_nulls,omitempty"`
}

type projectViewFix20260224122023 struct {
	ID                  int64                            `xorm:"autoincr not null unique pk"`
	BucketConfiguration []*bucketConfigFix20260224122023 `xorm:"json"`
}

func (projectViewFix20260224122023) TableName() string {
	return "project_views"
}

type projectViewFixNew20260224122023 struct {
	ID                  int64                               `xorm:"autoincr not null unique pk"`
	BucketConfiguration []*bucketConfigFixNew20260224122023 `xorm:"json"`
}

func (projectViewFixNew20260224122023) TableName() string {
	return "project_views"
}

type projectViewFilterFix20260224122023 struct {
	ID     int64  `xorm:"autoincr not null unique pk"`
	Filter string `xorm:"json null default null"`
}

func (projectViewFilterFix20260224122023) TableName() string {
	return "project_views"
}

type projectViewFilterFixNew20260224122023 struct {
	ID     int64                         `xorm:"autoincr not null unique pk"`
	Filter *taskCollection20241118123644 `xorm:"json null default null"`
}

func (projectViewFilterFixNew20260224122023) TableName() string {
	return "project_views"
}

func convertBucketFilter20260224122023(raw json.RawMessage) (*bucketFilterFix20260224122023, bool, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, false, nil
	}

	switch trimmed[0] {
	case '"':
		var filterString string
		if err := json.Unmarshal(trimmed, &filterString); err != nil {
			return nil, false, err
		}
		if filterString == "" {
			return nil, true, nil
		}
		return &bucketFilterFix20260224122023{
			Filter: filterString,
		}, true, nil
	case '{':
		var existingFilter bucketFilterFix20260224122023
		if err := json.Unmarshal(trimmed, &existingFilter); err != nil {
			return nil, false, err
		}
		return &existingFilter, false, nil
	default:
		return nil, false, fmt.Errorf("unexpected bucket filter JSON value: %s", string(trimmed))
	}
}

func convertBucketConfigurations20260224122023(configs []*bucketConfigFix20260224122023) ([]*bucketConfigFixNew20260224122023, bool, error) {
	converted := make([]*bucketConfigFixNew20260224122023, 0, len(configs))
	changed := false

	for _, config := range configs {
		if config == nil {
			converted = append(converted, nil)
			continue
		}

		filter, filterChanged, err := convertBucketFilter20260224122023(config.Filter)
		if err != nil {
			return nil, false, err
		}
		if filterChanged {
			changed = true
		}

		converted = append(converted, &bucketConfigFixNew20260224122023{
			Title:  config.Title,
			Filter: filter,
		})
	}

	return converted, changed, nil
}

func bucketConfigurationWhereClause20260224122023(dbType schemas.DBType) string {
	if dbType == schemas.POSTGRES {
		return "bucket_configuration IS NOT NULL AND bucket_configuration::text != '' AND bucket_configuration::text != '[]' AND bucket_configuration::text != 'null'"
	}

	return "bucket_configuration IS NOT NULL AND bucket_configuration != '' AND bucket_configuration != '[]' AND bucket_configuration != 'null'"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260224122023",
		Description: "comprehensive catchup for bucket_configuration and view filter format conversions",
		Migrate: func(tx *xorm.Engine) (err error) {
			allViews := []*projectViewFix20260224122023{}
			err = tx.Where(bucketConfigurationWhereClause20260224122023(tx.Dialect().URI().DBType)).
				Find(&allViews)
			if err != nil {
				return
			}

			for _, view := range allViews {
				converted, needsUpdate, convErr := convertBucketConfigurations20260224122023(view.BucketConfiguration)
				if convErr != nil {
					return convErr
				}
				if !needsUpdate {
					continue
				}

				newView := &projectViewFixNew20260224122023{
					ID:                  view.ID,
					BucketConfiguration: converted,
				}

				_, err = tx.Where("id = ?", view.ID).
					Cols("bucket_configuration").
					Update(newView)
				if err != nil {
					return
				}
			}

			oldFilterViews := []*projectViewFilterFix20260224122023{}
			if tx.Dialect().URI().DBType == schemas.POSTGRES {
				err = tx.Where("filter IS NOT NULL AND filter::text != '' AND filter::text NOT LIKE '{%'").Find(&oldFilterViews)
			} else {
				err = tx.Where("filter IS NOT NULL AND filter != '' AND filter NOT LIKE '{%'").Find(&oldFilterViews)
			}
			if err != nil {
				return
			}

			for _, view := range oldFilterViews {
				newView := &projectViewFilterFixNew20260224122023{
					ID: view.ID,
					Filter: &taskCollection20241118123644{
						Filter: view.Filter,
					},
				}

				_, err = tx.Where("id = ?", view.ID).
					Cols("filter").
					Update(newView)
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
