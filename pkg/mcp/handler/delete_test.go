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
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestMcpHandler_DeleteTool(t *testing.T) {
	t.Run("task handler", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Task{}
			},
		}
		tool := h.DeleteTool()
		assert.Equal(t, "delete_task", tool.Name)
		assert.Contains(t, tool.Description, "task")
		assert.Contains(t, tool.InputSchema.Properties, "id")
		props := tool.InputSchema.Properties["id"].(map[string]any)
		assert.Equal(t, "integer", props["type"])
		assert.Contains(t, props["description"], "id")
	})

	t.Run("project handler", func(t *testing.T) {
		h := &McpHandler{
			EmptyStruct: func() CObject {
				return &models.Project{}
			},
		}
		tool := h.DeleteTool()
		assert.Equal(t, "delete_project", tool.Name)
		assert.Contains(t, tool.Description, "project")
		assert.Contains(t, tool.InputSchema.Properties, "id")
	})
}
