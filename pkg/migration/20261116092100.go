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
	"fmt"

	"code.vikunja.io/api/pkg/log"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// maxFrontendSettingsSize20261116092100 is models.MaxFrontendSettingsSize when this migration was written.
const maxFrontendSettingsSize20261116092100 = 32 * 1024

const frontendSettingsBatch20261116092100 = 500

// A text column, so xorm returns the raw stored value and does not decode it.
type userFrontendSettings20261116092100 struct {
	ID               int64  `xorm:"bigint autoincr not null unique pk"`
	FrontendSettings string `xorm:"text null"`
}

func (userFrontendSettings20261116092100) TableName() string {
	return "users"
}

// Frozen copy of models.ResetInvalidFrontendSettings - migrations must not follow later model changes.

// validFrontendSettings20261116092100 mirrors the check of the settings update:
// a JSON object of at most maxFrontendSettingsSize20261116092100 bytes.
func validFrontendSettings20261116092100(raw string) bool {
	var settings map[string]any
	if err := json.Unmarshal([]byte(raw), &settings); err != nil || settings == nil {
		return false
	}
	marshalled, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	return len(marshalled) <= maxFrontendSettingsSize20261116092100
}

// resetInvalidFrontendSettings20261116092100 sets every frontend_settings value
// that the settings update would reject to NULL, so the frontend uses its defaults.
func resetInvalidFrontendSettings20261116092100(tx *xorm.Engine) error {
	lastID := int64(0)
	for {
		rows := []*userFrontendSettings20261116092100{}
		err := tx.
			Where(builder.Gt{"id": lastID}).
			And("frontend_settings IS NOT NULL").
			OrderBy("id").
			Limit(frontendSettingsBatch20261116092100).
			Find(&rows)
		if err != nil {
			return fmt.Errorf("could not read frontend settings: %w", err)
		}

		for _, row := range rows {
			if validFrontendSettings20261116092100(row.FrontendSettings) {
				continue
			}
			if _, err := tx.Exec("UPDATE users SET frontend_settings = NULL WHERE id = ?", row.ID); err != nil {
				return err
			}
			log.Infof("Reset the frontend settings of user %d: not a JSON object or larger than %d bytes", row.ID, maxFrontendSettingsSize20261116092100)
		}

		if len(rows) < frontendSettingsBatch20261116092100 {
			break
		}
		lastID = rows[len(rows)-1].ID
	}

	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261116092100",
		Description: "reset frontend_settings that are not a JSON object or are too large",
		Migrate:     resetInvalidFrontendSettings20261116092100,
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
