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

// usersFrontendSettings20260627101958 reads the raw frontend_settings JSON text.
// The json tag makes xorm hand back the stored column text verbatim instead of
// decoding it into a typed value.
type usersFrontendSettings20260627101958 struct {
	ID               int64  `xorm:"bigint autoincr not null unique pk"`
	FrontendSettings string `xorm:"frontend_settings json null"`
}

func (usersFrontendSettings20260627101958) TableName() string {
	return "users"
}

// legacyFrontendSettingsWhereClause20260627101958 selects rows whose
// frontend_settings is a JSON string ('"…"'). Healthy values are JSON objects
// ('{…}'); only the legacy double-encoded values are stored as JSON strings.
func legacyFrontendSettingsWhereClause20260627101958(dbType schemas.DBType) string {
	if dbType == schemas.POSTGRES {
		return `frontend_settings IS NOT NULL AND frontend_settings::text LIKE '"%'`
	}
	return `frontend_settings IS NOT NULL AND frontend_settings LIKE '"%'`
}

// repairLegacyFrontendSettings20260627101958 reverses the historical double
// encoding of frontend_settings. The pre-fix UpdateUser stored
// json.Marshal(FrontendSettings) back into the interface field, so xorm
// base64-encoded the resulting []byte: a nil value became the JSON string
// "bnVsbA==" (base64 of "null") and a real settings object became the base64 of
// its JSON.
//
// changed reports whether a repair applies; setNull reports the value should
// become SQL NULL; value carries the repaired JSON object otherwise. Only values
// that base64-decode to JSON null or a JSON object are touched, so a legitimate
// string-valued setting is never rewritten.
func repairLegacyFrontendSettings20260627101958(raw string) (value string, setNull, changed bool) {
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
		Description: "repair double-encoded frontend_settings stored before UpdateUser stopped re-marshalling the column",
		Migrate: func(tx *xorm.Engine) error {
			users := []*usersFrontendSettings20260627101958{}
			if err := tx.
				Where(legacyFrontendSettingsWhereClause20260627101958(tx.Dialect().URI().DBType)).
				Find(&users); err != nil {
				return err
			}

			for _, u := range users {
				value, setNull, changed := repairLegacyFrontendSettings20260627101958(u.FrontendSettings)
				if !changed {
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
