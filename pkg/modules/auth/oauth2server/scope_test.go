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

package oauth2server

import (
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseScope(t *testing.T) {
	for raw, want := range map[string]string{
		"mcp:read":                    "mcp:read",
		"mcp:write":                   "mcp:read mcp:write",
		"offline_access mcp:read":     "mcp:read",
		"MCP:WRITE mcp:read":          "mcp:read",
		"offline_access mcp:readonly": "",
	} {
		scopes, err := parseScope(raw)
		if want == "" {
			require.Error(t, err, raw)
			assert.True(t, models.IsErrOAuthInvalidScope(err), raw)
			continue
		}
		require.NoError(t, err, raw)
		assert.Equal(t, want, formatScope(scopes), raw)
	}
}

func TestSameResource(t *testing.T) {
	const canonical = "https://vikunja.example/api/v2/mcp"
	for _, same := range []string{
		"https://vikunja.example/api/v2/mcp",
		"https://vikunja.example/api/v2/mcp/",
		"HTTPS://Vikunja.Example/api/v2/mcp",
	} {
		assert.True(t, SameResource(same, canonical), same)
	}
	for _, other := range []string{
		"http://vikunja.example/api/v2/mcp",
		"https://evil.example/api/v2/mcp",
		"https://vikunja.example/api/v2/MCP",
		"https://vikunja.example/api/v2/mcp?x=1",
		"https://vikunja.example/api/v2/mcp?",
		"https://vikunja.example/api/v2/mcp#",
		"https://user@vikunja.example/api/v2/mcp",
	} {
		assert.False(t, SameResource(other, canonical), other)
	}
	assert.False(t, SameResource("/api/v2/mcp", "/api/v2/mcp"), "relative URLs never match")
}
