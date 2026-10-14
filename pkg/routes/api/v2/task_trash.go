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

type trashedTaskListBody struct {
	Body Paginated[*models.Task]
}

func init() { AddRouteRegistrar(RegisterTaskTrashRoutes) }

func RegisterTaskTrashRoutes(api huma.API) {
	tags := []string{"tasks"}

	Register(api, huma.Operation{
		OperationID: "tasks-trash-list",
		Summary:     "List trashed tasks",
		Description: "Returns soft-deleted tasks in non-archived projects the user can write to, newest deletion first. Trashed tasks are purged 30 days after deleted_at. q is ignored. Not available to link shares.",
		Method:      http.MethodGet,
		Path:        "/tasks/deleted",
		Tags:        tags,
	}, tasksTrashList)

	Register(api, huma.Operation{
		OperationID:   "tasks-restore",
		Summary:       "Restore a task from the trash",
		Description:   "Requires write access to the task's project; refused while the project is archived. The task lands in the default bucket (done bucket if done) at the end of every view. Link shares with write access may restore.",
		Method:        http.MethodPost,
		Path:          "/tasks/{task}/restore",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, tasksRestore)

	Register(api, huma.Operation{
		OperationID: "tasks-delete-permanently",
		Summary:     "Permanently delete a trashed task",
		Description: "Purges a task that is already in the trash, with everything attached to it. Its time entries move to the task's project. Requires write access to the project; not available to link shares.",
		Method:      http.MethodDelete,
		Path:        "/tasks/{task}/permanent",
		Tags:        tags,
	}, tasksDeletePermanently)

	Register(api, huma.Operation{
		OperationID: "tasks-trash-empty",
		Summary:     "Empty the task trash",
		Description: "Permanently deletes every task GET /tasks/deleted would return, including tasks other users deleted.",
		Method:      http.MethodDelete,
		Path:        "/tasks/deleted",
		Tags:        tags,
	}, tasksTrashEmpty)
}

func tasksTrashList(ctx context.Context, in *struct {
	ListParams
	Format string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
}) (*trashedTaskListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	tasks, total, err := models.GetTrashedTasks(s, a, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	convertTasksToMarkdown(ctx, tasks...)
	return &trashedTaskListBody{Body: NewPaginated(tasks, total, in.Page, in.PerPage)}, nil
}

func tasksRestore(ctx context.Context, in *struct {
	ID     int64  `path:"task" doc:"The numeric id of the task."`
	Format string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
}) (*singleBody[models.Task], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	defer events.CleanupPending(s)

	task := &models.Task{ID: in.ID}
	can, err := task.CanRestore(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !can {
		_ = s.Rollback()
		return nil, huma.Error403Forbidden("forbidden")
	}
	if err := task.Restore(s, a); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	convertTasksToMarkdown(ctx, task)
	return &singleBody[models.Task]{Body: task}, nil
}

func tasksDeletePermanently(ctx context.Context, in *struct {
	ID int64 `path:"task" doc:"The numeric id of the task."`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	task := &models.Task{ID: in.ID}
	can, err := task.CanDeletePermanently(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !can {
		_ = s.Rollback()
		return nil, huma.Error403Forbidden("forbidden")
	}
	if err := task.DeletePermanently(s); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

func tasksTrashEmpty(ctx context.Context, _ *struct{}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	if _, err := models.EmptyTaskTrash(s, a); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
