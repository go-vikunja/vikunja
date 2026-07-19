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
	"encoding/base64"
	"encoding/json"
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type userFrontendSettings20260627101958 struct {
	ID int64 `xorm:"bigint autoincr not null unique pk"`
	// Keep this as a string; interface{} would decode the value before repair.
	FrontendSettings string `xorm:"frontend_settings json null"`
}

func (userFrontendSettings20260627101958) TableName() string {
	return "users"
}

func frontendSettingsStringWhere20260627101958(dbType schemas.DBType) string {
	if dbType == schemas.POSTGRES {
		return `frontend_settings IS NOT NULL AND frontend_settings::text LIKE '"%'`
	}
	return `frontend_settings IS NOT NULL AND frontend_settings LIKE '"%'`
}

// Legacy rows contain base64 JSON because UpdateUser wrote []byte into an xorm json field.
func decodeLegacyFrontendSettings20260627101958(raw string) (value string, setNull, ok bool) {
	trimmed := strings.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '"' {
		return "", false, false
	}

	var inner string
	if err := json.Unmarshal([]byte(trimmed), &inner); err != nil {
		return "", false, false
	}

	decoded, err := base64.StdEncoding.DecodeString(inner)
	if err != nil {
		return "", false, false
	}

	decoded = bytes.TrimSpace(decoded)
	if bytes.Equal(decoded, []byte("null")) {
		return "", true, true
	}
	if len(decoded) > 0 && decoded[0] == '{' && json.Valid(decoded) {
		return string(decoded), false, true
	}
	return "", false, false
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260627101958",
		Description: "repair double-encoded frontend_settings",
		Migrate: func(tx *xorm.Engine) error {
			users := []*userFrontendSettings20260627101958{}
			if err := tx.
				Where(frontendSettingsStringWhere20260627101958(tx.Dialect().URI().DBType)).
				Find(&users); err != nil {
				return err
			}

			for _, u := range users {
				value, setNull, ok := decodeLegacyFrontendSettings20260627101958(u.FrontendSettings)
				if !ok {
					continue
				}

				if setNull {
					if _, err := tx.Exec("UPDATE users SET frontend_settings = NULL WHERE id = ?", u.ID); err != nil {
						return err
					}
					continue
				}

				if _, err := tx.Exec("UPDATE users SET frontend_settings = ? WHERE id = ?", value, u.ID); err != nil {
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
