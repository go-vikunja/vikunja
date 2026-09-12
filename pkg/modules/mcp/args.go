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

package mcp

import (
	"encoding/json"
	"errors"
)

func decodeArgs(spec *toolSpec, raw json.RawMessage) (map[string]json.RawMessage, error) {
	instance := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &instance); err != nil || instance == nil {
			return nil, errors.New("arguments must be a JSON object")
		}
	}
	if err := spec.resolved.Validate(instance); err != nil {
		return nil, err
	}
	args := map[string]json.RawMessage{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &args); err != nil {
			return nil, errors.New("arguments must be a JSON object")
		}
	}
	return args, nil
}
