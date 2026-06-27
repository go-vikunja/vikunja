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
	"encoding/base64"
	"testing"

	"xorm.io/xorm/schemas"
)

func legacyFrontendSettingsRaw20260627101958(jsonValue string) string {
	return `"` + base64.StdEncoding.EncodeToString([]byte(jsonValue)) + `"`
}

func TestRepairLegacyFrontendSettings20260627101958(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantValue  string
		wantNull   bool
		wantChange bool
	}{
		{
			name:       "legacy null heals to NULL",
			raw:        legacyFrontendSettingsRaw20260627101958("null"),
			wantNull:   true,
			wantChange: true,
		},
		{
			name:       "legacy object is decoded back to the object",
			raw:        legacyFrontendSettingsRaw20260627101958(`{"color_schema":"dark"}`),
			wantValue:  `{"color_schema":"dark"}`,
			wantChange: true,
		},
		{
			name:       "healthy object is left untouched",
			raw:        `{"color_schema":"dark"}`,
			wantChange: false,
		},
		{
			name:       "non-base64 string is left untouched",
			raw:        `"hello world"`,
			wantChange: false,
		},
		{
			name:       "base64 of a scalar is left untouched",
			raw:        legacyFrontendSettingsRaw20260627101958("123"),
			wantChange: false,
		},
		{
			name:       "base64 of an array is left untouched",
			raw:        legacyFrontendSettingsRaw20260627101958("[]"),
			wantChange: false,
		},
		{
			name:       "empty value is left untouched",
			raw:        "",
			wantChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, setNull, changed := repairLegacyFrontendSettings20260627101958(tt.raw)
			if changed != tt.wantChange {
				t.Fatalf("changed = %v, want %v", changed, tt.wantChange)
			}
			if setNull != tt.wantNull {
				t.Fatalf("setNull = %v, want %v", setNull, tt.wantNull)
			}
			if value != tt.wantValue {
				t.Fatalf("value = %q, want %q", value, tt.wantValue)
			}
		})
	}
}

func TestLegacyFrontendSettingsWhereClause20260627101958(t *testing.T) {
	postgres := legacyFrontendSettingsWhereClause20260627101958(schemas.POSTGRES)
	if want := `frontend_settings IS NOT NULL AND frontend_settings::text LIKE '"%'`; postgres != want {
		t.Fatalf("postgres clause\nwant: %s\ngot:  %s", want, postgres)
	}

	other := legacyFrontendSettingsWhereClause20260627101958(schemas.SQLITE)
	if want := `frontend_settings IS NOT NULL AND frontend_settings LIKE '"%'`; other != want {
		t.Fatalf("default clause\nwant: %s\ngot:  %s", want, other)
	}
}
