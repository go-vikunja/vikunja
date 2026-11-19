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
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// occurrenceReasons is the documented reason list (OccurrenceAnswer.reason), including task_done
// (a done task lists nothing: completing it again changes nothing).
var occurrenceReasons = []string{
	"count_exhausted", "until_passed", "no_rule", "no_due_date", "task_done", "no_occurrence",
	"invalid_rule", "interval_beyond_horizon", "anchor_beyond_horizon", "outside_bounds",
	"policy_rejected", "rule_too_long", "evaluation_failed", "outside_storable_range", "no_zone",
}

// TestTaskOccurrences_CountTagsMatchModel: struct tags cannot name a constant, so the literal
// count bounds of both operations are pinned to the model's here.
func TestTaskOccurrences_CountTagsMatchModel(t *testing.T) {
	for name, typ := range map[string]reflect.Type{
		"tasks-occurrences query":        reflect.TypeFor[taskOccurrencesInput](),
		"tasks-occurrences-preview body": reflect.TypeFor[previewInput](),
	} {
		t.Run(name, func(t *testing.T) {
			field, ok := typ.FieldByName("Count")
			require.True(t, ok)
			assert.Equal(t, "1", field.Tag.Get("minimum"))
			assert.Equal(t, strconv.Itoa(models.MaxOccurrencesCount), field.Tag.Get("maximum"))
			assert.Equal(t, strconv.Itoa(models.DefaultOccurrencesCount), field.Tag.Get("default"))
			assert.Equal(t, fmt.Sprintf("How many occurrences to return, 1 to %d. Defaults to %d.", models.MaxOccurrencesCount, models.DefaultOccurrencesCount), field.Tag.Get("doc"))
		})
	}
}

// TestTaskOccurrences_OpenAPI pins the query's externally visible shape (tasks-occurrences and
// OccurrenceAnswer).
func TestTaskOccurrences_OpenAPI(t *testing.T) {
	r := newRuleTextSchemas(t)
	op := r.op("tasks-occurrences")
	assert.Equal(t, "/tasks/{task}/occurrences", r.pathOf(op))
	assert.Equal(t, []string{"tasks"}, op.Tags)
	assert.Equal(t, "List a task's next occurrences", op.Summary)
	for _, fact := range []string{
		"completed once per date", "the caller, then the project owner, then the server",
		"Read-only", "Requires read access", "reason says why",
		"ends by UNTIL within the list reports no_occurrence",
		"until_passed means UNTIL was already behind the request time",
		"reports task_done", "stayed done because its UNTIL had passed",
		"Without read access the answer is 403 with the same body as a task read",
		"does not exist is 404", "count outside 1 to 100 is 422",
		// The on-time forecast sentence.
		"For a task that repeats from the current date, the list assumes each completion happens on time, " +
			"so the first simulated completion is the later of now and the due date; it is not a promise " +
			"about actual completion.",
	} {
		assert.Contains(t, op.Description, fact)
	}

	params := map[string]*huma.Param{}
	for _, p := range op.Parameters {
		params[p.Name] = p
	}
	require.Contains(t, params, "task")
	assert.Equal(t, "path", params["task"].In)
	assert.Equal(t, huma.TypeInteger, params["task"].Schema.Type)
	count := params["count"]
	require.NotNil(t, count)
	assert.Equal(t, "query", count.In)
	assert.False(t, count.Required)
	assert.Equal(t, huma.TypeInteger, count.Schema.Type)
	require.NotNil(t, count.Schema.Minimum)
	require.NotNil(t, count.Schema.Maximum)
	assert.InDelta(t, 1, *count.Schema.Minimum, 0)
	assert.InDelta(t, models.MaxOccurrencesCount, *count.Schema.Maximum, 0)
	assert.EqualValues(t, models.DefaultOccurrencesCount, count.Schema.Default)
	assert.NotEmpty(t, count.Description)

	assertOccurrenceAnswer(t, r, r.response("tasks-occurrences", "200"))
}

// TestOccurrencePreview_OpenAPI pins the preview's shape (tasks-occurrences-preview and
// PreviewInput): a POST answering 200, both rule forms and
// both anchor forms optional, the zone context fields, and PreviewAnswer.
func TestOccurrencePreview_OpenAPI(t *testing.T) {
	r := newRuleTextSchemas(t)
	op := r.op("tasks-occurrences-preview")
	assert.Equal(t, "/tasks/occurrences/preview", r.pathOf(op))
	assert.Equal(t, []string{"tasks"}, op.Tags)
	assert.Equal(t, "Preview the occurrences of an unsaved rule", op.Summary)
	for _, fact := range []string{
		"without saving anything", "exactly one of repeat or repeats_rrule",
		"refused here with the same error", "the caller's time zone is used, then the server's",
		"400 with code 2002 and the same detail", "unknown timezone is 400 with code 2003",
		"both or neither of repeat and repeats_rrule, or a count outside 1 to 100, is 422",
		"exactly one of due_date or initial_due_time", "due_date_source is then generated, otherwise explicit",
		"due_date is absent, occurrences is empty and reason says why",
		"render [due_date, ...occurrences] in timezone, as returned",
		"with task_id or project_id, the zone a change to a task in that project uses",
		"need read access", "both or neither of due_date and initial_due_time, both task_id and project_id, a negative project_id",
		"an initial_due_time that is not HH:MM from 00:00 to 23:59",
		"the list assumes each completion happens on time", "it is not a promise about actual completion",
		"evaluated_at",
	} {
		assert.Contains(t, op.Description, fact)
	}
	assert.Equal(t, 200, op.DefaultStatus)
	assert.Contains(t, op.Responses, "200")
	assert.NotContains(t, op.Responses, "201")
	require.NotNil(t, op.RequestBody)
	assert.True(t, op.RequestBody.Required)

	body := r.request("tasks-occurrences-preview")
	assert.Equal(t, "PreviewInput", schemaName(op.RequestBody.Content["application/json"].Schema))
	assert.Equal(t, []string{"count", "due_date", "initial_due_time", "project_id", "repeat", "repeats_from_current_date", "repeats_rrule", "task_id", "timezone"}, propertyNames(body))
	assert.Empty(t, body.Required, "due_date is no longer required: exactly one of due_date and initial_due_time")
	assert.Equal(t, huma.TypeString, body.Properties["initial_due_time"].Type)
	assert.Equal(t, "^([01][0-9]|2[0-3]):[0-5][0-9]$", body.Properties["initial_due_time"].Pattern)
	for _, name := range []string{"task_id", "project_id"} {
		assert.Equal(t, huma.TypeInteger, body.Properties[name].Type, name)
		assert.Equal(t, "int64", body.Properties[name].Format, name)
	}
	for name, prop := range body.Properties {
		if name != "$schema" {
			assert.NotEmpty(t, prop.Description, "%s has no description", name)
		}
	}
	assert.Equal(t, huma.TypeString, body.Properties["due_date"].Type)
	assert.Equal(t, "date-time", body.Properties["due_date"].Format)
	assert.Equal(t, huma.TypeString, body.Properties["repeats_rrule"].Type)
	assert.Contains(t, body.Properties["repeats_rrule"].Description, "An empty value is refused here (400, code 2002)")
	assert.Equal(t, huma.TypeString, body.Properties["timezone"].Type)
	assert.Equal(t, huma.TypeBoolean, body.Properties["repeats_from_current_date"].Type)
	assert.Equal(t, false, body.Properties["repeats_from_current_date"].Default)
	count := body.Properties["count"]
	require.NotNil(t, count.Minimum)
	require.NotNil(t, count.Maximum)
	assert.InDelta(t, 1, *count.Minimum, 0)
	assert.InDelta(t, models.MaxOccurrencesCount, *count.Maximum, 0)
	assert.EqualValues(t, models.DefaultOccurrencesCount, count.Default)
	assert.True(t, r.mentions(body.Properties["repeat"], "freq", map[string]bool{}), "repeat is the task's structured recurrence")

	answer := r.response("tasks-occurrences-preview", "200")
	assert.Equal(t, []string{"due_date", "due_date_source", "evaluated_at", "occurrences", "reason", "timezone"}, propertyNames(answer))
	assert.ElementsMatch(t, []string{"occurrences", "timezone", "evaluated_at"}, answer.Required)
	for name, prop := range answer.Properties {
		if name != "$schema" {
			assert.NotEmpty(t, prop.Description, "%s has no description", name)
		}
	}
	occurrences := answer.Properties["occurrences"]
	assert.Equal(t, huma.TypeArray, occurrences.Type)
	require.NotNil(t, occurrences.Items)
	assert.Equal(t, "date-time", occurrences.Items.Format)
	assert.Contains(t, occurrences.Description, "strictly after due_date")
	for _, name := range []string{"due_date", "evaluated_at"} {
		assert.Equal(t, huma.TypeString, answer.Properties[name].Type, name)
		assert.Equal(t, "date-time", answer.Properties[name].Format, name)
	}
	assert.Equal(t, huma.TypeString, answer.Properties["timezone"].Type)
	assert.Equal(t, []any{"explicit", "generated"}, answer.Properties["due_date_source"].Enum)
	reason := answer.Properties["reason"]
	assert.Contains(t, reason.Description, "Present only when fewer than count dates come back")
	assert.Contains(t, reason.Description, "beside a valid due_date")
	for _, code := range occurrenceReasons {
		assert.Contains(t, reason.Description, code)
	}
	assert.NotNil(t, r.doc.Components.Schemas.Map()["PreviewAnswer"])
}

// assertOccurrenceAnswer pins tasks-occurrences' answer: exactly occurrences and reason, with no
// field or status change from the preview work.
func assertOccurrenceAnswer(t *testing.T, r *ruleTextSchemas, answer *huma.Schema) {
	t.Helper()
	assert.Equal(t, []string{"occurrences", "reason"}, propertyNames(answer))
	assert.Equal(t, []string{"occurrences"}, answer.Required)
	occurrences := answer.Properties["occurrences"]
	assert.Equal(t, huma.TypeArray, occurrences.Type)
	require.NotNil(t, occurrences.Items)
	assert.Equal(t, huma.TypeString, occurrences.Items.Type)
	assert.Equal(t, "date-time", occurrences.Items.Format)
	assert.Equal(t, "Due dates in ascending order, at most count.", occurrences.Description)
	reason := answer.Properties["reason"]
	assert.Equal(t, huma.TypeString, reason.Type)
	assert.Contains(t, reason.Description, "Present only when fewer than count dates come back.")
	for _, code := range occurrenceReasons {
		assert.Contains(t, reason.Description, code)
	}
	assert.NotNil(t, r.doc.Components.Schemas.Map()["OccurrenceAnswer"], "the documented schema name")
}

func schemaName(s *huma.Schema) string {
	const prefix = "#/components/schemas/"
	if len(s.Ref) > len(prefix) {
		return s.Ref[len(prefix):]
	}
	return ""
}

// pathOf finds the path an operation is registered on.
func (r *ruleTextSchemas) pathOf(op *huma.Operation) string {
	for path, item := range r.doc.Paths {
		for _, candidate := range []*huma.Operation{item.Get, item.Put, item.Post, item.Patch, item.Delete} {
			if candidate == op {
				return path
			}
		}
	}
	return ""
}

// TestTaskOccurrences_TaskIDZeroIsNotFound is a regression test: GET
// /tasks/0/occurrences panicked, because the optional zone-context helper read task id 0 as "no
// task" and handed a nil task to models.TaskOccurrences. The GET reads its task through
// authorizeTaskRead, which runs tasks-read's CanRead for every id, so an id below 1 gets the same
// 404 as GET /tasks/0 and no task comes back for the handler to pass on. With the trash-aware task
// read, that CanRead looks the missing id up in the trash, so the 404 needs a database:
// TestTaskOccurrences_NonPositiveIDMatchesRead (pkg/webtests) asserts it over HTTP. This test keeps the
// guarantees that need none.
func TestTaskOccurrences_TaskIDZeroIsNotFound(t *testing.T) {
	u := &user.User{ID: 1, Username: "user1"}

	t.Run("the preview's optional context: 0 is none", func(t *testing.T) {
		project, task, err := authorizeZoneContext(nil, u, 0, 0)
		require.NoError(t, err)
		assert.Zero(t, project)
		assert.Nil(t, task, "no context, and the preview never uses the task")
	})

	t.Run("tasks-occurrences reads its task with the required check only", func(t *testing.T) {
		// Structural pin: the GET handler calls authorizeTaskRead, never the optional
		// authorizeZoneContext, whose (0, nil, nil) answer is what reached TaskOccurrences.
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "task_occurrences.go", nil, 0)
		require.NoError(t, err)
		calls := map[string]bool{}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "tasksOccurrences" {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if ident, ok := call.Fun.(*ast.Ident); ok {
						calls[ident.Name] = true
					}
				}
				return true
			})
		}
		assert.True(t, calls["authorizeTaskRead"], "tasksOccurrences calls authorizeTaskRead")
		assert.False(t, calls["authorizeZoneContext"], "tasksOccurrences must not use the optional context helper")
	})
}
