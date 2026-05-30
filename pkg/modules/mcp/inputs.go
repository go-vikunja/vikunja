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
	"time"

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
//   - For value-typed src fields, zero values are skipped so partial
//     updates work naturally — only fields the caller actually supplied
//     get propagated. This mirrors the REST update handler's "omitted
//     JSON keys leave the row untouched" behaviour.
//   - For pointer-typed src fields, a nil pointer is treated as "absent"
//     and skipped. A non-nil pointer is dereferenced and assigned even
//     when its pointee is the zero value, so wrappers can explicitly set
//     `false` / `0` / `""` by modelling the field as a pointer.
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
		// A non-nil pointer source is treated as "caller explicitly set
		// this" — even a zero pointee gets propagated so wrappers can
		// clear booleans / numerics. Value-typed sources fall back to
		// the IsZero heuristic for partial-update semantics.
		fromPointer := false
		if srcVal.Kind() == reflect.Pointer {
			if srcVal.IsNil() {
				continue
			}
			srcVal = srcVal.Elem()
			fromPointer = true
		}
		if !fromPointer && srcVal.IsZero() {
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
	// New parent project id. Omit to leave unchanged; pass 0 to move to root.
	ParentProjectID *int64 `json:"parent_project_id,omitempty" jsonschema:"new parent project id; 0 moves to root, omit to leave unchanged"`
	// New ordering position. Omit to leave unchanged; pass 0 to reset.
	Position *float64 `json:"position,omitempty" jsonschema:"new ordering position among siblings; 0 resets to the start, omit to leave unchanged"`
	// Archive state. Omit to leave unchanged.
	IsArchived *bool `json:"is_archived,omitempty" jsonschema:"true to archive, false to un-archive, omit to leave unchanged"`
	// Favorite state for the caller. Omit to leave unchanged.
	IsFavorite *bool `json:"is_favorite,omitempty" jsonschema:"true to favorite for the caller, false to un-favorite, omit to leave unchanged"`
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

// TaskCreateInput is the input wrapper for the `tasks_create` tool.
//
// Only the fields the caller is allowed to set at creation are exposed.
// Server-managed/computed fields (Reminders, Assignees, Labels, Attachments,
// Identifier, Index, Position, IsFavorite, Subscription, Created/Updated,
// CreatedBy(ID), Reactions, RelatedTasks, etc.) are intentionally absent so
// the generated input schema stays narrow.
//
// Title and ProjectID are the only required fields; everything else has
// `omitempty` so the SDK marks them optional.
type TaskCreateInput struct {
	// Title of the task. Required.
	Title string `json:"title" jsonschema:"title of the task"`
	// ID of the project this task belongs to. Required.
	ProjectID int64 `json:"project_id" jsonschema:"id of the project this task belongs to"`
	// Longer-form description (optional).
	Description string `json:"description,omitempty" jsonschema:"longer-form description for the task"`
	// Whether the task is already done at creation time.
	Done bool `json:"done,omitempty" jsonschema:"set to true to create the task in a done state"`
	// When the task is due (RFC 3339 timestamp).
	DueDate time.Time `json:"due_date,omitempty" jsonschema:"due date as an RFC 3339 timestamp"`
	// When the task starts (RFC 3339 timestamp).
	StartDate time.Time `json:"start_date,omitempty" jsonschema:"start date as an RFC 3339 timestamp"`
	// When the task ends (RFC 3339 timestamp).
	EndDate time.Time `json:"end_date,omitempty" jsonschema:"end date as an RFC 3339 timestamp"`
	// Repeat interval in seconds.
	RepeatAfter int64 `json:"repeat_after,omitempty" jsonschema:"repeat interval in seconds"`
	// Repeat mode: 0 = repeat after RepeatAfter, 1 = monthly, 3 = from current date.
	RepeatMode int `json:"repeat_mode,omitempty" jsonschema:"repeat mode: 0 = after interval, 1 = monthly, 3 = from current date"`
	// Priority (sortable, no fixed range).
	Priority int64 `json:"priority,omitempty" jsonschema:"priority value (sortable, caller-defined range)"`
	// PercentDone between 0 and 1.
	PercentDone float64 `json:"percent_done,omitempty" jsonschema:"completion percentage as a float between 0 and 1"`
	// Hex color code (without leading #).
	HexColor string `json:"hex_color,omitempty" jsonschema:"hex color without leading #"`
	// Bucket id (only meaningful when the task is moved into a kanban view).
	BucketID int64 `json:"bucket_id,omitempty" jsonschema:"id of the kanban bucket the task should land in"`
	// ID of the attachment to use as the cover image.
	CoverImageAttachmentID int64 `json:"cover_image_attachment_id,omitempty" jsonschema:"id of the attachment to display as cover image"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.Task.
func (in *TaskCreateInput) ApplyTo(dst handler.CObject) error {
	t, ok := dst.(*models.Task)
	if !ok {
		return fmt.Errorf("mcp: TaskCreateInput.ApplyTo: unexpected destination %T", dst)
	}
	if err := copyByJSONTag(in, t); err != nil {
		return err
	}
	if in.RepeatMode != 0 {
		t.RepeatMode = models.TaskRepeatMode(in.RepeatMode)
	}
	return nil
}

// TaskUpdateInput is the input wrapper for the `tasks_update` tool.
//
// Mirrors TaskCreateInput's writable surface and adds the required ID. Only
// the columns Task.updateSingleTask persists (title, description, done,
// due_date, repeat_after, priority, start_date, end_date, hex_color,
// percent_done, project_id, bucket_id, repeat_mode, cover_image_attachment_id)
// are exposed.
//
// Booleans and numerics whose zero value carries real meaning ("not done",
// "no priority", "0% complete", "no bucket") are modelled as pointers so
// callers can explicitly clear them. A nil pointer means "omit"; a non-nil
// pointer to the zero value means "set to zero".
type TaskUpdateInput struct {
	// ID of the task to update. Required.
	ID int64 `json:"id" jsonschema:"id of the task to update"`
	// New title.
	Title string `json:"title,omitempty" jsonschema:"new title; omit to leave unchanged"`
	// New project id (move the task to a different project).
	ProjectID int64 `json:"project_id,omitempty" jsonschema:"move the task to a different project; omit to leave unchanged"`
	// New description.
	Description string `json:"description,omitempty" jsonschema:"new description; omit to leave unchanged"`
	// Mark the task as done (true) or undone (false). Omit to leave unchanged.
	Done *bool `json:"done,omitempty" jsonschema:"true marks the task as done, false marks it as not done; omit to leave unchanged"`
	// New due date.
	DueDate time.Time `json:"due_date,omitempty" jsonschema:"new due date as an RFC 3339 timestamp"`
	// New start date.
	StartDate time.Time `json:"start_date,omitempty" jsonschema:"new start date as an RFC 3339 timestamp"`
	// New end date.
	EndDate time.Time `json:"end_date,omitempty" jsonschema:"new end date as an RFC 3339 timestamp"`
	// New repeat interval (seconds). Pass 0 to clear.
	RepeatAfter *int64 `json:"repeat_after,omitempty" jsonschema:"new repeat interval in seconds; 0 clears the repeat"`
	// New repeat mode. Pass 0 for the after-interval mode.
	RepeatMode *int `json:"repeat_mode,omitempty" jsonschema:"new repeat mode: 0 = after interval, 1 = monthly, 3 = from current date"`
	// New priority. Pass 0 to clear.
	Priority *int64 `json:"priority,omitempty" jsonschema:"new priority value; 0 clears the priority"`
	// New percent done between 0 and 1. Pass 0 to reset.
	PercentDone *float64 `json:"percent_done,omitempty" jsonschema:"new completion percentage between 0 and 1; 0 resets progress"`
	// New hex color.
	HexColor string `json:"hex_color,omitempty" jsonschema:"new hex color without leading #"`
	// New bucket id (move within a kanban view). Pass 0 to detach.
	BucketID *int64 `json:"bucket_id,omitempty" jsonschema:"new kanban bucket id; 0 detaches from any bucket"`
	// New cover image attachment id. Pass 0 to clear.
	CoverImageAttachmentID *int64 `json:"cover_image_attachment_id,omitempty" jsonschema:"new cover image attachment id; 0 clears the cover"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.Task. ID is always
// copied so the model knows which row to update.
func (in *TaskUpdateInput) ApplyTo(dst handler.CObject) error {
	t, ok := dst.(*models.Task)
	if !ok {
		return fmt.Errorf("mcp: TaskUpdateInput.ApplyTo: unexpected destination %T", dst)
	}
	t.ID = in.ID
	if err := copyByJSONTag(in, t); err != nil {
		return err
	}
	if in.RepeatMode != nil {
		t.RepeatMode = models.TaskRepeatMode(*in.RepeatMode)
	}
	return nil
}

// LabelCreateInput is the input wrapper for the `labels_create` tool.
//
// Label.Create only persists Title, Description, HexColor (plus the
// auto-assigned CreatedBy/ID derived from the authed user), so the wrapper
// exposes exactly those.
type LabelCreateInput struct {
	// Title of the label. Required.
	Title string `json:"title" jsonschema:"title of the label"`
	// Optional longer-form description.
	Description string `json:"description,omitempty" jsonschema:"longer-form description of the label"`
	// Optional hex color (without leading #).
	HexColor string `json:"hex_color,omitempty" jsonschema:"hex color without leading #"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.Label.
func (in *LabelCreateInput) ApplyTo(dst handler.CObject) error {
	l, ok := dst.(*models.Label)
	if !ok {
		return fmt.Errorf("mcp: LabelCreateInput.ApplyTo: unexpected destination %T", dst)
	}
	return copyByJSONTag(in, l)
}

// LabelUpdateInput is the input wrapper for the `labels_update` tool.
//
// Label.Update persists exactly Title, Description, HexColor (see the Cols
// list in pkg/models/label.go). The wrapper exposes those plus the required
// ID.
type LabelUpdateInput struct {
	// ID of the label to update. Required.
	ID int64 `json:"id" jsonschema:"id of the label to update"`
	// New title.
	Title string `json:"title,omitempty" jsonschema:"new title; omit to leave unchanged"`
	// New description.
	Description string `json:"description,omitempty" jsonschema:"new description; omit to leave unchanged"`
	// New hex color (without leading #).
	HexColor string `json:"hex_color,omitempty" jsonschema:"new hex color without leading #; omit to leave unchanged"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.Label. ID is always
// copied so the model knows which row to update.
func (in *LabelUpdateInput) ApplyTo(dst handler.CObject) error {
	l, ok := dst.(*models.Label)
	if !ok {
		return fmt.Errorf("mcp: LabelUpdateInput.ApplyTo: unexpected destination %T", dst)
	}
	l.ID = in.ID
	return copyByJSONTag(in, l)
}

// TeamCreateInput is the input wrapper for the `teams_create` tool.
//
// Team.Create persists Name, Description, IsPublic (plus an auto-assigned
// CreatedByID derived from the authed user). ExternalID and Issuer are
// reserved for SSO/sync flows; we deliberately do not expose them via MCP.
type TeamCreateInput struct {
	// Name of the team. Required.
	Name string `json:"name" jsonschema:"name of the team"`
	// Optional longer-form description.
	Description string `json:"description,omitempty" jsonschema:"longer-form description of the team"`
	// Make the team public (anyone with the URL can see the member list).
	IsPublic bool `json:"is_public,omitempty" jsonschema:"set to true to make the team publicly listable"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.Team.
func (in *TeamCreateInput) ApplyTo(dst handler.CObject) error {
	t, ok := dst.(*models.Team)
	if !ok {
		return fmt.Errorf("mcp: TeamCreateInput.ApplyTo: unexpected destination %T", dst)
	}
	return copyByJSONTag(in, t)
}

// TeamUpdateInput is the input wrapper for the `teams_update` tool.
//
// Team.Update overwrites every column of the row (via xorm s.ID(id).Update),
// so Name/Description/IsPublic round-trip cleanly. The wrapper mirrors the
// same fields plus the required ID.
type TeamUpdateInput struct {
	// ID of the team to update. Required.
	ID int64 `json:"id" jsonschema:"id of the team to update"`
	// New name.
	Name string `json:"name,omitempty" jsonschema:"new team name; omit to leave unchanged"`
	// New description.
	Description string `json:"description,omitempty" jsonschema:"new description; omit to leave unchanged"`
	// New public flag.
	IsPublic bool `json:"is_public,omitempty" jsonschema:"true makes the team publicly listable, false keeps it private"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.Team. ID is always
// copied so the model knows which row to update.
func (in *TeamUpdateInput) ApplyTo(dst handler.CObject) error {
	t, ok := dst.(*models.Team)
	if !ok {
		return fmt.Errorf("mcp: TeamUpdateInput.ApplyTo: unexpected destination %T", dst)
	}
	t.ID = in.ID
	return copyByJSONTag(in, t)
}

// TaskCommentCreateInput is the input wrapper for the
// `tasks_comments_create` tool.
//
// TaskComment.TaskID is `json:"-"` on the model because the REST layer binds
// it from the URL path (`/tasks/:task/comments`). MCP tools take everything as
// JSON args, so the wrapper exposes `task_id` as a required field.
type TaskCommentCreateInput struct {
	// ID of the task to attach the comment to. Required.
	TaskID int64 `json:"task_id" jsonschema:"id of the task the comment belongs to"`
	// The comment text. Required.
	Comment string `json:"comment" jsonschema:"comment body (markdown is supported by the UI but stored verbatim)"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.TaskComment, lifting
// TaskID onto the model field that's otherwise unreachable via JSON.
func (in *TaskCommentCreateInput) ApplyTo(dst handler.CObject) error {
	tc, ok := dst.(*models.TaskComment)
	if !ok {
		return fmt.Errorf("mcp: TaskCommentCreateInput.ApplyTo: unexpected destination %T", dst)
	}
	tc.TaskID = in.TaskID
	tc.Comment = in.Comment
	return nil
}

// TaskCommentReadOneInput is the input wrapper for the
// `tasks_comments_read_one` tool. Both the comment id and the parent task id
// are required: the parent guard inside getTaskCommentSimple rejects requests
// where the comment doesn't belong to the supplied task (IDOR defence).
type TaskCommentReadOneInput struct {
	// ID of the comment to read. Required.
	ID int64 `json:"id" jsonschema:"id of the comment to read"`
	// ID of the parent task. Required.
	TaskID int64 `json:"task_id" jsonschema:"id of the parent task"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.TaskComment.
func (in *TaskCommentReadOneInput) ApplyTo(dst handler.CObject) error {
	tc, ok := dst.(*models.TaskComment)
	if !ok {
		return fmt.Errorf("mcp: TaskCommentReadOneInput.ApplyTo: unexpected destination %T", dst)
	}
	tc.ID = in.ID
	tc.TaskID = in.TaskID
	return nil
}

// TaskCommentReadAllInput is the input wrapper for the
// `tasks_comments_read_all` tool. The parent task id is required (comments
// only make sense scoped to a task); search/page/per_page follow the standard
// pagination contract.
type TaskCommentReadAllInput struct {
	// ID of the parent task. Required.
	TaskID int64 `json:"task_id" jsonschema:"id of the parent task whose comments to list"`
	// Filter comments by substring match.
	Search string `json:"search,omitempty" jsonschema:"filter comments by substring match"`
	// Page (1-based). 0 means server default.
	Page int `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	// Page size. 0 means server default.
	PerPage int `json:"per_page,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// ApplyTo copies TaskID onto the model. Pagination/search are returned via
// ReadAllParams below.
func (in *TaskCommentReadAllInput) ApplyTo(dst handler.CObject) error {
	tc, ok := dst.(*models.TaskComment)
	if !ok {
		return fmt.Errorf("mcp: TaskCommentReadAllInput.ApplyTo: unexpected destination %T", dst)
	}
	tc.TaskID = in.TaskID
	return nil
}

// ReadAllParams exposes search/page/per_page to the dispatcher.
func (in *TaskCommentReadAllInput) ReadAllParams() (search string, page, perPage int) {
	return in.Search, in.Page, in.PerPage
}

// TaskCommentUpdateInput is the input wrapper for the
// `tasks_comments_update` tool. The parent task id is required so the IDOR
// guard inside getTaskCommentSimple can verify the comment belongs to that
// task.
type TaskCommentUpdateInput struct {
	// ID of the comment to update. Required.
	ID int64 `json:"id" jsonschema:"id of the comment to update"`
	// ID of the parent task. Required.
	TaskID int64 `json:"task_id" jsonschema:"id of the parent task"`
	// New comment body. Required (Update only persists this column).
	Comment string `json:"comment" jsonschema:"new comment body"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.TaskComment.
func (in *TaskCommentUpdateInput) ApplyTo(dst handler.CObject) error {
	tc, ok := dst.(*models.TaskComment)
	if !ok {
		return fmt.Errorf("mcp: TaskCommentUpdateInput.ApplyTo: unexpected destination %T", dst)
	}
	tc.ID = in.ID
	tc.TaskID = in.TaskID
	tc.Comment = in.Comment
	return nil
}

// TaskCommentDeleteInput is the input wrapper for the
// `tasks_comments_delete` tool. Both the comment id and parent task id are
// required (the parent guard rejects mismatches).
type TaskCommentDeleteInput struct {
	// ID of the comment to delete. Required.
	ID int64 `json:"id" jsonschema:"id of the comment to delete"`
	// ID of the parent task. Required.
	TaskID int64 `json:"task_id" jsonschema:"id of the parent task"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.TaskComment.
func (in *TaskCommentDeleteInput) ApplyTo(dst handler.CObject) error {
	tc, ok := dst.(*models.TaskComment)
	if !ok {
		return fmt.Errorf("mcp: TaskCommentDeleteInput.ApplyTo: unexpected destination %T", dst)
	}
	tc.ID = in.ID
	tc.TaskID = in.TaskID
	return nil
}

// TaskAssigneeCreateInput is the input wrapper for the
// `tasks_assignees_create` tool. Both task and user IDs are required: TaskID
// identifies the task (REST binds it from `/tasks/:task/assignees`) and
// UserID identifies the user to assign.
type TaskAssigneeCreateInput struct {
	// ID of the task to assign the user to. Required.
	TaskID int64 `json:"task_id" jsonschema:"id of the task to assign the user to"`
	// ID of the user to assign. Required.
	UserID int64 `json:"user_id" jsonschema:"id of the user to assign"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.TaskAssginee
// (note the legacy spelling on the model type).
func (in *TaskAssigneeCreateInput) ApplyTo(dst handler.CObject) error {
	ta, ok := dst.(*models.TaskAssginee)
	if !ok {
		return fmt.Errorf("mcp: TaskAssigneeCreateInput.ApplyTo: unexpected destination %T", dst)
	}
	ta.TaskID = in.TaskID
	ta.UserID = in.UserID
	return nil
}

// TaskAssigneeDeleteInput is the input wrapper for the
// `tasks_assignees_delete` tool. The REST path is
// `/tasks/:task/assignees/:user` — both ids are required.
type TaskAssigneeDeleteInput struct {
	// ID of the task. Required.
	TaskID int64 `json:"task_id" jsonschema:"id of the task"`
	// ID of the user to unassign. Required.
	UserID int64 `json:"user_id" jsonschema:"id of the user to unassign"`
}

// ApplyTo copies the wrapper fields onto a fresh *models.TaskAssginee.
func (in *TaskAssigneeDeleteInput) ApplyTo(dst handler.CObject) error {
	ta, ok := dst.(*models.TaskAssginee)
	if !ok {
		return fmt.Errorf("mcp: TaskAssigneeDeleteInput.ApplyTo: unexpected destination %T", dst)
	}
	ta.TaskID = in.TaskID
	ta.UserID = in.UserID
	return nil
}

// TaskAssigneeReadAllInput is the input wrapper for the
// `tasks_assignees_read_all` tool. The parent task id is required;
// pagination/search follow the standard contract.
type TaskAssigneeReadAllInput struct {
	// ID of the parent task. Required.
	TaskID int64 `json:"task_id" jsonschema:"id of the task whose assignees to list"`
	// Filter assignees by substring match on their username.
	Search string `json:"search,omitempty" jsonschema:"filter assignees by username substring"`
	// Page (1-based). 0 means server default.
	Page int `json:"page,omitempty" jsonschema:"1-based page number; 0 uses the server default"`
	// Page size. 0 means server default.
	PerPage int `json:"per_page,omitempty" jsonschema:"page size; 0 uses the server default"`
}

// ApplyTo copies TaskID onto the model. Pagination is forwarded via
// ReadAllParams below.
func (in *TaskAssigneeReadAllInput) ApplyTo(dst handler.CObject) error {
	ta, ok := dst.(*models.TaskAssginee)
	if !ok {
		return fmt.Errorf("mcp: TaskAssigneeReadAllInput.ApplyTo: unexpected destination %T", dst)
	}
	ta.TaskID = in.TaskID
	return nil
}

// ReadAllParams exposes search/page/per_page to the dispatcher.
func (in *TaskAssigneeReadAllInput) ReadAllParams() (search string, page, perPage int) {
	return in.Search, in.Page, in.PerPage
}
