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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/humaecho5"
	webfiles "code.vikunja.io/api/pkg/web/files"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

// models.TaskAttachment.ReadAll returns []*models.TaskAttachment.
type taskAttachmentListBody struct {
	Body Paginated[*models.TaskAttachment]
}

type taskAttachmentUploadInput struct {
	TaskID int64 `path:"task" doc:"The id of the task to attach the files to."`
	// Accept any upload; the byte-level mime detection happens in files.CreateWithSession,
	// so there is no part content-type allow-list to enforce here (unlike the avatar endpoint).
	RawBody huma.MultipartFormFiles[struct {
		Files []huma.FormFile `form:"files" required:"true" doc:"One or more files to upload as task attachments. Send multiple parts under the same \"files\" field to upload several at once."`
	}]
}

type taskAttachmentUploadBody struct {
	Body *webfiles.AttachmentUploadResult
}

// RegisterTaskAttachmentRoutes wires task-attachment list/upload/download/delete onto
// the Huma API. The whole resource is gated by the service.enabletaskattachments config
// flag; the check runs here (not at init()) because RegisterAll fires after config loads.
func RegisterTaskAttachmentRoutes(api huma.API) {
	if !config.ServiceEnableTaskAttachments.GetBool() {
		return
	}

	tags := []string{"task"}

	Register(api, huma.Operation{
		OperationID: "task-attachments-list",
		Summary:     "List a task's attachments",
		Description: "Returns the attachment metadata for one task, paginated. Requires read access to the task. The file bytes are not included; fetch them from the download endpoint.",
		Method:      http.MethodGet,
		Path:        "/tasks/{task}/attachments",
		Tags:        tags,
	}, taskAttachmentsList)

	Register(api, huma.Operation{
		OperationID: "task-attachments-upload",
		Summary:     "Upload task attachments",
		Description: "Uploads one or more files as attachments to a task via multipart/form-data under the \"files\" field. Requires write access to the task. Each file is processed independently: a file that fails (for example, exceeding the configured size limit) is reported in the errors list while the others still succeed, so the request returns 201 even on a partial upload. The max size per file is the server's configured file size limit.",
		Method:      http.MethodPost,
		Path:        "/tasks/{task}/attachments",
		Tags:        tags,
		// +2 MB mirrors Echo's global BodyLimit overhead so a max-sized file isn't rejected by multipart boundary/header bytes.
		// #nosec G115 - configured value won't exceed int64 max in practice.
		MaxBodyBytes: (int64(config.GetMaxFileSizeInMBytes()) + 2) * 1024 * 1024,
	}, taskAttachmentsUpload)

	Register(api, huma.Operation{
		OperationID: "task-attachments-download",
		Summary:     "Download a task attachment",
		Description: "Returns the raw bytes of one attachment. Requires read access to the task. Pass preview_size to get a downscaled PNG preview instead — only for image attachments; for non-images or an unknown size the original file is returned. The Content-Type header carries the file's real mime type.",
		Method:      http.MethodGet,
		Path:        "/tasks/{task}/attachments/{attachment}",
		Tags:        tags,
		// Spell out the binary response; a bare []byte Body would otherwise be
		// modeled as a base64 JSON string instead of binary file data.
		Responses: map[string]*huma.Response{
			"200": {
				Description: "The attachment file bytes. The Content-Type header carries the file's mime type.",
				Content: map[string]*huma.MediaType{
					"application/octet-stream": {
						Schema: &huma.Schema{Type: huma.TypeString, Format: "binary"},
					},
				},
			},
		},
	}, taskAttachmentsDownload)

	Register(api, huma.Operation{
		OperationID: "task-attachments-delete",
		Summary:     "Delete a task attachment",
		Description: "Deletes one attachment and its underlying file. Requires write access to the task. The attachment must belong to the task in the path.",
		Method:      http.MethodDelete,
		Path:        "/tasks/{task}/attachments/{attachment}",
		Tags:        tags,
	}, taskAttachmentsDelete)
}

func init() { AddRouteRegistrar(RegisterTaskAttachmentRoutes) }

func taskAttachmentsList(ctx context.Context, in *struct {
	TaskID int64 `path:"task" doc:"The id of the task whose attachments to list."`
	ListParams
}) (*taskAttachmentListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.TaskAttachment{TaskID: in.TaskID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.TaskAttachment)
	if !ok {
		return nil, fmt.Errorf("taskAttachments.ReadAll returned unexpected type %T (expected []*models.TaskAttachment)", result)
	}
	return &taskAttachmentListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

// taskAttachmentsUpload owns auth, the session and the permission check because
// there is no handler.Do* for multipart uploads (see the api-v2-routes skill's
// "Non-CRUDable / custom routes" section).
func taskAttachmentsUpload(ctx context.Context, in *taskAttachmentUploadInput) (*taskAttachmentUploadBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	formFiles := in.RawBody.Data().Files
	uploads := make([]*models.AttachmentToUpload, 0, len(formFiles))
	for _, file := range formFiles {
		uploads = append(uploads, &models.AttachmentToUpload{Reader: file, Filename: file.Filename, Size: uint64(file.Size)})
	}

	success, failures, err := models.UploadTaskAttachments(s, a, in.TaskID, uploads)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	return &taskAttachmentUploadBody{Body: webfiles.BuildUploadResult(success, failures)}, nil
}

// taskAttachmentsDownload owns auth, the session and the permission check; there is
// no handler.Do* for a file body. It loads the attachment, then streams the bytes
// from the StreamResponse callback (no buffering — attachments can be large).
func taskAttachmentsDownload(ctx context.Context, in *struct {
	TaskID       int64  `path:"task" doc:"The id of the task the attachment belongs to."`
	AttachmentID int64  `path:"attachment" doc:"The id of the attachment to download."`
	PreviewSize  string `query:"preview_size" enum:"sm,md,lg,xl" doc:"If set and the attachment is an image, return a downscaled PNG preview instead of the original: sm=100px, md=200px, lg=400px, xl=800px. Ignored for non-image attachments."`
}) (*huma.StreamResponse, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	previewSize := models.GetPreviewSizeFromString(in.PreviewSize)
	ta, preview, err := models.LoadTaskAttachmentForDownload(s, a, in.TaskID, in.AttachmentID, previewSize)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	// The file reader comes from object storage, not the DB session, so it stays
	// valid after the commit; the StreamResponse callback runs after this returns.
	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	return &huma.StreamResponse{Body: func(hctx huma.Context) {
		c := humaecho5.Unwrap(hctx)
		webfiles.WriteAttachmentDownload((*c).Response(), (*c).Request(), ta, preview)
	}}, nil
}

func taskAttachmentsDelete(ctx context.Context, in *struct {
	TaskID       int64 `path:"task" doc:"The id of the task the attachment belongs to."`
	AttachmentID int64 `path:"attachment" doc:"The id of the attachment to delete."`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.TaskAttachment{ID: in.AttachmentID, TaskID: in.TaskID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
