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

type avatarUploadInput struct {
	// Broad allow-list because Huma's MimeTypeValidator rejects the part pre-handler; octet-stream covers programmatic clients. The real gate is mimetype.DetectReader in the handler.
	RawBody huma.MultipartFormFiles[struct {
		Avatar huma.FormFile `form:"avatar" contentType:"image/png,image/jpeg,image/gif,image/webp,image/svg+xml,application/octet-stream" required:"true" doc:"The avatar image to upload. Must be an image; it is resized server-side and re-encoded as PNG."`
	}]
}

type avatarUploadBody struct {
	Body *models.Message
}

func RegisterAvatarRoutes(api huma.API) {
	tags := []string{"user"}

	Register(api, huma.Operation{
		OperationID: "user-avatar-upload",
		Summary:     "Upload your avatar",
		Description: "Uploads an image as the authenticated user's avatar and switches their avatar provider to \"upload\". The image is validated to be an image, resized server-side, and stored as PNG. Replaces any previously uploaded avatar (idempotent replace, hence PUT).",
		Method:      http.MethodPut,
		Path:        "/user/settings/avatar",
		Tags:        tags,
		// +2 MB mirrors Echo's global BodyLimit overhead so a max-sized file isn't rejected by multipart boundary/header bytes.
		// #nosec G115 - configured value won't exceed int64 max in practice.
		MaxBodyBytes:  (int64(config.GetMaxFileSizeInMBytes()) + 2) * 1024 * 1024,
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

	// Re-fetch the full user: the auth user from the JWT claims is partial.
	u, err := user.GetUserByID(s, authUser.ID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	src := in.RawBody.Data().Avatar
	defer func() { _ = src.Close() }()

	// Byte-level image check, same allow-list as v1's UploadAvatar.
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
