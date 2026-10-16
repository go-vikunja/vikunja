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

	"github.com/danielgtaylor/huma/v2"
)

func init() { AddRouteRegistrar(RegisterProjectTrashRoutes) }

func RegisterProjectTrashRoutes(api huma.API) {
	tags := []string{"projects"}

	Register(api, huma.Operation{
		OperationID: "projects-trash-list",
		Summary:     "List trashed projects",
		Description: "Returns trashed projects the user is admin of, newest deletion first. A trashed sub-project is folded into its trashed ancestor and counted in trashed_sub_projects. Each project is purged 30 days after its own deleted_at. q is ignored. Not available to link shares.",
		Method:      http.MethodGet,
		Path:        "/projects/deleted",
		Tags:        tags,
	}, projectsTrashList)

	Register(api, huma.Operation{
		OperationID:   "projects-restore",
		Summary:       "Restore a project from the trash",
		Description:   "Restores the project with all its trashed sub-projects, whenever they were deleted. Tasks trashed on their own stay in the trash. Requires admin access; refused while the parent project is in the trash or archived.",
		Method:        http.MethodPost,
		Path:          "/projects/{id}/restore",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, projectsRestore)

	Register(api, huma.Operation{
		OperationID: "projects-delete-permanently",
		Summary:     "Permanently delete a trashed project",
		Description: "Purges a project that is already in the trash, with all its sub-projects and tasks. Time entries survive without a project, visible only to their authors. Requires admin access.",
		Method:      http.MethodDelete,
		Path:        "/projects/{id}/permanent",
		Tags:        tags,
	}, projectsDeletePermanently)

	Register(api, huma.Operation{
		OperationID: "projects-trash-empty",
		Summary:     "Empty the project trash",
		Description: "Permanently deletes every project GET /projects/deleted would return, including projects other users deleted.",
		Method:      http.MethodDelete,
		Path:        "/projects/deleted",
		Tags:        tags,
	}, projectsTrashEmpty)

	Register(api, huma.Operation{
		OperationID: "projects-count",
		Summary:     "Count a project's contents",
		Description: "Counts the live sub-projects, tasks and time entries of a project and everything below it, e.g. to tell what deleting it moves to the trash. Requires admin access.",
		Method:      http.MethodGet,
		Path:        "/projects/{id}/count",
		Tags:        tags,
	}, projectsCount)
}

func projectsTrashList(ctx context.Context, in *ListParams) (*projectListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	projects, total, err := models.GetTrashedProjects(s, a, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &projectListBody{Body: NewPaginated(projects, total, in.Page, in.PerPage)}, nil
}

func projectsRestore(ctx context.Context, in *struct {
	ID int64 `path:"id" doc:"The numeric id of the project."`
}) (*singleBody[models.Project], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	defer events.CleanupPending(s)

	project := &models.Project{ID: in.ID}
	can, err := project.CanRestore(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !can {
		_ = s.Rollback()
		return nil, huma.Error403Forbidden("forbidden")
	}
	if err := project.Restore(s, a); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	return &singleBody[models.Project]{Body: project}, nil
}

func projectsDeletePermanently(ctx context.Context, in *struct {
	ID int64 `path:"id" doc:"The numeric id of the project."`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	project := &models.Project{ID: in.ID}
	can, err := project.CanDeletePermanently(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !can {
		_ = s.Rollback()
		return nil, huma.Error403Forbidden("forbidden")
	}
	if err := project.DeletePermanently(s); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

func projectsTrashEmpty(ctx context.Context, _ *struct{}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	if _, err := models.EmptyProjectTrash(s, a); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

func projectsCount(ctx context.Context, in *struct {
	ID int64 `path:"id" doc:"The numeric id of the project."`
}) (*singleBody[models.ProjectCount], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewReadSession()
	defer s.Close()

	count, err := models.GetProjectCount(s, a, in.ID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.ProjectCount]{Body: count}, nil
}
