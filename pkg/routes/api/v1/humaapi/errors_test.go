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

package humaapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVikunjaErrorShape_BasicCodeMessage(t *testing.T) {
	err := NewVikunjaError(http.StatusForbidden, "Forbidden")
	b, marshalErr := json.Marshal(err)
	require.NoError(t, marshalErr)

	var got map[string]any
	require.NoError(t, json.Unmarshal(b, &got))
	assert.Equal(t, "Forbidden", got["message"])
	// must not include RFC 9457 fields
	_, hasType := got["type"]
	_, hasTitle := got["title"]
	assert.False(t, hasType, "unexpected RFC 9457 field 'type'")
	assert.False(t, hasTitle, "unexpected RFC 9457 field 'title'")
}

func TestVikunjaErrorShape_StatusCoderInterface(t *testing.T) {
	e := NewVikunjaError(http.StatusNotFound, "not found")
	assert.Equal(t, http.StatusNotFound, e.GetStatus())
}
