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
	"math"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The non-zero starting state lets tests tell "explicitly cleared" from "left untouched".
func applyOnTask(t *testing.T, raw string, start models.Task) *models.Task {
	t.Helper()
	registerAllResources(t)
	spec := specFor(t, "tasks", OpUpdate)
	args, err := validateAndDecodeArgs(spec, json.RawMessage(raw))
	require.NoError(t, err)
	task := start
	require.NoError(t, applyArgs(&task, spec, args))
	return &task
}

func TestApply_PresentZeroValueClears(t *testing.T) {
	// No pointer wrappers: an explicitly sent zero value must overwrite, an
	// omitted key must not.
	start := models.Task{Done: true, Priority: 5, PercentDone: 0.5}

	updated := applyOnTask(t, `{"id":1,"done":false,"priority":0}`, start)
	assert.False(t, updated.Done, "explicit done=false must clear")
	assert.Zero(t, updated.Priority, "explicit priority=0 must clear")
	assert.InDelta(t, 0.5, updated.PercentDone, 1e-9, "omitted percent_done must stay")
	assert.Equal(t, int64(1), updated.ID)
}

func TestApply_OmittedKeysLeaveModelUntouched(t *testing.T) {
	start := models.Task{Title: "keep me", Done: true}

	updated := applyOnTask(t, `{"id":3}`, start)
	assert.Equal(t, "keep me", updated.Title)
	assert.True(t, updated.Done)
}

func TestApply_TypedFields(t *testing.T) {
	updated := applyOnTask(t, `{"id":4,"repeat_mode":1,"due_date":"2026-07-06T10:00:00Z"}`, models.Task{})
	assert.Equal(t, models.TaskRepeatModeMonth, updated.RepeatMode)
	assert.Equal(t, 2026, updated.DueDate.Year())
}

func TestValidate_UnknownArgumentRejected(t *testing.T) {
	registerAllResources(t)
	spec := specFor(t, "tasks", OpUpdate)

	_, err := validateAndDecodeArgs(spec, json.RawMessage(`{"id":1,"nonsense":true}`))
	require.Error(t, err)
}

func TestValidate_MissingRequiredRejected(t *testing.T) {
	registerAllResources(t)

	_, err := validateAndDecodeArgs(specFor(t, "tasks", OpCreate), json.RawMessage(`{"title":"no project"}`))
	require.Error(t, err, "create without project_id must fail validation")

	_, err = validateAndDecodeArgs(specFor(t, "tasks", OpUpdate), json.RawMessage(`{"title":"no id"}`))
	require.Error(t, err, "update without id must fail validation")
}

func TestValidate_WrongTypeRejected(t *testing.T) {
	registerAllResources(t)
	spec := specFor(t, "tasks", OpUpdate)

	_, err := validateAndDecodeArgs(spec, json.RawMessage(`{"id":"not a number"}`))
	require.Error(t, err)
}

func TestValidate_NonObjectArgumentsRejected(t *testing.T) {
	registerAllResources(t)
	spec := specFor(t, "tasks", OpReadAll)

	_, err := validateAndDecodeArgs(spec, json.RawMessage(`[1,2,3]`))
	require.Error(t, err)
}

func TestPopReadAllParams(t *testing.T) {
	config.InitDefaultConfig()
	args := map[string]json.RawMessage{
		argSearch:  json.RawMessage(`"foo"`),
		argPage:    json.RawMessage(`2`),
		argPerPage: json.RawMessage(`50`),
		"filter":   json.RawMessage(`"done = false"`),
	}
	search, page, perPage, err := popReadAllParams(args)
	require.NoError(t, err)
	assert.Equal(t, "foo", search)
	assert.Equal(t, 2, page)
	assert.Equal(t, 50, perPage)
	// Reserved keys are consumed; model-bound keys stay for applyArgs.
	assert.Equal(t, map[string]json.RawMessage{"filter": json.RawMessage(`"done = false"`)}, args)
}

func TestApply_ReadAllFilterLandsOnTaskCollection(t *testing.T) {
	registerAllResources(t)
	spec := specFor(t, "tasks", OpReadAll)

	args, err := validateAndDecodeArgs(spec, json.RawMessage(`{"filter":"done = false","sort_by":["due_date"],"project_id":7}`))
	require.NoError(t, err)
	_, _, _, err = popReadAllParams(args)
	require.NoError(t, err)

	tc := &models.TaskCollection{}
	require.NoError(t, applyArgs(tc, spec, args))
	assert.Equal(t, "done = false", tc.Filter)
	assert.Equal(t, []string{"due_date"}, tc.SortBy)
	assert.Equal(t, int64(7), tc.ProjectID)
}

func TestApply_IntegralFloatsAccepted(t *testing.T) {
	// JSON Schema's "integer" accepts 1.0 and 1e3; json.Unmarshal into an int64 does not.
	updated := applyOnTask(t, `{"id":1.0,"priority":1e3}`, models.Task{})
	assert.Equal(t, int64(1), updated.ID)
	assert.Equal(t, int64(1000), updated.Priority)
}

func TestApply_NonIntegralFloatRejectedCleanly(t *testing.T) {
	// The schema rejects 1.5 for an integer property already; applyArgs must
	// still say so in the same language rather than leaking json internals.
	registerAllResources(t)
	spec := specFor(t, "tasks", OpUpdate)

	err := applyArgs(&models.Task{}, spec, map[string]json.RawMessage{"priority": json.RawMessage(`1.5`)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid value for "priority": must be an integer`)
}

func TestPopReadAllParams_IntegralFloats(t *testing.T) {
	config.InitDefaultConfig()
	args := map[string]json.RawMessage{
		argPage:    json.RawMessage(`1.0`),
		argPerPage: json.RawMessage(`1e1`),
	}
	_, page, perPage, err := popReadAllParams(args)
	require.NoError(t, err)
	assert.Equal(t, 1, page)
	assert.Equal(t, 10, perPage)

	_, _, _, err = popReadAllParams(map[string]json.RawMessage{argPerPage: json.RawMessage(`1.5`)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid value for "per_page": must be an integer`)
}

func TestIntegralNumber_OutOfRangePassedThrough(t *testing.T) {
	// float64(math.MaxInt64) rounds up to 2^63, so a naive `f > math.MaxInt64`
	// guard let 2^63 through and int64(f) wrapped it to MinInt64.
	for _, raw := range []string{
		"9223372036854775808",     // 2^63
		"9223372036854775807.0",   // MaxInt64 written as a float, which rounds to 2^63
		"-9223372036854775809",    // -2^63 - 1
		"9.223372036854775808e18", // 2^63 in exponent form
	} {
		out, err := integralNumber(json.RawMessage(raw))
		require.NoError(t, err, raw)
		assert.JSONEq(t, raw, string(out), "%s must be handed to json.Unmarshal untouched", raw)

		var into int64
		require.Error(t, json.Unmarshal(out, &into), "%s must be reported as out of range", raw)
	}
}

func TestIntegralNumber_InRangeBounds(t *testing.T) {
	for raw, want := range map[string]int64{
		"9223372036854775807":      math.MaxInt64,
		"-9223372036854775808":     math.MinInt64,
		"-9.223372036854775808e18": math.MinInt64,
	} {
		out, err := integralNumber(json.RawMessage(raw))
		require.NoError(t, err, raw)
		var into int64
		require.NoError(t, json.Unmarshal(out, &into), raw)
		assert.Equal(t, want, into, raw)
	}
}

func TestApply_IntegralFloatOnPointerField(t *testing.T) {
	// Project.ParentProjectID is *int64: the integer normalisation has to look
	// through the pointer or json.Unmarshal chokes on 1.0.
	registerAllResources(t)
	spec := specFor(t, "projects", OpUpdate)

	args, err := validateAndDecodeArgs(spec, json.RawMessage(`{"id":1,"parent_project_id":2.0}`))
	require.NoError(t, err)
	p := &models.Project{}
	require.NoError(t, applyArgs(p, spec, args))
	require.NotNil(t, p.ParentProjectID)
	assert.Equal(t, int64(2), *p.ParentProjectID)
}
