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
	"errors"
	"io"
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/userimport"

	"github.com/danielgtaylor/huma/v2"
)

// The user list import section of the people area. Same gate as manage_users.go.

type importCounts struct {
	Created          int `json:"created" doc:"People created, or that would be created."`
	Failed           int `json:"failed" doc:"People that could not be created. They are skipped and tried again on the next run."`
	Updated          int `json:"updated" doc:"Users whose name, job title, department or status changed."`
	Reenabled        int `json:"reenabled" doc:"Deactivated users that were activated again because they are in the list."`
	Disabled         int `json:"disabled" doc:"Users that are, or would be, deactivated because they are not in the list."`
	Unchanged        int `json:"unchanged" doc:"Users that needed no change."`
	Exempt           int `json:"exempt" doc:"Users the import never changes: admins, accounts of other login methods, locked and unconfirmed users."`
	SkippedInvalid   int `json:"skipped_invalid" doc:"Rows without a valid id or mail."`
	SkippedDuplicate int `json:"skipped_duplicate" doc:"Rows that repeat an earlier id or mail."`
}

type importEntry struct {
	Username string `json:"username" doc:"The username. For a person who will be created, the one they would get."`
	Name     string `json:"name" doc:"The display name."`
	Email    string `json:"email" doc:"The email address."`
}

type importDetails struct {
	Create    []importEntry `json:"create" doc:"People who would be created, capped at 200."`
	Update    []importEntry `json:"update" doc:"Users who would be updated, capped at 200."`
	Disable   []importEntry `json:"disable" doc:"Users who would be deactivated, capped at 200."`
	Truncated bool          `json:"truncated" doc:"True if a list was cut off at 200 entries; the counts are complete."`
}

type importLastRun struct {
	Trigger    string        `json:"trigger" enum:"schedule,manual" doc:"What started the run."`
	Actor      string        `json:"actor,omitempty" doc:"The username of the admin who started a manual run."`
	FinishedAt time.Time     `json:"finished_at" doc:"When the run ended."`
	DryRun     bool          `json:"dry_run" doc:"True if the run only reported what it would do."`
	Counts     *importCounts `json:"counts,omitempty" doc:"What the run did. Missing if it did not get as far as planning."`
	Error      string        `json:"error,omitempty" doc:"Why the run did not complete, if it did not."`
}

type importStatus struct {
	Enabled      bool           `json:"enabled" doc:"Whether the scheduled import is switched on (userimport.enabled)."`
	Schedule     string         `json:"schedule" doc:"The cron expression of the scheduled run."`
	DryRun       bool           `json:"dry_run" doc:"Whether the scheduled run only reports (userimport.dryrun). Manual runs always change data."`
	TenantSet    bool           `json:"tenant_set" doc:"Whether userimport.tenantid restricts the import to one Entra tenant."`
	FileExists   bool           `json:"file_exists" doc:"Whether the user list file exists."`
	FileSize     int64          `json:"file_size" doc:"Size of the user list file in bytes."`
	FileModified *time.Time     `json:"file_modified,omitempty" doc:"When the user list file was last modified."`
	Running      bool           `json:"running" doc:"Whether a run is going right now."`
	LastRun      *importLastRun `json:"last_run,omitempty" doc:"The last finished run. Kept in memory only, so it is gone after a restart."`
}

type importPreview struct {
	Counts  importCounts   `json:"counts" doc:"What a run would do."`
	Details *importDetails `json:"details,omitempty" doc:"Who would be affected."`
	Blocked string         `json:"blocked,omitempty" doc:"Set when a run would be refused, for example by the safety limit. The counts still show what was planned."`
}

type importUploadResult struct {
	Rows             int `json:"rows" doc:"Valid rows in the uploaded list."`
	SkippedInvalid   int `json:"skipped_invalid" doc:"Rows without a valid id or mail."`
	SkippedDuplicate int `json:"skipped_duplicate" doc:"Rows that repeat an earlier id or mail."`
}

type importStatusBody struct{ Body importStatus }
type importPreviewBody struct{ Body importPreview }
type importUploadBody struct{ Body importUploadResult }

type importFileInput struct {
	RawBody huma.MultipartFormFiles[struct {
		List huma.FormFile `form:"list" contentType:"text/csv,application/vnd.ms-excel,text/plain,application/octet-stream" required:"true" doc:"The user list as CSV. The header row needs the columns id, mail and displayName; jobTitle, department, userPrincipalName and accountEnabled are optional."`
	}]
}

func toCounts(r *userimport.Result) importCounts {
	return importCounts{
		Created:          r.Created,
		Failed:           r.Failed,
		Updated:          r.Updated,
		Reenabled:        r.Reenabled,
		Disabled:         r.Disabled,
		Unchanged:        r.Unchanged,
		Exempt:           r.Exempt,
		SkippedInvalid:   r.SkippedInvalid,
		SkippedDuplicate: r.SkippedDuplicate,
	}
}

func toEntries(in []userimport.Entry) []importEntry {
	out := make([]importEntry, 0, len(in))
	for _, e := range in {
		out = append(out, importEntry{Username: e.Username, Name: e.Name, Email: e.Email})
	}
	return out
}

func init() { AddRouteRegistrar(RegisterManageUserImportRoutes) }

// RegisterManageUserImportRoutes registers the user list import operations of the people area.
func RegisterManageUserImportRoutes(api huma.API) {
	tags := []string{"manage"}

	Register(api, huma.Operation{
		OperationID: "manage-user-import-status",
		Summary:     "Get the user list import status",
		Description: "Returns the import configuration, the state of the user list file, whether a run is going, and the last finished run. The last run is kept in memory and is lost when the server restarts.",
		Method:      http.MethodGet,
		Path:        "/manage/user-import",
		Tags:        tags,
	}, manageImportStatus)

	op := huma.Operation{
		OperationID:   "manage-user-import-upload",
		Summary:       "Upload a new user list",
		Description:   "Replaces the user list file with an uploaded CSV. The file must have the header columns id, mail and displayName and at least one valid row, otherwise nothing is replaced. It is written next to the target and renamed into place, and the previous file is kept with the suffix .previous. The upload does not run an import.",
		Method:        http.MethodPut,
		Path:          "/manage/user-import/file",
		Tags:          tags,
		DefaultStatus: http.StatusOK,
	}
	op = withUploadLimits(op)
	// Lists are small; do not inherit the generous attachment limit.
	op.MaxBodyBytes = userimport.MaxUploadBytes + 1<<20
	Register(api, op, manageImportUpload)

	Register(api, huma.Operation{
		OperationID:   "manage-user-import-preview",
		Summary:       "Preview a user list import",
		Description:   "Computes what a run would do with the current file and writes nothing, whatever the dry run setting says. Lists at most 200 people each to create, update and deactivate; the counts are complete. The response contains names and email addresses and is not logged. Always do this before a manual run.",
		Method:        http.MethodPost,
		Path:          "/manage/user-import/preview",
		Tags:          tags,
		DefaultStatus: http.StatusOK,
	}, manageImportPreview)

	Register(api, huma.Operation{
		OperationID:   "manage-user-import-run",
		Summary:       "Run the user list import now",
		Description:   "Starts a real import in the background and returns 202. It uses the current file, ignores the dry run setting and the file freshness checks, because you chose the file, but keeps the safety limit that refuses to deactivate too many users at once. Returns 409 while a run is going. Poll the status endpoint for the outcome.",
		Method:        http.MethodPost,
		Path:          "/manage/user-import/run",
		Tags:          tags,
		DefaultStatus: http.StatusAccepted,
	}, manageImportRun)
}

func manageImportStatus(_ context.Context, _ *struct{}) (*importStatusBody, error) {
	st := userimport.CurrentStatus()
	out := importStatus{
		Enabled:    st.Enabled,
		Schedule:   st.Schedule,
		DryRun:     st.DryRun,
		TenantSet:  st.TenantSet,
		FileExists: st.FileExists,
		FileSize:   st.FileSize,
		Running:    st.Running,
	}
	if st.FileExists {
		modified := st.FileModified
		out.FileModified = &modified
	}
	if st.LastRun != nil {
		last := &importLastRun{
			Trigger:    string(st.LastRun.Trigger),
			Actor:      st.LastRun.Actor,
			FinishedAt: st.LastRun.FinishedAt,
			Error:      st.LastRun.Error,
		}
		if st.LastRun.Result != nil {
			counts := toCounts(st.LastRun.Result)
			last.Counts = &counts
			last.DryRun = st.LastRun.Result.DryRun
		}
		out.LastRun = last
	}
	return &importStatusBody{Body: out}, nil
}

func manageImportUpload(ctx context.Context, in *importFileInput) (*importUploadBody, error) {
	doer, err := manageDoer(ctx)
	if err != nil {
		return nil, err
	}

	file := in.RawBody.Data().List
	defer func() { _ = file.Close() }()

	// One byte more than allowed tells an oversized file from one that is exactly at the limit.
	data, err := io.ReadAll(io.LimitReader(file, userimport.MaxUploadBytes+1))
	if err != nil {
		return nil, huma.Error400BadRequest("The file could not be read.")
	}
	if len(data) > userimport.MaxUploadBytes {
		return nil, huma.NewError(http.StatusRequestEntityTooLarge, "The user list is larger than 10 MiB.")
	}

	parsed, err := userimport.ReplaceFile(data)
	if err != nil {
		if errors.Is(err, userimport.ErrInvalidUpload) {
			return nil, huma.Error422UnprocessableEntity(err.Error())
		}
		log.Errorf("[User Import] Could not store the uploaded user list: %s", err)
		return nil, huma.Error500InternalServerError("The user list could not be stored. Check that the directory of userimport.file is writable.")
	}

	dispatchErr := events.DispatchWithContext(ctx, &models.ManageUserImportEvent{Doer: doer, Action: "uploaded"})
	if dispatchErr != nil {
		log.Errorf("Could not dispatch the user import event: %s", dispatchErr)
	}

	return &importUploadBody{Body: importUploadResult{
		Rows:             len(parsed.People),
		SkippedInvalid:   parsed.SkippedInvalid,
		SkippedDuplicate: parsed.SkippedDuplicate,
	}}, nil
}

func manageImportPreview(ctx context.Context, _ *struct{}) (*importPreviewBody, error) {
	result, err := userimport.Preview(ctx)
	if result == nil {
		if errors.Is(err, userimport.ErrAlreadyRunning) {
			return nil, huma.Error409Conflict("An import is running right now.")
		}
		return nil, huma.Error422UnprocessableEntity(err.Error())
	}

	out := importPreview{Counts: toCounts(result)}
	if err != nil {
		// The plan exists but a run would be refused, for example by the safety limit.
		out.Blocked = err.Error()
	}
	if result.Details != nil {
		out.Details = &importDetails{
			Create:    toEntries(result.Details.Create),
			Update:    toEntries(result.Details.Update),
			Disable:   toEntries(result.Details.Disable),
			Truncated: result.Details.Truncated,
		}
	}
	return &importPreviewBody{Body: out}, nil
}

func manageImportRun(ctx context.Context, _ *struct{}) (*messageBody, error) {
	doer, err := manageDoer(ctx)
	if err != nil {
		return nil, err
	}

	err = userimport.StartAsync(doer.Username)
	if err != nil {
		if errors.Is(err, userimport.ErrAlreadyRunning) {
			return nil, huma.Error409Conflict("An import is running right now.")
		}
		return nil, translateDomainError(err)
	}

	dispatchErr := events.DispatchWithContext(ctx, &models.ManageUserImportEvent{Doer: doer, Action: "started"})
	if dispatchErr != nil {
		log.Errorf("Could not dispatch the user import event: %s", dispatchErr)
	}

	out := &messageBody{}
	out.Body.Message = "The import was started."
	return out, nil
}
