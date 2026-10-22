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

func TestDecodeLegacyFrontendSettings20260627101958(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantValue string
		wantNull  bool
		wantOK    bool
	}{
		{
			name:     "legacy null",
			raw:      legacyFrontendSettingsRaw20260627101958("null"),
			wantNull: true,
			wantOK:   true,
		},
		{
			name:      "legacy object",
			raw:       legacyFrontendSettingsRaw20260627101958(`{"color_schema":"dark"}`),
			wantValue: `{"color_schema":"dark"}`,
			wantOK:    true,
		},
		{
			name:   "object",
			raw:    `{"color_schema":"dark"}`,
			wantOK: false,
		},
		{
			name:   "string",
			raw:    `"hello world"`,
			wantOK: false,
		},
		{
			name:   "encoded scalar",
			raw:    legacyFrontendSettingsRaw20260627101958("123"),
			wantOK: false,
		},
		{
			name:   "encoded array",
			raw:    legacyFrontendSettingsRaw20260627101958("[]"),
			wantOK: false,
		},
		{
			name:   "empty",
			raw:    "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, setNull, ok := decodeLegacyFrontendSettings20260627101958(tt.raw)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
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

func TestFrontendSettingsStringWhere20260627101958(t *testing.T) {
	postgres := frontendSettingsStringWhere20260627101958(schemas.POSTGRES)
	if want := `frontend_settings IS NOT NULL AND frontend_settings::text LIKE '"%'`; postgres != want {
		t.Fatalf("postgres clause\nwant: %s\ngot:  %s", want, postgres)
	}

	other := frontendSettingsStringWhere20260627101958(schemas.SQLITE)
	if want := `frontend_settings IS NOT NULL AND frontend_settings LIKE '"%'`; other != want {
		t.Fatalf("default clause\nwant: %s\ngot:  %s", want, other)
	}
}
