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
	webfiles "code.vikunja.io/api/pkg/web/files"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
)

// RegisterCardBackgroundRoutes wires the project-card-image actions onto the Huma
// API. It reuses the Unsplash search/proxy endpoints already registered by
// RegisterBackgroundRoutes (they're generic, not tied to the full background) and
// only adds the project-scoped get/delete/upload/unsplash-set endpoints for the
// separate card image field.
func RegisterCardBackgroundRoutes(api huma.API) {
	if !config.BackgroundsEnabled.GetBool() {
		return
	}

	tags := []string{"project"}

	Register(api, huma.Operation{
		OperationID: "projects-card-background-delete",
		Summary:     "Remove a project card image",
		Description: "Removes a project's card image, whichever provider set it. Succeeds even when the project has no card image. Requires write access to the project. Returns the updated project. Does not affect the project's full background.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/card-background",
		// Return the updated project with 200, not the wrapper's DELETE default 204.
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, cardBackgroundRemove)

	Register(api, huma.Operation{
		OperationID: "projects-card-background-get",
		Summary:     "Get a project card image",
		Description: "Streams a project's card image, whichever provider set it. Requires read access to the project. Always served as image/jpeg with a revalidation Last-Modified header, so a conditional If-Modified-Since request gets a 304. Returns 404 when the project has no card image.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/card-background",
		Tags:        tags,
		// Spell out the binary response; the default would be modeled as JSON.
		Responses: map[string]*huma.Response{
			"200": {
				Description: "The project card image as a jpeg image.",
				Content: map[string]*huma.MediaType{
					"image/jpeg": {
						Schema: &huma.Schema{Type: huma.TypeString, Format: "binary"},
					},
				},
			},
		},
	}, cardBackgroundGet)

	if config.BackgroundsUploadEnabled.GetBool() {
		Register(api, huma.Operation{
			OperationID: "projects-card-background-upload",
			Summary:     "Upload a project card image",
			Description: "Uploads an image via multipart/form-data under the \"background\" field and sets it as the project's card image. Requires write access to the project. The image is resized server-side and stored as JPEG; it replaces any previous card image (idempotent replace, hence PUT). Returns the updated project. Does not affect the project's full background.",
			Method:      http.MethodPut,
			Path:        "/projects/{project}/card-backgrounds/upload",
			// Return the updated project with 200, the natural code for an idempotent PUT.
			DefaultStatus: http.StatusOK,
			Tags:          tags,
			// +2 MB mirrors Echo's global BodyLimit overhead so a max-sized file isn't rejected by multipart boundary/header bytes.
			// #nosec G115 - configured value won't exceed int64 max in practice.
			MaxBodyBytes: (int64(config.GetMaxFileSizeInMBytes()) + 2) * 1024 * 1024,
		}, cardBackgroundUpload)
	}

	if config.BackgroundsUnsplashEnabled.GetBool() {
		Register(api, huma.Operation{
			OperationID: "projects-card-background-unsplash-set",
			Summary:     "Set an Unsplash image as project card image",
			Description: "Sets a previously searched Unsplash image (from GET /backgrounds/unsplash/search) as the project's card image, identified by the image id from the search results. Requires write access to the project. Does not affect the project's full background.",
			Method:      http.MethodPut,
			Path:        "/projects/{project}/card-backgrounds/unsplash",
			Tags:        tags,
		}, cardBackgroundUnsplashSet)
	}
}

func init() { AddRouteRegistrar(RegisterCardBackgroundRoutes) }

func cardBackgroundUnsplashSet(ctx context.Context, in *struct {
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
	can, err := project.CanWrite(s, a)
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

	in.Body.Target = background.TargetCard

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

type cardBackgroundUploadInput struct {
	ProjectID int64 `path:"project" doc:"The id of the project to set the card image on."`
	// Allow-list mirrors the formats background uploads can actually be decoded as
	// (handler.ValidateAndSaveBackgroundUpload's allowedImageMimes); octet-stream covers
	// programmatic clients. Huma's MimeTypeValidator rejects the part pre-handler, so the
	// byte-level image check in the shared function is the real gate.
	RawBody huma.MultipartFormFiles[struct {
		Background huma.FormFile `form:"background" contentType:"image/jpeg,image/png,image/gif,image/bmp,image/tiff,image/webp,application/octet-stream" required:"true" doc:"The card image to upload. Must be a decodable raster image (JPEG, PNG, GIF, BMP, TIFF or WebP); it is resized server-side and re-encoded as JPEG."`
	}]
}

// cardBackgroundUpload mirrors backgroundUpload (backgrounds.go): owns auth, the
// session and the permission check because there is no handler.Do* for multipart
// uploads (see the api-v2-routes skill's "Non-CRUDable / custom routes" section).
func cardBackgroundUpload(ctx context.Context, in *cardBackgroundUploadInput) (*singleBody[models.Project], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	project := &models.Project{ID: in.ProjectID}
	can, err := project.CanWrite(s, a)
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

	if err := backgroundHandler.ValidateAndSaveBackgroundUpload(s, a, project, file, file.Filename, uint64(file.Size), background.TargetCard); err != nil {
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

// cardBackgroundGet mirrors backgroundGet (backgrounds.go).
func cardBackgroundGet(ctx context.Context, in *struct {
	ProjectID int64 `path:"project" doc:"The id of the project whose card image to fetch."`
}) (*huma.StreamResponse, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	project := &models.Project{ID: in.ProjectID}
	can, _, err := project.CanRead(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !can {
		_ = s.Rollback()
		return nil, huma.Error403Forbidden("forbidden")
	}

	bgFile, stat, err := backgroundHandler.LoadProjectCardBackgroundForDownload(s, project)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	// The file reader comes from object storage, not the DB session, so it stays
	// valid after the commit; the StreamResponse callback runs after this returns.
	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		// The stream callback (which closes the reader) won't run on this error path.
		_ = bgFile.File.Close()
		return nil, translateDomainError(err)
	}

	return &huma.StreamResponse{Body: func(hctx huma.Context) {
		defer func() { _ = bgFile.File.Close() }()
		c := humaecho.Unwrap(hctx)
		webfiles.WriteProjectBackground((*c).Response(), (*c).Request(), bgFile, stat)
	}}, nil
}

func cardBackgroundRemove(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
}) (*singleBody[models.Project], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	project := &models.Project{ID: in.ProjectID}
	can, err := project.CanWrite(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !can {
		_ = s.Rollback()
		return nil, huma.Error403Forbidden("forbidden")
	}

	if err := project.DeleteCardBackgroundFileIfExists(s); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := models.ClearProjectCardBackground(s, project.ID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &singleBody[models.Project]{Body: project}, nil
}
