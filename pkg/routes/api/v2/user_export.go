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
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/humaecho5"
	"code.vikunja.io/api/pkg/user"
	webfiles "code.vikunja.io/api/pkg/web/files"

	"github.com/danielgtaylor/huma/v2"
	"xorm.io/xorm"
)

type userExportPasswordBody struct {
	Body struct {
		Password string `json:"password" doc:"The authenticated user's password. Required for local users; ignored for users authenticated via an external provider."`
	}
}

type userExportStatusBody struct {
	Body *models.UserExportStatus
}

func RegisterUserExportRoutes(api huma.API) {
	tags := []string{"user"}

	Register(api, huma.Operation{
		OperationID:   "user-export-request",
		Summary:       "Request a data export",
		Description:   "Starts building a full export of the authenticated user's data. Local users must confirm with their password. The export runs in the background; an email is sent when it is ready to download.",
		Method:        http.MethodPost,
		Path:          "/user/export/request",
		Tags:          tags,
		DefaultStatus: http.StatusOK,
	}, userExportRequest)

	Register(api, huma.Operation{
		OperationID: "user-export-download",
		Summary:     "Download the data export",
		Description: "Streams the authenticated user's prepared data export as a zip file. Local users must confirm with their password. Fails with 404 if no export has been prepared. A POST (not GET) because the password is sent in the body.",
		Method:      http.MethodPost,
		Path:        "/user/export/download",
		Tags:        tags,
		// Spell out the binary response; the default would be modeled as JSON.
		Responses: map[string]*huma.Response{
			"200": {
				Description: "The data export as a zip file.",
				Content: map[string]*huma.MediaType{
					"application/zip": {
						Schema: &huma.Schema{Type: huma.TypeString, Format: "binary"},
					},
				},
			},
		},
	}, userExportDownload)

	Register(api, huma.Operation{
		OperationID: "user-export-status",
		Summary:     "Get the current data export",
		Description: "Returns metadata about the authenticated user's current data export (id, size, creation and expiry time), or null if none has been prepared.",
		Method:      http.MethodGet,
		Path:        "/user/export",
		Tags:        tags,
	}, userExportStatus)
}

func init() { AddRouteRegistrar(RegisterUserExportRoutes) }

// confirmExportPassword resolves the full DB user and, for local accounts, verifies
// the supplied password — mirroring v1's checkExportRequest. External-provider users
// cannot supply a password and are passed through, as in v1.
func confirmExportPassword(ctx context.Context, s *xorm.Session, password string) (*user.User, error) {
	u, err := authUserFromCtx(ctx, s)
	if err != nil {
		return nil, err
	}
	if u.IsLocalUser() {
		if err := user.CheckUserPassword(u, password); err != nil {
			return nil, translateDomainError(err)
		}
	}
	return u, nil
}

func userExportRequest(ctx context.Context, in *userExportPasswordBody) (*messageBody, error) {
	s := db.NewSession()
	defer s.Close()

	u, err := confirmExportPassword(ctx, s, in.Body.Password)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	events.DispatchOnCommit(s, &models.UserDataExportRequestedEvent{User: u})

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)

	out := &messageBody{}
	out.Body.Message = "Successfully requested data export. We will send you an email when it's ready."
	return out, nil
}

func userExportDownload(ctx context.Context, in *userExportPasswordBody) (*huma.StreamResponse, error) {
	s := db.NewSession()
	defer s.Close()

	u, err := confirmExportPassword(ctx, s, in.Body.Password)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	exportFile, err := models.GetUserDataExportFile(u)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	// The file reader comes from object storage, not the DB session, so it stays
	// valid after the commit; the StreamResponse callback runs after this returns.
	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		// The stream callback (which closes the reader) won't run on this error path.
		_ = exportFile.File.Close()
		return nil, translateDomainError(err)
	}

	return &huma.StreamResponse{Body: func(hctx huma.Context) {
		defer func() { _ = exportFile.File.Close() }()
		c := humaecho5.Unwrap(hctx)
		webfiles.WriteFileDownload((*c).Response(), (*c).Request(), exportFile)
	}}, nil
}

func userExportStatus(ctx context.Context, _ *struct{}) (*userExportStatusBody, error) {
	s := db.NewSession()
	defer s.Close()

	u, err := authUserFromCtx(ctx, s)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	status, err := models.GetUserDataExportStatus(u)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &userExportStatusBody{Body: status}, nil
}
