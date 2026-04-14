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

package handler

import (
	"regexp"
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestMcpHandler_UpdateTool(t *testing.T) {
	keyPattern := regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,64}$`)

	t.Run("task handler", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Task{}
			},
		}
		tool := h.UpdateTool()
		assert.Equal(t, "update_task", tool.Name)
		assert.Contains(t, tool.Description, "task")
		assert.Contains(t, tool.InputSchema.Properties, "id")
		assert.Contains(t, tool.InputSchema.Properties, "title")
		assert.Contains(t, tool.InputSchema.Properties, "description")

		for key := range tool.InputSchema.Properties {
			assert.True(t, keyPattern.MatchString(key), "key %q should match pattern ^[a-zA-Z0-9_.-]{1,64}$", key)
		}
	})

	t.Run("project handler", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Project{}
			},
		}
		tool := h.UpdateTool()
		assert.Equal(t, "update_project", tool.Name)
		assert.Contains(t, tool.Description, "project")
		assert.Contains(t, tool.InputSchema.Properties, "id")
		assert.Contains(t, tool.InputSchema.Properties, "title")

		for key := range tool.InputSchema.Properties {
			assert.True(t, keyPattern.MatchString(key), "key %q should match pattern ^[a-zA-Z0-9_.-]{1,64}$", key)
		}
	})
}
