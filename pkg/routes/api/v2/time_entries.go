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
	"context"
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/conditional"
)

// timeTrackingGate is Huma operation middleware that 404s a time-tracking op when the license
// feature is off. It's a middleware because license state can change while the instance is running.
func timeTrackingGate(api huma.API) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !license.IsFeatureEnabled(license.FeatureTimeTracking) {
			_ = huma.WriteErr(api, ctx, http.StatusNotFound, "Not Found")
			return
		}
		next(ctx)
	}
}

func registerGated[I, O any](api huma.API, op huma.Operation, handler func(context.Context, *I) (*O, error)) {
	op.Middlewares = append(op.Middlewares, timeTrackingGate(api))
	Register(api, op, handler)
}

type timeEntryListBody struct {
	Body Paginated[*models.TimeEntry]
}

// RegisterTimeEntryRoutes wires the time-entry CRUD surface onto the Huma API.
func RegisterTimeEntryRoutes(api huma.API) {
	tags := []string{"time-entries"}

	registerGated(api, huma.Operation{
		OperationID: "time-entries-list",
		Summary:     "List time entries",
		Description: "Returns the time entries the authenticated user can see, paginated. Filterable by date range, project, task and user.",
		Method:      http.MethodGet,
		Path:        "/time-entries",
		Tags:        tags,
	}, timeEntriesList)

	registerGated(api, huma.Operation{
		OperationID: "time-entries-read",
		Summary:     "Get a time entry",
		Description: "Returns a single time entry. Sends an ETag; pass it as If-None-Match on a later read to get a 304 Not Modified.",
		Method:      http.MethodGet,
		Path:        "/time-entries/{id}",
		Tags:        tags,
	}, timeEntriesRead)

	registerGated(api, huma.Operation{
		OperationID: "time-entries-create",
		Summary:     "Create a time entry",
		Description: "Logs a manual time entry for the authenticated user. Exactly one of task_id / project_id must be set.",
		Method:      http.MethodPost,
		Path:        "/time-entries",
		Tags:        tags,
	}, timeEntriesCreate)

	registerGated(api, huma.Operation{
		OperationID: "time-entries-update",
		Summary:     "Update a time entry",
		Description: "Updates a time entry. Only the author may update it. The entry can be moved between a task and a project — exactly one of task_id / project_id must be set, and you need read access to the new one. PUT replaces all editable fields; use PATCH for a partial update.",
		Method:      http.MethodPut,
		Path:        "/time-entries/{id}",
		Tags:        tags,
	}, timeEntriesUpdate)

	registerGated(api, huma.Operation{
		OperationID: "time-entries-delete",
		Summary:     "Delete a time entry",
		Description: "Deletes a time entry. Only the author may delete it. If it is the running timer, deleting it removes that timer.",
		Method:      http.MethodDelete,
		Path:        "/time-entries/{id}",
		Tags:        tags,
	}, timeEntriesDelete)

	registerGated(api, huma.Operation{
		OperationID: "task-time-entries-list",
		Summary:     "List a task's time entries",
		Description: "Returns the time entries logged against the given task, across all users, paginated. Scoped to what you can read: an inaccessible or unknown task yields an empty list, not an error.",
		Method:      http.MethodGet,
		Path:        "/tasks/{task_id}/time-entries",
		Tags:        tags,
	}, taskTimeEntriesList)

	registerGated(api, huma.Operation{
		OperationID: "project-time-entries-list",
		Summary:     "List a project's time entries",
		Description: "Returns the time entries for the given project — both standalone project entries and entries on tasks currently in the project — paginated. Scoped to what you can read: an inaccessible or unknown project yields an empty list, not an error.",
		Method:      http.MethodGet,
		Path:        "/projects/{project_id}/time-entries",
		Tags:        tags,
	}, projectTimeEntriesList)

	registerGated(api, huma.Operation{
		OperationID: "time-entries-timer-stop",
		Summary:     "Stop the running timer",
		Description: "Stops the authenticated user's running timer, setting its end time to the server's current time, and returns the stopped entry. Returns 404 when no timer is running. Starting a timer and editing entries go through the regular create/update endpoints.",
		Method:      http.MethodPost,
		Path:        "/time-entries/timer/stop",
		// Override the wrapper's POST→201: this stops an existing entry, it creates nothing.
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, timeEntriesTimerStop)
}

func init() { AddRouteRegistrar(RegisterTimeEntryRoutes) }

// timeEntriesTimerStop is a custom action scoped to the caller: it stops their
// own running timer, so it owns its session and needs no resource permission
// beyond authentication.
func timeEntriesTimerStop(ctx context.Context, _ *struct{}) (*singleBody[models.TimeEntry], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	entry, err := models.StopRunningTimer(s, a)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	events.DispatchPending(s)
	return &singleBody[models.TimeEntry]{Body: entry}, nil
}

func timeEntriesList(ctx context.Context, in *struct {
	ListParams
	Filter         string `query:"filter" doc:"Filter entries with the task filter syntax over user_id, task_id, project_id, start_time and end_time — e.g. \"project_id = 5 && start_time > now-7d\". Use end_time = null to match running timers."`
	FilterTimezone string `query:"filter_timezone" doc:"IANA timezone name used to resolve relative dates (now, now-7d) in the filter, e.g. Europe/Berlin."`
}) (*timeEntryListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	m := &models.TimeEntry{
		Filter:         in.Filter,
		FilterTimezone: in.FilterTimezone,
	}
	result, _, total, err := handler.DoReadAll(ctx, m, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return timeEntriesListResponse(result, total, in.Page, in.PerPage)
}

type timeEntryReadBody struct {
	models.TimeEntry
	MaxPermission models.Permission `json:"max_permission" readOnly:"true" doc:"The maximum permission the requesting user has on this time entry (0=read, 1=read/write, 2=admin)."`
}

func timeEntriesRead(ctx context.Context, in *struct {
	ID int64 `path:"id"`
	conditional.Params
}) (*singleReadBody[timeEntryReadBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	entry := &models.TimeEntry{ID: in.ID}
	maxPermission, err := handler.DoReadOne(ctx, entry, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	body := &timeEntryReadBody{TimeEntry: *entry, MaxPermission: models.Permission(maxPermission)}
	return conditionalReadResponse(&in.Params, body, entry.Updated, maxPermission)
}

func timeEntriesCreate(ctx context.Context, in *struct {
	Body models.TimeEntry
}) (*singleBody[models.TimeEntry], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TimeEntry]{Body: &in.Body}, nil
}

// Body matches the read shape so AutoPatch's GET→PUT echo of max_permission validates.
func timeEntriesUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body timeEntryReadBody
}) (*singleBody[models.TimeEntry], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	entry := &in.Body.TimeEntry
	entry.ID = in.ID // URL wins over body
	if err := handler.DoUpdate(ctx, entry, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TimeEntry]{Body: entry}, nil
}

func timeEntriesDelete(ctx context.Context, in *struct {
	ID int64 `path:"id"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.TimeEntry{ID: in.ID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

func taskTimeEntriesList(ctx context.Context, in *struct {
	TaskID int64 `path:"task_id"`
	ListParams
}) (*timeEntryListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.TimeEntry{TaskID: in.TaskID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return timeEntriesListResponse(result, total, in.Page, in.PerPage)
}

func projectTimeEntriesList(ctx context.Context, in *struct {
	ProjectID int64 `path:"project_id"`
	ListParams
}) (*timeEntryListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.TimeEntry{ProjectID: in.ProjectID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return timeEntriesListResponse(result, total, in.Page, in.PerPage)
}

// timeEntriesListResponse turns the any-typed DoReadAll result into the list
// envelope, hard-failing on a type mismatch (the generic-any silent-empty trap).
func timeEntriesListResponse(result any, total int64, page, perPage int) (*timeEntryListBody, error) {
	items, ok := result.([]*models.TimeEntry)
	if !ok {
		return nil, fmt.Errorf("timeEntries.ReadAll returned unexpected type %T (expected []*models.TimeEntry)", result)
	}
	return &timeEntryListBody{Body: NewPaginated(items, total, page, perPage)}, nil
}
