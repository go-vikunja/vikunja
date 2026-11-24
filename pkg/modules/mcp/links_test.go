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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebUIBases(t *testing.T) {
	for _, raw := range []string{
		"https://example.org", "https://example.org/vikunja/", "http://100.64.0.1:3456/",
		"http://[::1]:3456/prefix", "https://example.org/with%20space/",
	} {
		t.Run(raw, func(t *testing.T) {
			_, err := normalizeWebUIBase(raw)
			require.NoError(t, err)
		})
	}
	for _, raw := range []string{
		"", "/relative", "ftp://example.org", "https://", "https://user:password@example.org",
		"https://example.org/?token=secret", "https://example.org/?", "https://example.org/#",
		"https://example.org/#secret", "https://example.org:abc", "https://example.org:65536",
		"https://example.org:", "https://example.org:0", "https://example.org/a/../b",
		"https://example.org/%2e%2e", "https://example.org/has space", "https://example.org/%zz",
	} {
		t.Run(raw, func(t *testing.T) {
			_, err := normalizeWebUIBase(raw)
			require.Error(t, err)
			assert.NotContains(t, err.Error(), "secret")
		})
	}
	args := map[string]json.RawMessage{
		webUIBasesParam: json.RawMessage(`[{"label":"public","url":"https://example.org/prefix///"}]`),
	}
	bases, err := takeWebUIBases(args)
	require.NoError(t, err)
	require.Len(t, bases, 1)
	assert.Equal(t, "https://example.org/prefix", bases[0].URL)
	assert.NotContains(t, args, webUIBasesParam)
}

func TestWebLinks(t *testing.T) {
	args := map[string]json.RawMessage{"project": json.RawMessage(`42`)}
	source := map[string]any{
		"id":   json.Number("123"),
		"hash": "capability",
	}
	bases := []webUIBase{
		{Label: "public", URL: "https://example.org/prefix"},
		{Label: "vpn", URL: "http://100.64.0.1:3456"},
	}
	task := resourceWebLinks("authenticated_task", args, bases, source)
	assert.Equal(t, "available", task["status"])
	urls := task["urls"].([]map[string]string)
	assert.Equal(t, "https://example.org/prefix/tasks/123", urls[0]["url"])
	assert.Equal(t, "http://100.64.0.1:3456/tasks/123", urls[1]["url"])
	assert.Equal(t, "web_ui_bases argument", urls[0]["source"])
	share := resourceWebLinks("project_share", args, bases, source)
	assert.Equal(t, json.Number("42"), share["project_id"])
	assert.Equal(t, "https://example.org/prefix/#share-auth-token=capability", share["urls"].([]map[string]string)[0]["url"])
	missing := resourceWebLinks("authenticated_task", args, nil, source)
	assert.Equal(t, "missing_base_url", missing["status"])
	assert.Empty(t, missing["urls"])
	projected := map[string]any{"title": "only selected field"}
	result := addWebLinks("tasks-read", args, bases, projected, source).(map[string]any)
	assert.NotContains(t, result, "hash")
	assert.Equal(t, json.Number("123"), result["web_links"].(map[string]any)["id"])
}

func TestShareAnnotations(t *testing.T) {
	assert.Nil(t, shareAnnotations("tasks-read"))
	assert.True(t, shareAnnotations("shares-read").ReadOnlyHint)
	assert.True(t, *shareAnnotations("shares-delete").DestructiveHint)
	assert.False(t, *shareAnnotations("shares-create").DestructiveHint)
	assert.False(t, shareAnnotations("shares-create").IdempotentHint)
}
