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
	"maps"
	"slices"
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ruleTextField = "repeats_rrule"

// ruleTextSchemas finds operations by id and resolves their body schemas in the canonical v2
// OpenAPI document.
type ruleTextSchemas struct {
	t   *testing.T
	doc *huma.OpenAPI
	ops map[string]*huma.Operation
}

func newRuleTextSchemas(t *testing.T) *ruleTextSchemas {
	t.Helper()
	api, err := NewCanonicalAPI()
	require.NoError(t, err)
	doc := api.OpenAPI()
	ops := map[string]*huma.Operation{}
	for _, item := range doc.Paths {
		for _, op := range []*huma.Operation{item.Get, item.Put, item.Post, item.Patch, item.Delete} {
			if op != nil {
				ops[op.OperationID] = op
			}
		}
	}
	return &ruleTextSchemas{t: t, doc: doc, ops: ops}
}

func (r *ruleTextSchemas) resolve(s *huma.Schema) *huma.Schema {
	r.t.Helper()
	require.NotNil(r.t, s)
	if s.Ref != "" {
		resolved := r.doc.Components.Schemas.SchemaFromRef(s.Ref)
		require.NotNil(r.t, resolved, "schema %s", s.Ref)
		return resolved
	}
	return s
}

func (r *ruleTextSchemas) op(id string) *huma.Operation {
	r.t.Helper()
	op := r.ops[id]
	require.NotNil(r.t, op, "operation %s", id)
	return op
}

func (r *ruleTextSchemas) request(id string) *huma.Schema {
	r.t.Helper()
	op := r.op(id)
	require.NotNil(r.t, op.RequestBody, "%s request body", id)
	media := op.RequestBody.Content["application/json"]
	require.NotNil(r.t, media, "%s request media", id)
	return r.resolve(media.Schema)
}

func (r *ruleTextSchemas) response(id, status string) *huma.Schema {
	r.t.Helper()
	resp := r.op(id).Responses[status]
	require.NotNil(r.t, resp, "%s %s response", id, status)
	media := resp.Content["application/json"]
	require.NotNil(r.t, media, "%s %s response media", id, status)
	return r.resolve(media.Schema)
}

// mentions reports whether the schema, followed through refs, items and compositions, has a
// property named name anywhere.
func (r *ruleTextSchemas) mentions(s *huma.Schema, name string, seen map[string]bool) bool {
	if s == nil {
		return false
	}
	if s.Ref != "" {
		if seen[s.Ref] {
			return false
		}
		seen[s.Ref] = true
		return r.mentions(r.doc.Components.Schemas.SchemaFromRef(s.Ref), name, seen)
	}
	if _, ok := s.Properties[name]; ok {
		return true
	}
	for _, p := range s.Properties {
		if r.mentions(p, name, seen) {
			return true
		}
	}
	for _, group := range [][]*huma.Schema{s.OneOf, s.AnyOf, s.AllOf} {
		for _, sub := range group {
			if r.mentions(sub, name, seen) {
				return true
			}
		}
	}
	return r.mentions(s.Items, name, seen)
}

// propertyNames lists a schema's fields, without the $schema link Huma adds to top-level bodies.
func propertyNames(s *huma.Schema) []string {
	names := slices.Sorted(maps.Keys(s.Properties))
	return slices.DeleteFunc(names, func(name string) bool { return name == "$schema" })
}

// TestTaskRuleTextSchema_OpenAPI pins where repeats_rrule appears in the v2 OpenAPI document: on
// the single-task read, create and update responses and on the create and update requests, each
// described for MCP readers; not on list or bulk operations; and the write responses gain only it
// plus the initial_due_time result fields.
func TestTaskRuleTextSchema_OpenAPI(t *testing.T) {
	r := newRuleTextSchemas(t)

	carriers := map[string]*huma.Schema{
		"tasks-read response":          r.response("tasks-read", "200"),
		"tasks-read-by-index response": r.response("tasks-read-by-index", "200"),
		"tasks-create response":        r.response("tasks-create", "201"),
		"tasks-update response":        r.response("tasks-update", "200"),
		"tasks-create request":         r.request("tasks-create"),
		"tasks-update request":         r.request("tasks-update"),
	}
	var description string
	for where, s := range carriers {
		prop := s.Properties[ruleTextField]
		require.NotNil(t, prop, "%s has no %s", where, ruleTextField)
		assert.Equal(t, huma.TypeString, prop.Type, where)
		assert.False(t, prop.ReadOnly, "%s: %s is writable", where, ruleTextField)
		assert.NotEmpty(t, prop.Description, where)
		if description == "" {
			description = prop.Description
		}
		assert.Equal(t, description, prop.Description, "%s: one description everywhere", where)
	}

	// The field's documented facts, spelled out for MCP tool readers.
	for _, fact := range []string{
		"RFC 5545", "without DTSTART", "BYEASTER", "UTC", "ending in Z",
		"differs from the stored rule", "disagreeing is refused",
		"empty string clears", "null is treated as absent",
		"on PUT, omitting repeat clears the rule",
		"higher COUNT on otherwise unchanged text counts as unchanged",
		"list and bulk endpoints do not",
	} {
		assert.Contains(t, description, fact)
	}

	task := r.doc.Components.Schemas.Map()["Task"]
	require.NotNil(t, task, "the Task component schema")
	assert.NotContains(t, task.Properties, ruleTextField, "models.Task, shared with v1 and the list and bulk endpoints, must not carry the field")
	taskWith := func(extra ...string) []string {
		return slices.Sorted(slices.Values(append(propertyNames(task), extra...)))
	}

	// The write responses are the task plus repeats_rrule and the initial_due_time result
	// (due_date_source, reason), and nothing else: no max_permission. The inputs add the
	// write-only initial_due_time.
	writeResult := []string{ruleTextField, "due_date_source", "reason"}
	assert.Equal(t, taskWith(writeResult...), propertyNames(r.response("tasks-create", "201")))
	assert.Equal(t, taskWith(writeResult...), propertyNames(r.response("tasks-update", "200")))
	assert.Equal(t, taskWith(ruleTextField, "initial_due_time"), propertyNames(r.request("tasks-create")))
	// The read body is also the update input, so AutoPatch's GET then PUT echo validates.
	assert.Equal(t, taskWith("max_permission", ruleTextField, "initial_due_time"), propertyNames(r.response("tasks-read", "200")))
	assert.Equal(t, propertyNames(r.response("tasks-read", "200")), propertyNames(r.request("tasks-update")))

	// Documented scope limit: list and bulk operations do not carry the field.
	for _, id := range []string{"tasks-list", "project-tasks-list", "project-view-tasks-list", "project-view-buckets-tasks-list", "tasks-bulk-update", "tasks-bulk-create"} {
		op := r.op(id)
		if op.RequestBody != nil {
			for mediaType, media := range op.RequestBody.Content {
				assert.False(t, r.mentions(media.Schema, ruleTextField, map[string]bool{}), "%s request (%s)", id, mediaType)
			}
		}
		for status, resp := range op.Responses {
			for mediaType, media := range resp.Content {
				assert.False(t, r.mentions(media.Schema, ruleTextField, map[string]bool{}), "%s %s response (%s)", id, status, mediaType)
			}
		}
	}
}

// TestTaskRuleTextSchema_V1Unchanged guards that models.Task, v1's wire shape, never
// serialises the rule text, whether stored (Repeats) or as sent on v2 (RepeatsText).
func TestTaskRuleTextSchema_V1Unchanged(t *testing.T) {
	text := "FREQ=YEARLY;BYEASTER=1"
	raw, err := json.Marshal(&models.Task{ID: 1, Title: "easter", Repeats: text, RepeatsText: &text})
	require.NoError(t, err)

	var fields map[string]any
	require.NoError(t, json.Unmarshal(raw, &fields))
	for _, key := range []string{ruleTextField, "repeats", "Repeats", "RepeatsText", "repeats_text"} {
		assert.NotContains(t, fields, key)
	}
	assert.NotContains(t, string(raw), "BYEASTER", "the rule text must not leak into v1 JSON")

	// Positive control: the v2 bodies do serialise it.
	raw, err = json.Marshal(newTaskWriteBody(&models.Task{ID: 1, Title: "easter", Repeats: text}, initialDueResult{}))
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"repeats_rrule":"FREQ=YEARLY;BYEASTER=1"`)
	raw, err = json.Marshal(newTaskWriteBody(&models.Task{ID: 1, Title: "plain"}, initialDueResult{}))
	require.NoError(t, err)
	assert.NotContains(t, string(raw), ruleTextField, "absent when the task does not repeat")
}
