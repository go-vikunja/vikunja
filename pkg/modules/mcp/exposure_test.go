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
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
)

func TestExposure(t *testing.T) {
	for _, tc := range []struct {
		id   string
		tier Tier
	}{
		{
			"tasks-create",
			TierTyped,
		},
		{
			"task-labels-create",
			TierCatalog,
		},
	} {
		op := &huma.Operation{
			OperationID: tc.id,
			Method:      http.MethodPost,
			RequestBody: &huma.RequestBody{Content: map[string]*huma.MediaType{"application/json": {Schema: &huma.Schema{Type: "object"}}}},
		}
		tier, ok := exposure(tc.id, op)
		assert.True(t, ok)
		assert.Equal(t, tc.tier, tier)
	}
	for _, id := range []string{
		"tokens-create",
		"token-test",
		"auth-login",
		"user-show",
		"webhooks-list",
		"shares-read",
		"migration-csv-migrate",
		"task-attachments-download",
		"admin-users-list",
		"bots-create",
		"health",
		"info",
		"notifications-atom-feed",
	} {
		_, ok := exposure(id, &huma.Operation{
			OperationID: id,
			Method:      http.MethodGet,
		})
		assert.False(t, ok, id)
	}
	for _, ct := range []string{
		"multipart/form-data",
		"application/merge-patch+json",
	} {
		op := &huma.Operation{
			Method:      http.MethodPatch,
			RequestBody: &huma.RequestBody{Content: map[string]*huma.MediaType{ct: {Schema: &huma.Schema{Type: "object"}}}},
		}
		tier, ok := exposure("tasks-update", op)
		assert.Equal(t, ct == "application/merge-patch+json", ok)
		if ok {
			assert.Equal(t, TierTyped, tier)
		}
	}
}

func TestToolNameFor(t *testing.T) {
	assert.Equal(t, "task_comments_create", toolNameFor("task-comments-create"))
}
