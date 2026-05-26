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

// Input wrappers and the wrapper→model adapter.
//
// The SDK's AddTool[In, Out] reflects over the In type's struct tags
// (`json:` for property names, `jsonschema:` for descriptions, omission of
// `omitempty`/`omitzero` for "required") to build the tool's input schema
// via github.com/google/jsonschema-go. We never write a schema by hand.
//
// Wrappers stay in the MCP layer rather than being bolted onto domain
// models: Vikunja models embed dozens of `xorm:"-" json:"..."` computed
// fields (e.g. `Project.Owner`, `Project.MaxPermission`, `Project.Views`)
// that would pollute the input schema if we fed `*models.X{}` directly to
// AddTool. The wrapper is the explicit, narrow shape of "what a caller is
// allowed to specify".
//
// Most resources have symmetric `read_one` and `delete` shapes ({id}) and a
// symmetric `read_all` shape ({search, page, per_page}); those three live
// in this file. Per-resource `<Resource>CreateInput` / `<Resource>UpdateInput`
// land in Task 5/7 next to the resource registrations.
//
// Path-param caveat for Task 7: Vikunja's REST layer binds some fields from
// the URL (e.g. `LabelTask.TaskID` from `/tasks/:task/labels`). MCP tools
// take everything as JSON arguments — there are no URL paths to bind from
// — so a `LabelTaskCreateInput` must include `task_id` as an explicit JSON
// field. The wrapper is the only contract; if the field isn't on the
// wrapper the caller cannot supply it.

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"
)

// ReadOneInput is the shared shape for every `<resource>_read_one` tool.
// Resources whose primary key isn't a top-level `ID int64` field on the
// model must define their own wrapper instead of reusing this one.
type ReadOneInput struct {
	// ID identifies the record to read.
	ID int64 `json:"id"`
}

// ApplyTo writes the wrapper's ID onto the destination model's ID field.
// The destination must be a pointer-to-struct with a top-level field named
// `ID` of type int64 — true for every CRUDable model in pkg/models/ at
// time of writing. If a future resource breaks that assumption it must
// supply its own wrapper.
func (in ReadOneInput) ApplyTo(dst handler.CObject) error {
	return setInt64Field(dst, "ID", in.ID)
}

// DeleteInput is the shared shape for every `<resource>_delete` tool.
type DeleteInput struct {
	// ID identifies the record to delete.
	ID int64 `json:"id"`
}

// ApplyTo writes the wrapper's ID onto the destination model.
func (in DeleteInput) ApplyTo(dst handler.CObject) error {
	return setInt64Field(dst, "ID", in.ID)
}

// ReadAllInput is the shared shape for every `<resource>_read_all` tool.
// Search/page/per_page are forwarded to handler.DoReadAll's positional
// args — they don't live on the model, so ApplyTo is a no-op.
type ReadAllInput struct {
	// Search filters results by case-insensitive substring match on the
	// resource's primary text fields (title, name, etc.).
	Search string `json:"search,omitempty"`
	// Page selects the page of results (1-based). 0 means "server default
	// (first page)", matching the REST layer's behaviour when the query
	// parameter is omitted.
	Page int `json:"page,omitempty"`
	// PerPage selects the page size. 0 means "server default", matching
	// the REST layer.
	PerPage int `json:"per_page,omitempty"`
}

// ApplyTo is a no-op for ReadAllInput. Pagination/search aren't model
// fields; the dispatcher reads them via the readAllInput interface and
// passes them to handler.DoReadAll directly.
func (in ReadAllInput) ApplyTo(_ handler.CObject) error {
	return nil
}

// ReadAllParams returns the pagination/search fields for the dispatcher.
// This is the readAllInput interface declared in dispatcher.go.
func (in ReadAllInput) ReadAllParams() (search string, page, perPage int) {
	return in.Search, in.Page, in.PerPage
}

// setInt64Field locates a top-level field by Go name on the destination
// (which must be a pointer to a struct) and sets it to v. Returns an
// informative error if dst isn't a struct pointer or doesn't have the
// expected field.
//
// Reflection is necessary because handler.CObject is an interface with no
// SetID method — every CRUDable model defines `ID int64` directly. If a
// future resource model breaks that pattern it must supply its own
// wrapper that does the assignment without going through this helper.
func setInt64Field(dst any, fieldName string, v int64) error {
	if dst == nil {
		return errors.New("mcp: cannot set field on nil destination")
	}
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("mcp: destination must be a non-nil pointer, got %s", rv.Kind())
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("mcp: destination must point to a struct, got %s", rv.Kind())
	}
	f := rv.FieldByName(fieldName)
	if !f.IsValid() {
		return fmt.Errorf("mcp: destination type %s has no field %s", rv.Type(), fieldName)
	}
	if !f.CanSet() {
		return fmt.Errorf("mcp: field %s on %s is not settable", fieldName, rv.Type())
	}
	if f.Kind() != reflect.Int64 {
		return fmt.Errorf("mcp: field %s on %s must be int64, got %s", fieldName, rv.Type(), f.Kind())
	}
	f.SetInt(v)
	return nil
}

// copyByJSONTag copies fields from src to dst by matching `json` tag
// names. Used by per-resource wrappers (Task 5/7) to lift writable fields
// onto a fresh model before calling handler.Do*.
//
// Rules:
//   - src may be a struct value or a struct pointer; dst must be a pointer
//     to a struct.
//   - Field matching is by the first segment of the `json` tag (i.e.
//     "title,omitempty" matches "title"). Fields without a json tag (or
//     tagged `json:"-"`) are skipped on both sides.
//   - Zero-valued src fields are skipped, so partial updates work
//     naturally — only fields the caller actually supplied get propagated.
//     This mirrors the REST update handler's "omitted JSON keys leave the
//     row untouched" behaviour. Wrappers that need to clear a field must
//     model it as a pointer (`*string`, `*int`, etc.) so the zero value
//     is distinguishable from "absent".
//   - Pointer src fields are dereferenced. A nil pointer is treated as
//     "absent" and skipped.
//   - Type compatibility: the helper assigns src's value to dst's field
//     when the types are directly assignable. time.Time / *time.Time work
//     out of the box because time.Time is a struct, not a basic type.
//   - Extra fields on src that have no match on dst are silently ignored.
//     Fields on dst that have no match on src are left at their existing
//     value.
func copyByJSONTag(src, dst any) error {
	if src == nil {
		return errors.New("mcp: cannot copy from nil src")
	}
	if dst == nil {
		return errors.New("mcp: cannot copy to nil dst")
	}

	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Pointer || dv.IsNil() {
		return fmt.Errorf("mcp: dst must be a non-nil pointer, got %s", dv.Kind())
	}
	dv = dv.Elem()
	if dv.Kind() != reflect.Struct {
		return fmt.Errorf("mcp: dst must point to a struct, got %s", dv.Kind())
	}

	sv := reflect.ValueOf(src)
	for sv.Kind() == reflect.Pointer {
		if sv.IsNil() {
			return errors.New("mcp: src pointer is nil")
		}
		sv = sv.Elem()
	}
	if sv.Kind() != reflect.Struct {
		return fmt.Errorf("mcp: src must be a struct or pointer-to-struct, got %s", sv.Kind())
	}

	dstFields := jsonTagIndex(dv.Type())

	st := sv.Type()
	for i := 0; i < st.NumField(); i++ {
		sf := st.Field(i)
		if !sf.IsExported() {
			continue
		}
		name, ok := jsonName(sf)
		if !ok {
			continue
		}
		dstIdx, ok := dstFields[name]
		if !ok {
			continue
		}
		srcVal := sv.Field(i)
		// Skip nil pointers (caller didn't supply the field) and
		// dereference non-nil ones.
		if srcVal.Kind() == reflect.Pointer {
			if srcVal.IsNil() {
				continue
			}
			srcVal = srcVal.Elem()
		}
		if srcVal.IsZero() {
			// Zero src value → caller didn't populate this field,
			// leave dst alone.
			continue
		}
		dstVal := dv.Field(dstIdx)
		if !dstVal.CanSet() {
			continue
		}
		if !srcVal.Type().AssignableTo(dstVal.Type()) {
			// Mismatched types: try one level of pointer adjustment
			// on the destination (rare in practice, models tend to
			// store values, not pointers).
			if dstVal.Kind() == reflect.Pointer && srcVal.Type().AssignableTo(dstVal.Type().Elem()) {
				ptr := reflect.New(dstVal.Type().Elem())
				ptr.Elem().Set(srcVal)
				dstVal.Set(ptr)
				continue
			}
			return fmt.Errorf("mcp: cannot assign %s to %s field %s", srcVal.Type(), dstVal.Type(), name)
		}
		dstVal.Set(srcVal)
	}
	return nil
}

// jsonTagIndex returns a name→field-index map for the JSON-tagged fields
// of the given struct type.
func jsonTagIndex(t reflect.Type) map[string]int {
	out := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name, ok := jsonName(f)
		if !ok {
			continue
		}
		out[name] = i
	}
	return out
}

// jsonName extracts the JSON property name from a struct field's `json`
// tag. Returns ("", false) for fields with no tag or tagged "-".
func jsonName(f reflect.StructField) (string, bool) {
	tag := f.Tag.Get("json")
	if tag == "" || tag == "-" {
		return "", false
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" || name == "-" {
		return "", false
	}
	return name, true
}

// ProjectCreateInput is the input wrapper for the `projects_create` tool.
//
// Only the fields the caller is allowed to set are exposed; computed and
// server-managed fields on models.Project (Owner, MaxPermission, Views,
// background information, IsFavorite, etc.) are intentionally absent so the
// generated JSON Schema stays narrow.
//
// Title is the only required field — every other field has `omitempty` so
// the SDK's reflected JSON Schema marks them optional.
type ProjectCreateInput struct {
	// Title of the project. Required.
	Title string `json:"title" jsonschema:"the title of the project"`
	// Optional longer description.
	Description string `json:"description,omitempty" jsonschema:"longer-form description of the project"`
	// Optional short identifier (max 10 chars) used as the prefix for task
	// identifiers within this project.
	Identifier string `json:"identifier,omitempty" jsonschema:"short identifier used as a prefix for task identifiers, max 10 chars"`
	// Optional hex color (without the leading #). Six characters, e.g.
	// "ff0000".
	HexColor string `json:"hex_color,omitempty" jsonschema:"hex color code for the project without leading hash, e.g. ff0000"`
	// Optional parent project id. Zero means top-level.
	ParentProjectID int64 `json:"parent_project_id,omitempty" jsonschema:"id of the parent project, omit or 0 for a top-level project"`
	// Optional ordering position among siblings.
	Position float64 `json:"position,omitempty" jsonschema:"ordering position of the project among its siblings"`
	// Optional archive flag. Defaults to false.
	IsArchived bool `json:"is_archived,omitempty" jsonschema:"set to true to create the project in an archived state"`
	// Optional favorite flag for the calling user. Defaults to false.
	IsFavorite bool `json:"is_favorite,omitempty" jsonschema:"set to true to mark the project as a favorite for the caller"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.Project before
// handler.DoCreate runs. CreateProject overwrites Owner / OwnerID from the
// authed user, so the wrapper does not (and must not) expose those fields.
func (in *ProjectCreateInput) ApplyTo(dst handler.CObject) error {
	p, ok := dst.(*models.Project)
	if !ok {
		return fmt.Errorf("mcp: ProjectCreateInput.ApplyTo: unexpected destination %T", dst)
	}
	return copyByJSONTag(in, p)
}

// ProjectUpdateInput is the input wrapper for the `projects_update` tool.
//
// All writable fields use `omitempty` so callers can supply partial updates;
// copyByJSONTag's "skip zero values" policy leaves omitted fields untouched
// (matching the REST update handler's PATCH-like behaviour). The one
// exception is ID, which is always required to identify the target row.
//
// Vikunja's Project.Update only persists a fixed list of columns (title,
// is_archived, identifier, hex_color, parent_project_id, position, and
// description if non-empty); fields outside that list are silently ignored
// at the model layer. The wrapper exposes exactly that list.
type ProjectUpdateInput struct {
	// ID of the project to update. Required.
	ID int64 `json:"id" jsonschema:"id of the project to update"`
	// New title. Omit to leave unchanged.
	Title string `json:"title,omitempty" jsonschema:"new title for the project; omit to leave unchanged"`
	// New description. Omit to leave unchanged.
	Description string `json:"description,omitempty" jsonschema:"new description; omit to leave unchanged"`
	// New short identifier. Omit to leave unchanged.
	Identifier string `json:"identifier,omitempty" jsonschema:"new short identifier (max 10 chars); omit to leave unchanged"`
	// New hex color (without leading #). Omit to leave unchanged.
	HexColor string `json:"hex_color,omitempty" jsonschema:"new hex color (without leading #); omit to leave unchanged"`
	// New parent project id. Omit (or zero) to leave unchanged.
	ParentProjectID int64 `json:"parent_project_id,omitempty" jsonschema:"new parent project id; omit or 0 to leave unchanged"`
	// New ordering position. Omit (or zero) to leave unchanged.
	Position float64 `json:"position,omitempty" jsonschema:"new ordering position among siblings; omit or 0 to leave unchanged"`
	// Archive state. Omit (or false) to leave un-archived.
	IsArchived bool `json:"is_archived,omitempty" jsonschema:"set to true to archive, omit or false to leave un-archived"`
	// Favorite state for the caller. Omit (or false) to leave un-favorited.
	IsFavorite bool `json:"is_favorite,omitempty" jsonschema:"set to true to favorite for the caller, omit or false to un-favorite"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.Project. ID is
// always copied so the model knows which row to update.
func (in *ProjectUpdateInput) ApplyTo(dst handler.CObject) error {
	p, ok := dst.(*models.Project)
	if !ok {
		return fmt.Errorf("mcp: ProjectUpdateInput.ApplyTo: unexpected destination %T", dst)
	}
	p.ID = in.ID
	return copyByJSONTag(in, p)
}
