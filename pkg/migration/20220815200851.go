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
	"encoding/json"
	"strconv"
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20220815200851",
		Description: "Migrate saved assignee filter to usernames instead of IDs",
		Migrate: func(tx *xorm.Engine) error {
			filters := []map[string]interface{}{} // not using the type here so that the migration does not depend on it
			err := tx.Select("*").
				Table("saved_filters").
				Find(&filters)
			if err != nil {
				return err
			}

			for _, f := range filters {
				filter := map[string]interface{}{}
				filterJSON, is := f["filters"].(string)
				if !is {
					continue
				}
				err = json.Unmarshal([]byte(filterJSON), &filter)
				if err != nil {
					return err
				}
				filterBy := filter["filter_by"].([]interface{})
				filterValue := filter["filter_value"].([]interface{})
				for p, fb := range filterBy {
					if fb == "assignees" || fb == "user_id" {
						userIDs := []int64{}
						for _, sid := range strings.Split(filterValue[p].(string), ",") {
							id, err := strconv.ParseInt(sid, 10, 64)
							if err != nil {
								return err
							}
							userIDs = append(userIDs, id)
						}

						usernames := []string{}
						err := tx.Select("username").
							Table("users").
							In("id", userIDs).
							Find(&usernames)
						if err != nil {
							return err
						}

						userfilter := ""
						for i, username := range usernames {
							if i > 0 {
								userfilter += ","
							}
							userfilter += username
						}
						filterValue[p] = userfilter
					}
				}

				filter["filter_value"] = filterValue
				filtersJSON, err := json.Marshal(filter)
				if err != nil {
					return err
				}

				f["filters"] = string(filtersJSON)

				_, err = tx.Where("id = ?", f["id"]).
					Cols("filters").
					NoAutoCondition().
					Table("saved_filters").
					Update(f)
				if err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
