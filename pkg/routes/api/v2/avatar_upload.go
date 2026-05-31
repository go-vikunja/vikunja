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
	"io"
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/avatar"
	"code.vikunja.io/api/pkg/modules/avatar/upload"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gabriel-vasile/mimetype"
)

// avatarUploadInput is the multipart/form-data request for the avatar upload.
// Huma's MultipartFormFiles renders the "avatar" field as a binary file in the
// generated OpenAPI spec; the file bytes are read from in.RawBody.Data().Avatar.
type avatarUploadInput struct {
	// contentType lists the part Content-Types Huma's MimeTypeValidator accepts
	// before our handler runs. Browsers set a real image Content-Type on the
	// part (image/png, image/jpeg, ...) while programmatic clients often send
	// application/octet-stream, so both must be allowed or a legitimate upload
	// would be rejected with a 422 before reaching the handler. This is NOT the
	// security gate: the real, byte-level image check is done in the handler via
	// mimetype.DetectReader (the same allow-list v1 uses); the part Content-Type
	// is client-controlled and must never be trusted on its own.
	RawBody huma.MultipartFormFiles[struct {
		Avatar huma.FormFile `form:"avatar" contentType:"image/png,image/jpeg,image/gif,image/webp,image/svg+xml,application/octet-stream" required:"true" doc:"The avatar image to upload. Must be an image; it is resized server-side and re-encoded as PNG."`
	}]
}

// avatarUploadBody wraps the success message returned after an upload.
type avatarUploadBody struct {
	Body *models.Message
}

// RegisterAvatarRoutes wires the authenticated user's avatar upload onto the Huma API.
func RegisterAvatarRoutes(api huma.API) {
	tags := []string{"user"}

	Register(api, huma.Operation{
		OperationID: "user-avatar-upload",
		Summary:     "Upload your avatar",
		Description: "Uploads an image as the authenticated user's avatar and switches their avatar provider to \"upload\". The image is validated to be an image, resized server-side, and stored as PNG. Replaces any previously uploaded avatar (idempotent replace, hence PUT).",
		Method:      http.MethodPut,
		Path:        "/user/settings/avatar",
		Tags:        tags,
		// Avatars can be larger than Huma's 1 MB default body limit; allow up to
		// the configured max file size so legitimate uploads aren't rejected before
		// the handler runs. Echo's global BodyLimit middleware still caps the total.
		// #nosec G115 - configured value won't exceed int64 max in practice.
		MaxBodyBytes:  int64(config.GetMaxFileSizeInMBytes()) * 1024 * 1024,
		DefaultStatus: http.StatusOK,
	}, avatarUpload)
}

func avatarUpload(ctx context.Context, in *avatarUploadInput) (*avatarUploadBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	// Only real users have avatars; a link share cannot upload one.
	authUser, is := a.(*user.User)
	if !is {
		return nil, huma.Error403Forbidden("only users can upload an avatar")
	}

	s := db.NewSession()
	defer s.Close()

	// Re-fetch the full user so AvatarFileID/Provider are current (the auth
	// user from the JWT claims is partial).
	u, err := user.GetUserByID(s, authUser.ID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	src := in.RawBody.Data().Avatar
	defer func() { _ = src.Close() }()

	// Validate we're dealing with an image (same allow-list as v1's UploadAvatar).
	mime, err := mimetype.DetectReader(src)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !strings.HasPrefix(mime.String(), "image") {
		_ = s.Rollback()
		return nil, huma.Error400BadRequest("Uploaded file is no image.")
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	u.AvatarProvider = "upload"
	if err := upload.StoreAvatarFile(s, u, src); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	avatar.FlushAllCaches(u)

	return &avatarUploadBody{Body: &models.Message{Message: "Avatar was uploaded successfully."}}, nil
}
