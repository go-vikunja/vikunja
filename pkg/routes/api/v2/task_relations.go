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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

// RegisterTaskRelationRoutes wires task-relation create/delete onto the Huma API.
//
// Both operations reuse handler.DoCreate/DoDelete; CanCreate enforces write on
// the base task + read on the other task and rejects invalid kinds, CanDelete
// enforces write on the base task. The only custom part is mapping the path
// segments onto the model.
func RegisterTaskRelationRoutes(api huma.API) {
	tags := []string{"tasks"}

	Register(api, huma.Operation{
		OperationID: "tasks-relations-create",
		Summary:     "Create a task relation",
		Description: "Relates two tasks. The authenticated user needs write access to the base task (in the path) and at least read access to the other task; the two tasks need not share a project. The inverse relation is created automatically (e.g. a subtask relation also stores the parenttask relation on the other task). Subtask/parenttask chains that would form a cycle are rejected.",
		Method:      http.MethodPost,
		Path:        "/tasks/{task}/relations",
		Tags:        tags,
	}, tasksRelationsCreate)

	Register(api, huma.Operation{
		OperationID: "tasks-relations-delete",
		Summary:     "Delete a task relation",
		Description: "Removes the relation identified by the base task, relation kind and other task. The automatically created inverse relation is removed as well. The authenticated user needs write access to the base task.",
		Method:      http.MethodDelete,
		Path:        "/tasks/{task}/relations/{relationKind}/{otherTask}",
		Tags:        tags,
	}, tasksRelationsDelete)
}

func init() { AddRouteRegistrar(RegisterTaskRelationRoutes) }

func tasksRelationsCreate(ctx context.Context, in *struct {
	TaskID int64 `path:"task" doc:"The numeric id of the base task to relate from."`
	Body   models.TaskRelation
}) (*singleBody[models.TaskRelation], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	rel := &in.Body
	rel.TaskID = in.TaskID // URL wins over body
	if err := handler.DoCreate(ctx, rel, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TaskRelation]{Body: rel}, nil
}

// The relationKind enum mirrors models.TaskRelation.RelationKind's tag (see the sync note there).
func tasksRelationsDelete(ctx context.Context, in *struct {
	TaskID       int64               `path:"task" doc:"The numeric id of the base task."`
	RelationKind models.RelationKind `path:"relationKind" enum:"subtask,parenttask,related,duplicateof,duplicates,blocking,blocked,precedes,follows,copiedfrom,copiedto" doc:"The kind of the relation to remove."`
	OtherTaskID  int64               `path:"otherTask" doc:"The numeric id of the other task in the relation."`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	rel := &models.TaskRelation{
		TaskID:       in.TaskID,
		RelationKind: in.RelationKind,
		OtherTaskID:  in.OtherTaskID,
	}
	if err := handler.DoDelete(ctx, rel, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
