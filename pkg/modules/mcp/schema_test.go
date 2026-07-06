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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// registerAllResources loads the real production declarations into a clean
// registry so the schema tests exercise the actual derived contracts.
func registerAllResources(t *testing.T) {
	t.Helper()
	resetRegistry(t)
	for _, r := range allResources() {
		require.NoError(t, Register(r))
	}
}

func specFor(t *testing.T, resource string, op Op) *opSpec {
	t.Helper()
	r, ok := lookupResource(resource)
	require.True(t, ok, "resource %s not registered", resource)
	spec := r.spec(op)
	require.NotNil(t, spec, "no spec for %s_%s", resource, op.ToolSuffix())
	return spec
}

func TestSchema_TaskCreate(t *testing.T) {
	registerAllResources(t)
	spec := specFor(t, "tasks", OpCreate)

	props := spec.schema.Properties
	// Writable model fields are exposed…
	for _, want := range []string{"title", "project_id", "description", "done", "due_date", "priority", "repeat_after", "repeat_mode", "percent_done", "hex_color", "bucket_id", "cover_image_attachment_id"} {
		assert.Contains(t, props, want)
	}
	// …server-controlled / relation fields are not.
	for _, banned := range []string{"id", "created", "updated", "created_by", "assignees", "labels", "attachments", "identifier", "related_tasks", "reactions"} {
		assert.NotContains(t, props, banned)
	}
	// Title (minLength) and project_id (URL-bound in REST) are required.
	assert.Equal(t, []string{"project_id", "title"}, spec.schema.Required)
	// doc: tags become property descriptions.
	assert.NotEmpty(t, props["title"].Description)
	// time.Time fields are date-time strings.
	assert.Equal(t, "string", props["due_date"].Type)
	assert.Equal(t, "date-time", props["due_date"].Format)
}

func TestSchema_TaskUpdateRequiresOnlyID(t *testing.T) {
	registerAllResources(t)
	spec := specFor(t, "tasks", OpUpdate)

	assert.Equal(t, []string{"id"}, spec.schema.Required)
	assert.Contains(t, spec.schema.Properties, "done")
	assert.Contains(t, spec.schema.Properties, "project_id")
}

func TestSchema_TaskReadAllUsesTaskCollection(t *testing.T) {
	registerAllResources(t)
	spec := specFor(t, "tasks", OpReadAll)

	props := spec.schema.Properties
	// The filter surface comes from TaskCollection's query-tagged fields.
	for _, want := range []string{"filter", "sort_by", "order_by", "filter_include_nulls", "project_id", argSearch, argPage, argPerPage} {
		assert.Contains(t, props, want)
	}
	// "s" duplicates search and the view path stays REST-only.
	assert.NotContains(t, props, "s")
	assert.NotContains(t, props, "project_view_id")
	// Listing must work across projects: nothing is required.
	assert.Empty(t, spec.schema.Required)
	assert.Equal(t, "array", props["sort_by"].Type)
}

func TestSchema_ReadOneAndDeleteIdentifyByID(t *testing.T) {
	registerAllResources(t)

	for _, op := range []Op{OpReadOne, OpDelete} {
		spec := specFor(t, "labels", op)
		assert.Equal(t, []string{"id"}, spec.schema.Required, op.ToolSuffix())
		assert.Len(t, spec.schema.Properties, 1, op.ToolSuffix())
	}
}

func TestSchema_TaskCommentsCarryParentTaskID(t *testing.T) {
	registerAllResources(t)

	// TaskComment.TaskID is json:"-" param:"task" — REST binds it from the
	// URL; MCP exposes it as a required task_id argument on every op.
	for _, op := range []Op{OpCreate, OpReadOne, OpReadAll, OpUpdate, OpDelete} {
		spec := specFor(t, "tasks_comments", op)
		assert.Contains(t, spec.schema.Properties, "task_id", op.ToolSuffix())
		assert.Contains(t, spec.schema.Required, "task_id", op.ToolSuffix())
	}
	assert.Contains(t, specFor(t, "tasks_comments", OpCreate).schema.Required, "comment")
}

func TestSchema_AssigneesIdentifyByParams(t *testing.T) {
	registerAllResources(t)

	// TaskAssginee has no JSON-exposed id; delete identifies the row via its
	// param-tagged fields instead.
	spec := specFor(t, "tasks_assignees", OpDelete)
	assert.NotContains(t, spec.schema.Properties, "id")
	assert.Equal(t, []string{"task_id", "user_id"}, spec.schema.Required)

	create := specFor(t, "tasks_assignees", OpCreate)
	assert.Equal(t, []string{"task_id", "user_id"}, create.schema.Required)
}

func TestSchema_ProjectCreateRequiresTitleOnly(t *testing.T) {
	registerAllResources(t)
	spec := specFor(t, "projects", OpCreate)

	assert.Equal(t, []string{"title"}, spec.schema.Required)
	assert.Contains(t, spec.schema.Properties, "parent_project_id")
	assert.NotContains(t, spec.schema.Properties, "owner")
	assert.NotContains(t, spec.schema.Properties, "views")
}

func TestSnakeCase(t *testing.T) {
	cases := map[string]string{
		"TaskID":      "task_id",
		"UserID":      "user_id",
		"OtherTaskID": "other_task_id",
		"ProjectView": "project_view",
		"ID":          "id",
		"Title":       "title",
	}
	for in, want := range cases {
		assert.Equal(t, want, snakeCase(in), in)
	}
}
