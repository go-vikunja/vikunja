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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/background"
	backgroundHandler "code.vikunja.io/api/pkg/modules/background/handler"
	"code.vikunja.io/api/pkg/modules/background/unsplash"

	"github.com/danielgtaylor/huma/v2"
)

type backgroundSearchBody struct {
	Body Paginated[*background.Image]
}

// RegisterBackgroundRoutes wires the project-background actions onto the Huma
// API. BackgroundsEnabled / BackgroundsUnsplashEnabled are static config, so the
// registrar early-returns instead of gating per request.
func RegisterBackgroundRoutes(api huma.API) {
	if !config.BackgroundsEnabled.GetBool() {
		return
	}

	tags := []string{"project"}

	Register(api, huma.Operation{
		OperationID: "projects-background-delete",
		Summary:     "Remove a project background",
		Description: "Removes a project's background, whichever provider set it. Succeeds even when the project has no background. Requires write access to the project. Returns the updated project.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/background",
		// Return the updated project with 200, not the wrapper's DELETE default 204.
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, backgroundRemove)

	if config.BackgroundsUploadEnabled.GetBool() {
		Register(api, huma.Operation{
			OperationID: "projects-background-upload",
			Summary:     "Upload a project background",
			Description: "Uploads an image via multipart/form-data under the \"background\" field and sets it as the project's background. Requires write access to the project. The image is resized server-side and stored as JPEG; it replaces any previous background (idempotent replace, hence PUT). Returns the updated project.",
			Method:      http.MethodPut,
			Path:        "/projects/{project}/backgrounds/upload",
			// Return the updated project with 200, the natural code for an idempotent PUT.
			DefaultStatus: http.StatusOK,
			Tags:          tags,
			// +2 MB mirrors Echo's global BodyLimit overhead so a max-sized file isn't rejected by multipart boundary/header bytes.
			// #nosec G115 - configured value won't exceed int64 max in practice.
			MaxBodyBytes: (int64(config.GetMaxFileSizeInMBytes()) + 2) * 1024 * 1024,
		}, backgroundUpload)
	}

	if config.BackgroundsUnsplashEnabled.GetBool() {
		Register(api, huma.Operation{
			OperationID: "backgrounds-unsplash-search",
			Summary:     "Search Unsplash backgrounds",
			Description: "Searches Unsplash for background images. With an empty query it returns the featured wallpaper collection. Results are paginated by Unsplash; total counts are not available.",
			Method:      http.MethodGet,
			Path:        "/backgrounds/unsplash/search",
			Tags:        tags,
		}, backgroundUnsplashSearch)

		Register(api, huma.Operation{
			OperationID: "projects-background-unsplash-set",
			Summary:     "Set an Unsplash image as project background",
			Description: "Sets a previously searched Unsplash image as the project's background, identified by the image id from the search results. Requires write access to the project.",
			Method:      http.MethodPut,
			Path:        "/projects/{project}/backgrounds/unsplash",
			Tags:        tags,
		}, backgroundUnsplashSet)
	}
}

func init() { AddRouteRegistrar(RegisterBackgroundRoutes) }

func backgroundUnsplashSearch(ctx context.Context, in *struct {
	Q    string `query:"q" doc:"Search query; empty returns the featured wallpaper collection."`
	Page int64  `query:"page" default:"1" minimum:"1" doc:"1-based page number."`
}) (*backgroundSearchBody, error) {
	if _, err := authFromCtx(ctx); err != nil {
		return nil, err
	}

	page := in.Page
	if page < 1 {
		page = 1
	}

	s := db.NewSession()
	defer s.Close()

	p := &unsplash.Provider{}
	result, err := p.Search(s, in.Q, page)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	// Unsplash paginates server-side and p.Search discards the total, so the
	// envelope's total is just this page's length (v1 returned a bare array).
	return &backgroundSearchBody{Body: NewPaginated(result, int64(len(result)), int(page), len(result))}, nil
}

func backgroundUnsplashSet(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	Body      background.Image
}) (*singleBody[models.Project], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	project := &models.Project{ID: in.ProjectID}
	can, err := project.CanUpdate(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !can {
		_ = s.Rollback()
		return nil, huma.Error403Forbidden("forbidden")
	}
	project, err = models.GetProjectSimpleByID(s, in.ProjectID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	p := &unsplash.Provider{}
	if err := p.Set(s, &in.Body, project, a); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := project.ReadOne(s, a); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &singleBody[models.Project]{Body: project}, nil
}

type backgroundUploadInput struct {
	ProjectID int64 `path:"project" doc:"The id of the project to set the background on."`
	// Allow-list mirrors the formats background uploads can actually be decoded as
	// (handler.ValidateAndSaveBackgroundUpload's allowedImageMimes); octet-stream covers
	// programmatic clients. Huma's MimeTypeValidator rejects the part pre-handler, so the
	// byte-level image check in the shared function is the real gate.
	RawBody huma.MultipartFormFiles[struct {
		Background huma.FormFile `form:"background" contentType:"image/jpeg,image/png,image/gif,image/bmp,image/tiff,image/webp,application/octet-stream" required:"true" doc:"The background image to upload. Must be a decodable raster image (JPEG, PNG, GIF, BMP, TIFF or WebP); it is resized server-side and re-encoded as JPEG."`
	}]
}

// backgroundUpload owns auth, the session and the permission check because there is
// no handler.Do* for multipart uploads (see the api-v2-routes skill's "Non-CRUDable
// / custom routes" section). It shares its body with v1 via
// handler.ValidateAndSaveBackgroundUpload.
func backgroundUpload(ctx context.Context, in *backgroundUploadInput) (*singleBody[models.Project], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	project := &models.Project{ID: in.ProjectID}
	can, err := project.CanUpdate(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !can {
		_ = s.Rollback()
		return nil, huma.Error403Forbidden("forbidden")
	}
	project, err = models.GetProjectSimpleByID(s, in.ProjectID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	file := in.RawBody.Data().Background
	defer func() { _ = file.Close() }()

	if err := backgroundHandler.ValidateAndSaveBackgroundUpload(s, a, project, file, file.Filename, uint64(file.Size)); err != nil {
		_ = s.Rollback()
		if backgroundHandler.IsErrFileIsNoImage(err) || backgroundHandler.IsErrFileUnsupportedImageFormat(err) {
			return nil, huma.Error400BadRequest(err.Error())
		}
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &singleBody[models.Project]{Body: project}, nil
}

func backgroundRemove(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
}) (*singleBody[models.Project], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	project := &models.Project{ID: in.ProjectID}
	can, err := project.CanUpdate(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !can {
		_ = s.Rollback()
		return nil, huma.Error403Forbidden("forbidden")
	}

	if err := project.DeleteBackgroundFileIfExists(s); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := models.ClearProjectBackground(s, project.ID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &singleBody[models.Project]{Body: project}, nil
}
