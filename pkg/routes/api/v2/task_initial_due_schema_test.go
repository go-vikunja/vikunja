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

package apiv2

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	initialDueField    = "initial_due_time"
	dueDateSourceField = "due_date_source"
	reasonField        = "reason"
)

// TestTaskInitialDueSchema_OpenAPI pins where the first-occurrence fields appear in the v2 OpenAPI
// document: initial_due_time on the create and update requests, write-only and
// with the HH:MM pattern; due_date_source and reason on the create and update responses only,
// read-only; none of them on task GET beyond the shared, write-only initial_due_time, nor on the
// list, bulk or Task component schemas.
func TestTaskInitialDueSchema_OpenAPI(t *testing.T) {
	r := newRuleTextSchemas(t)

	for _, id := range []string{"tasks-create", "tasks-update"} {
		prop := r.request(id).Properties[initialDueField]
		require.NotNil(t, prop, "%s request has %s", id, initialDueField)
		assert.Equal(t, huma.TypeString, prop.Type, id)
		assert.Equal(t, "^([01][0-9]|2[0-3]):[0-5][0-9]$", prop.Pattern, id)
		assert.True(t, prop.WriteOnly, "%s: %s is write-only", id, initialDueField)
		assert.False(t, prop.ReadOnly, id)
		for _, fact := range []string{"no due date", "strictly after now", "stored as due_date", "Never stored or returned"} {
			assert.Contains(t, prop.Description, fact, id)
		}
	}

	// tasks-read shares its body with the update input, so it lists initial_due_time, write-only;
	// the GET JSON never carries it (TestTaskInitialDue_ReadJSON). The write-response fields stay
	// off it.
	read := r.response("tasks-read", "200")
	require.NotNil(t, read.Properties[initialDueField])
	assert.True(t, read.Properties[initialDueField].WriteOnly)
	for _, id := range []string{"tasks-read", "tasks-read-by-index"} {
		for _, field := range []string{dueDateSourceField, reasonField} {
			assert.NotContains(t, r.response(id, "200").Properties, field, "%s response", id)
		}
	}

	for _, w := range []struct{ id, status string }{{"tasks-create", "201"}, {"tasks-update", "200"}} {
		s := r.response(w.id, w.status)
		source := s.Properties[dueDateSourceField]
		require.NotNil(t, source, "%s response has %s", w.id, dueDateSourceField)
		assert.True(t, source.ReadOnly, w.id)
		assert.ElementsMatch(t, []any{models.DueDateSourceExplicit, models.DueDateSourceGenerated}, source.Enum, w.id)
		reason := s.Properties[reasonField]
		require.NotNil(t, reason, "%s response has %s", w.id, reasonField)
		assert.True(t, reason.ReadOnly, w.id)
		assert.Contains(t, reason.Description, "until_passed", w.id)
		assert.NotContains(t, s.Properties, initialDueField, "%s response: never returned", w.id)
	}
	for _, id := range []string{"tasks-create", "tasks-update"} {
		for _, field := range []string{dueDateSourceField, reasonField} {
			assert.NotContains(t, r.request(id).Properties, field, "%s request", id)
		}
	}

	task := r.doc.Components.Schemas.Map()["Task"]
	require.NotNil(t, task)
	for _, field := range []string{initialDueField, dueDateSourceField, reasonField} {
		assert.NotContains(t, task.Properties, field, "models.Task (v1, list, bulk) must not carry %s", field)
	}

	// Not accepted or documented on list and bulk operations.
	for _, id := range []string{"tasks-list", "project-tasks-list", "project-view-tasks-list", "project-view-buckets-tasks-list", "tasks-bulk-update", "tasks-bulk-create"} {
		op := r.op(id)
		for _, field := range []string{initialDueField, dueDateSourceField, reasonField} {
			if op.RequestBody != nil {
				for mediaType, media := range op.RequestBody.Content {
					assert.False(t, r.mentions(media.Schema, field, map[string]bool{}), "%s request (%s): %s", id, mediaType, field)
				}
			}
			for status, resp := range op.Responses {
				for mediaType, media := range resp.Content {
					assert.False(t, r.mentions(media.Schema, field, map[string]bool{}), "%s %s response (%s): %s", id, status, mediaType, field)
				}
			}
		}
	}
}

// TestTaskInitialDue_ModelUnchanged guards that models.Task, v1's and CalDAV's shape,
// has no such field at all.
func TestTaskInitialDue_ModelUnchanged(t *testing.T) {
	typ := reflect.TypeOf(models.Task{})
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		tag := strings.Split(f.Tag.Get("json"), ",")[0]
		for _, field := range []string{initialDueField, dueDateSourceField, reasonField} {
			assert.NotEqual(t, field, tag, "models.Task.%s", f.Name)
		}
		assert.NotContains(t, strings.ToLower(f.Name), "initialdue", "models.Task.%s", f.Name)
	}
}

// TestTaskInitialDue_ReadJSON: the GET body type lists initial_due_time, but its JSON never does,
// since a read never fills it; and the write body reports the source only when there is a date.
func TestTaskInitialDue_ReadJSON(t *testing.T) {
	due := time.Date(2030, 1, 15, 14, 0, 0, 0, time.UTC)
	raw, err := json.Marshal(&taskReadOneBody{Task: models.Task{ID: 1, Title: "read", DueDate: due}})
	require.NoError(t, err)
	assert.NotContains(t, string(raw), initialDueField)
	assert.NotContains(t, string(raw), dueDateSourceField)

	body := func(task *models.Task, sel initialDueResult) map[string]any {
		raw, err := json.Marshal(newTaskWriteBody(task, sel))
		require.NoError(t, err)
		out := map[string]any{}
		require.NoError(t, json.Unmarshal(raw, &out))
		return out
	}
	undated := body(&models.Task{ID: 1}, initialDueResult{})
	assert.NotContains(t, undated, dueDateSourceField, "absent without a due date")
	assert.NotContains(t, undated, reasonField)

	explicit := body(&models.Task{ID: 1, DueDate: due}, initialDueResult{})
	assert.Equal(t, models.DueDateSourceExplicit, explicit[dueDateSourceField])

	generated := body(&models.Task{ID: 1, DueDate: due}, initialDueResult{source: models.DueDateSourceGenerated})
	assert.Equal(t, models.DueDateSourceGenerated, generated[dueDateSourceField])
	assert.NotContains(t, generated, reasonField)

	none := body(&models.Task{ID: 1}, initialDueResult{reason: "until_passed"})
	assert.NotContains(t, none, dueDateSourceField)
	assert.Equal(t, "until_passed", none[reasonField])
}
