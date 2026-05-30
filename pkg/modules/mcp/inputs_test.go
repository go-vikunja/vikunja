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
	"time"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// modelWithID is a minimal CObject used by the ApplyTo tests so we can verify
// ID assignment without standing up a database. The Permissions methods are
// trivial stubs — ApplyTo never invokes them, the dispatcher does.
type modelWithID struct {
	ID int64 `json:"id"`
}

func (m *modelWithID) CanRead(_ *xorm.Session, _ web.Auth) (bool, int, error) {
	return true, 0, nil
}
func (m *modelWithID) CanDelete(_ *xorm.Session, _ web.Auth) (bool, error) { return true, nil }
func (m *modelWithID) CanUpdate(_ *xorm.Session, _ web.Auth) (bool, error) { return true, nil }
func (m *modelWithID) CanCreate(_ *xorm.Session, _ web.Auth) (bool, error) { return true, nil }
func (m *modelWithID) Create(_ *xorm.Session, _ web.Auth) error            { return nil }
func (m *modelWithID) ReadOne(_ *xorm.Session, _ web.Auth) error           { return nil }
func (m *modelWithID) ReadAll(_ *xorm.Session, _ web.Auth, _ string, _, _ int) (any, int, int64, error) {
	return nil, 0, 0, nil
}
func (m *modelWithID) Update(_ *xorm.Session, _ web.Auth) error { return nil }
func (m *modelWithID) Delete(_ *xorm.Session, _ web.Auth) error { return nil }

func TestReadOneInputApplyTo(t *testing.T) {
	m := &modelWithID{}
	in := ReadOneInput{ID: 42}
	require.NoError(t, in.ApplyTo(m))
	assert.Equal(t, int64(42), m.ID)
}

func TestReadOneInputApplyToProject(t *testing.T) {
	// Real model coverage: Project embeds web.CRUDable / web.Permissions but
	// the ID field is still a plain top-level int64. The reflection helper
	// must find it.
	p := &models.Project{}
	in := ReadOneInput{ID: 123}
	require.NoError(t, in.ApplyTo(p))
	assert.Equal(t, int64(123), p.ID)
}

func TestDeleteInputApplyTo(t *testing.T) {
	m := &modelWithID{}
	in := DeleteInput{ID: 7}
	require.NoError(t, in.ApplyTo(m))
	assert.Equal(t, int64(7), m.ID)
}

func TestReadAllInputApplyToIsNoop(t *testing.T) {
	m := &modelWithID{ID: 99}
	in := ReadAllInput{Search: "foo", Page: 3, PerPage: 50}
	require.NoError(t, in.ApplyTo(m))
	// The model was untouched: ApplyTo for ReadAll is a no-op because the
	// pagination/search fields go through DoReadAll's positional args, not
	// the model.
	assert.Equal(t, int64(99), m.ID)
}

func TestReadAllInputReadAllParams(t *testing.T) {
	in := ReadAllInput{Search: "foo", Page: 2, PerPage: 50}
	search, page, perPage := in.ReadAllParams()
	assert.Equal(t, "foo", search)
	assert.Equal(t, 2, page)
	assert.Equal(t, 50, perPage)
}

func TestReadAllInputDefaults(t *testing.T) {
	// Zero values must pass through unchanged — DoReadAll interprets
	// page=0/perPage=0 as "first page / server default", matching the
	// existing REST behaviour when callers omit the query parameters.
	in := ReadAllInput{}
	search, page, perPage := in.ReadAllParams()
	assert.Empty(t, search)
	assert.Zero(t, page)
	assert.Zero(t, perPage)
}

func TestReadOneInputSchema(t *testing.T) {
	s, err := jsonschema.For[ReadOneInput](nil)
	require.NoError(t, err)
	assert.Equal(t, "object", s.Type)
	require.Contains(t, s.Properties, "id")
	assert.Equal(t, "integer", s.Properties["id"].Type)
	assert.Contains(t, s.Required, "id")
}

func TestDeleteInputSchema(t *testing.T) {
	s, err := jsonschema.For[DeleteInput](nil)
	require.NoError(t, err)
	require.Contains(t, s.Properties, "id")
	assert.Contains(t, s.Required, "id")
}

func TestReadAllInputSchema(t *testing.T) {
	s, err := jsonschema.For[ReadAllInput](nil)
	require.NoError(t, err)
	assert.Equal(t, "object", s.Type)
	for _, prop := range []string{"search", "page", "per_page"} {
		assert.Contains(t, s.Properties, prop, "ReadAllInput schema must expose %s", prop)
	}
	// None of the three are required: search/page/per_page all carry
	// omitempty so the SDK treats them as optional.
	assert.NotContains(t, s.Required, "search")
	assert.NotContains(t, s.Required, "page")
	assert.NotContains(t, s.Required, "per_page")
}

// timeSchemaCheck verifies that the bundled jsonschema-go translates time.Time
// fields to {type: string, format: date-time}. That's load-bearing for Task 5
// (project create/update wrappers carry due_date and the like).
func TestTimeFieldSchema(t *testing.T) {
	type withTime struct {
		Due time.Time `json:"due"`
	}
	s, err := jsonschema.For[withTime](nil)
	require.NoError(t, err)
	require.Contains(t, s.Properties, "due")
	assert.Equal(t, "string", s.Properties["due"].Type)
	// The library translates time.Time via the standard library MarshalJSON.
	// Format is set on the *value* schema for time.Time when present.
	// jsonschema-go currently sets only Type=string for time.Time (no format)
	// — both behaviours are acceptable for our use, so we don't assert on
	// the format string.
}

// copyByJSONTag round-trip tests --------------------------------------------

type srcWrapper struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	HexColor    string  `json:"hex_color"`
	Skipped     string  `json:"skipped"`
	Position    float64 `json:"position"`
}

type dstWrapper struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	HexColor    string  `json:"hex_color"`
	Position    float64 `json:"position"`
	// LeftAlone has no matching tag on src; copyByJSONTag must leave it
	// untouched.
	LeftAlone string `json:"left_alone"`
}

func TestCopyByJSONTagBasicFields(t *testing.T) {
	src := srcWrapper{
		Title:       "hello",
		Description: "world",
		HexColor:    "ff0000",
		Skipped:     "ignored",
		Position:    1.5,
	}
	dst := dstWrapper{LeftAlone: "untouched"}
	require.NoError(t, copyByJSONTag(src, &dst))

	assert.Equal(t, "hello", dst.Title)
	assert.Equal(t, "world", dst.Description)
	assert.Equal(t, "ff0000", dst.HexColor)
	assert.InEpsilon(t, 1.5, dst.Position, 0.0001)
	// Field on dst with no matching tag on src stays at its prior value.
	assert.Equal(t, "untouched", dst.LeftAlone)
	// Field on src with no matching tag on dst is silently skipped — no
	// error from copyByJSONTag.
}

func TestCopyByJSONTagSrcAsPointer(t *testing.T) {
	src := &srcWrapper{Title: "ptr-src"}
	dst := dstWrapper{}
	require.NoError(t, copyByJSONTag(src, &dst))
	assert.Equal(t, "ptr-src", dst.Title)
}

func TestCopyByJSONTagDstMustBePointer(t *testing.T) {
	src := srcWrapper{Title: "x"}
	var dst dstWrapper
	err := copyByJSONTag(src, dst)
	require.Error(t, err)
}

func TestCopyByJSONTagSkipsZeroValuesForOptional(t *testing.T) {
	// Optional fields on src that the caller didn't populate (zero value)
	// must not clobber the dst — otherwise PATCH-style update wrappers
	// can't be partial. For Task 4 we keep the policy simple: zero values
	// are skipped. This matches how the REST update handler treats omitted
	// JSON fields.
	src := srcWrapper{Title: "only-title"}
	dst := dstWrapper{
		Title:       "old-title",
		Description: "keep-me",
		HexColor:    "00ff00",
		Position:    9.9,
	}
	require.NoError(t, copyByJSONTag(src, &dst))
	assert.Equal(t, "only-title", dst.Title)
	// Description was zero on src, so dst keeps its existing value.
	assert.Equal(t, "keep-me", dst.Description)
	assert.Equal(t, "00ff00", dst.HexColor)
	assert.InEpsilon(t, 9.9, dst.Position, 0.0001)
}

// TestCopyByJSONTagPointerSrcAllowsZero verifies that pointer-typed src
// fields propagate their pointee even when it's the zero value — this is
// the escape hatch update wrappers use to let callers explicitly set
// `done: false` / `priority: 0` / `is_archived: false`.
func TestCopyByJSONTagPointerSrcAllowsZero(t *testing.T) {
	type ptrSrc struct {
		Done     *bool    `json:"done"`
		Priority *int64   `json:"priority"`
		Position *float64 `json:"position"`
		HexColor *string  `json:"hex_color"`
	}
	type valDst struct {
		Done     bool    `json:"done"`
		Priority int64   `json:"priority"`
		Position float64 `json:"position"`
		HexColor string  `json:"hex_color"`
	}

	falseVal := false
	zeroInt := int64(0)
	zeroFloat := 0.0
	empty := ""
	src := ptrSrc{
		Done:     &falseVal,
		Priority: &zeroInt,
		Position: &zeroFloat,
		HexColor: &empty,
	}
	dst := valDst{
		Done:     true,
		Priority: 5,
		Position: 1.5,
		HexColor: "ff0000",
	}
	require.NoError(t, copyByJSONTag(src, &dst))
	assert.False(t, dst.Done, "non-nil pointer with false pointee must overwrite true")
	assert.Equal(t, int64(0), dst.Priority)
	assert.InDelta(t, 0.0, dst.Position, 0.0001)
	assert.Empty(t, dst.HexColor)
}

// TestCopyByJSONTagNilPointerSrcSkips verifies that nil pointer src fields
// are treated as "absent" — the dst keeps whatever it had.
func TestCopyByJSONTagNilPointerSrcSkips(t *testing.T) {
	type ptrSrc struct {
		Done     *bool  `json:"done"`
		Priority *int64 `json:"priority"`
	}
	type valDst struct {
		Done     bool  `json:"done"`
		Priority int64 `json:"priority"`
	}

	src := ptrSrc{} // both nil
	dst := valDst{Done: true, Priority: 7}
	require.NoError(t, copyByJSONTag(src, &dst))
	assert.True(t, dst.Done, "nil pointer must not overwrite")
	assert.Equal(t, int64(7), dst.Priority)
}

type srcWithPointers struct {
	Title *string    `json:"title"`
	Due   *time.Time `json:"due"`
}

type dstWithTime struct {
	Title string    `json:"title"`
	Due   time.Time `json:"due"`
}

func TestCopyByJSONTagPointerToValue(t *testing.T) {
	title := "from-pointer"
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	src := srcWithPointers{Title: &title, Due: &now}
	dst := dstWithTime{}
	require.NoError(t, copyByJSONTag(src, &dst))
	assert.Equal(t, "from-pointer", dst.Title)
	assert.Equal(t, now, dst.Due)
}

func TestCopyByJSONTagNilPointerSkipped(t *testing.T) {
	dst := dstWithTime{Title: "keep"}
	src := srcWithPointers{Title: nil, Due: nil}
	require.NoError(t, copyByJSONTag(src, &dst))
	// nil src pointer behaves like a zero value — dst is untouched.
	assert.Equal(t, "keep", dst.Title)
	assert.True(t, dst.Due.IsZero())
}

type srcWithValueTime struct {
	Due time.Time `json:"due"`
}

func TestCopyByJSONTagTimeValue(t *testing.T) {
	now := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	src := srcWithValueTime{Due: now}
	dst := dstWithTime{}
	require.NoError(t, copyByJSONTag(src, &dst))
	assert.Equal(t, now, dst.Due)
}

// TestProjectUpdateInputClearsBooleans verifies that a wrapper carrying
// `is_archived: false` (via a non-nil *bool) actually clears IsArchived
// on the destination Project, even when the dst started with IsArchived=true.
// This guards the regression flagged in PR review: prior to the pointer-source
// fix, all zero values were silently dropped by copyByJSONTag.
func TestProjectUpdateInputClearsBooleans(t *testing.T) {
	falseVal := false
	in := &ProjectUpdateInput{ID: 1, IsArchived: &falseVal, IsFavorite: &falseVal}
	p := &models.Project{ID: 1, IsArchived: true, IsFavorite: true}
	require.NoError(t, in.ApplyTo(p))
	assert.False(t, p.IsArchived, "IsArchived must clear when explicitly set to false")
	assert.False(t, p.IsFavorite, "IsFavorite must clear when explicitly set to false")
}

// TestTaskUpdateInputClearsBoolsAndZeros mirrors the project test for tasks
// — done can flip to false, priority can drop to 0, percent_done resets.
func TestTaskUpdateInputClearsBoolsAndZeros(t *testing.T) {
	falseVal := false
	zeroPriority := int64(0)
	zeroPercent := 0.0
	in := &TaskUpdateInput{
		ID:          1,
		Done:        &falseVal,
		Priority:    &zeroPriority,
		PercentDone: &zeroPercent,
	}
	tk := &models.Task{
		ID:          1,
		Done:        true,
		Priority:    5,
		PercentDone: 0.75,
	}
	require.NoError(t, in.ApplyTo(tk))
	assert.False(t, tk.Done)
	assert.Equal(t, int64(0), tk.Priority)
	assert.InDelta(t, 0.0, tk.PercentDone, 0.0001)
}
