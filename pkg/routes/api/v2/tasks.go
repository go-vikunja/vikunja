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
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/conditional"
)

// expandDoc lists the accepted expand values; shared between the by-id and
// by-index operations so the docs stay in sync.
const expandDoc = "Embed extra, more expensive data in each task. Repeatable. One of: subtasks, buckets, reactions, comments, comment_count, time_entries_count, is_unread. Expanding can return more tasks than the page limit (subtasks) and inflate the response."

// parseTaskExpand turns the raw `expand` query values into validated
// TaskCollectionExpandable entries. Kept package-level for the TaskCollection
// list endpoint, which accepts the same option. An invalid value returns the
// model's own validation error, which translateDomainError maps to 422.
func parseTaskExpand(raw []string) ([]models.TaskCollectionExpandable, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	expand := make([]models.TaskCollectionExpandable, 0, len(raw))
	for _, e := range raw {
		v := models.TaskCollectionExpandable(e)
		if err := v.Validate(); err != nil {
			return nil, err
		}
		expand = append(expand, v)
	}
	return expand, nil
}

// RegisterTaskRoutes wires Task CRUD onto the Huma API. The list lives on
// TaskCollection, not here.
func RegisterTaskRoutes(api huma.API) {
	tags := []string{"tasks"}

	Register(api, huma.Operation{
		OperationID: "tasks-read",
		Summary:     "Get a task",
		Description: "Returns a single task by its numeric id. Sends an ETag; pass it as If-None-Match on a later read to get a 304 Not Modified. " + expandDoc,
		Method:      "GET",
		Path:        "/tasks/{projecttask}",
		Tags:        tags,
	}, tasksRead)

	Register(api, huma.Operation{
		OperationID: "tasks-read-by-index",
		Summary:     "Get a task by its project index",
		Description: "Returns a single task addressed by its per-project index. The {project} segment accepts either a numeric project id or a textual project identifier (e.g. \"PROJ\"); a value made solely of digits is always treated as an id. " + expandDoc,
		Method:      "GET",
		Path:        "/projects/{project}/tasks/by-index/{index}",
		Tags:        tags,
	}, tasksReadByIndex)

	Register(api, huma.Operation{
		OperationID: "tasks-create",
		Summary:     "Create a task",
		Description: "Creates a task in the project from the URL. The authenticated user needs write access to that project and becomes the task's creator.",
		Method:      "POST",
		Path:        "/projects/{project}/tasks",
		Tags:        tags,
	}, tasksCreate)

	Register(api, huma.Operation{
		OperationID: "tasks-update",
		Summary:     "Update a task",
		Description: "Replaces all of a task's fields; requires write access. Setting project_id to a different project moves the task and also requires write access to the target project. Use PATCH for a partial update.",
		Method:      "PUT",
		Path:        "/tasks/{projecttask}",
		Tags:        tags,
	}, tasksUpdate)

	Register(api, huma.Operation{
		OperationID: "tasks-delete",
		Summary:     "Delete a task",
		Description: "Deletes a task. Requires write access to its project.",
		Method:      "DELETE",
		Path:        "/tasks/{projecttask}",
		Tags:        tags,
	}, tasksDelete)
}

func init() { AddRouteRegistrar(RegisterTaskRoutes) }

type taskReadOneBody struct {
	models.Task
	MaxPermission models.Permission `json:"max_permission" readOnly:"true" doc:"The maximum permission the requesting user has on this task (0=read, 1=read/write, 2=admin)."`
}

func tasksRead(ctx context.Context, in *struct {
	ID     int64    `path:"projecttask" doc:"The numeric id of the task."`
	Expand []string `query:"expand,explode" enum:"subtasks,buckets,reactions,comments,comment_count,time_entries_count,is_unread" doc:"Embed extra data per task. Repeatable."`
	Format string   `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	conditional.Params
}) (*singleReadBody[taskReadOneBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	expand, err := parseTaskExpand(in.Expand)
	if err != nil {
		return nil, translateDomainError(err)
	}
	task := &models.Task{ID: in.ID, Expand: expand}
	maxPermission, err := handler.DoReadOne(ctx, task, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	body := &taskReadOneBody{Task: *task, MaxPermission: models.Permission(maxPermission)}
	convertTasksToMarkdown(ctx, &body.Task)
	return conditionalReadResponse(&in.Params, body, task.Updated, maxPermission)
}

func tasksReadByIndex(ctx context.Context, in *struct {
	Project string   `path:"project" doc:"A numeric project id or a textual project identifier (e.g. \"PROJ\")."`
	Index   int64    `path:"index" doc:"The per-project task index."`
	Expand  []string `query:"expand,explode" enum:"subtasks,buckets,reactions,comments,comment_count,time_entries_count,is_unread" doc:"Embed extra data per task. Repeatable."`
	Format  string   `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	conditional.Params
}) (*singleReadBody[taskReadOneBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	expand, err := parseTaskExpand(in.Expand)
	if err != nil {
		return nil, translateDomainError(err)
	}
	projectID, err := resolveProjectIdentifier(in.Project)
	if err != nil {
		return nil, err
	}

	// ID 0 + ProjectID + Index makes the model resolve the id from the
	// (project, index) pair in both CanRead and ReadOne.
	task := &models.Task{ProjectID: projectID, Index: in.Index, Expand: expand}
	maxPermission, err := handler.DoReadOne(ctx, task, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	body := &taskReadOneBody{Task: *task, MaxPermission: models.Permission(maxPermission)}
	convertTasksToMarkdown(ctx, &body.Task)
	return conditionalReadResponse(&in.Params, body, task.Updated, maxPermission)
}

func tasksCreate(ctx context.Context, in *struct {
	Project int64  `path:"project" doc:"The numeric id of the project to create the task in."`
	Format  string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	Body    models.Task
}) (*singleBody[models.Task], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	task := &in.Body
	task.ProjectID = in.Project // URL wins over body
	if err := convertToHTML(ctx, &task.Description); err != nil {
		return nil, translateDomainError(err)
	}
	if err := handler.DoCreate(ctx, task, a); err != nil {
		return nil, translateDomainError(err)
	}
	convertTasksToMarkdown(ctx, task)
	return &singleBody[models.Task]{Body: task}, nil
}

// Body matches the read shape so AutoPatch's GET→PUT echo of max_permission validates.
func tasksUpdate(ctx context.Context, in *struct {
	ID     int64  `path:"projecttask"`
	Format string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	Body   taskReadOneBody
}) (*singleBody[models.Task], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	task := &in.Body.Task
	task.ID = in.ID // URL wins over body
	if err := convertToHTML(ctx, &task.Description); err != nil {
		return nil, translateDomainError(err)
	}
	if err := handler.DoUpdate(ctx, task, a); err != nil {
		return nil, translateDomainError(err)
	}
	convertTasksToMarkdown(ctx, task)
	return &singleBody[models.Task]{Body: task}, nil
}

func tasksDelete(ctx context.Context, in *struct {
	ID int64 `path:"projecttask"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.Task{ID: in.ID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

// resolveProjectIdentifier turns the {project} path segment into a numeric
// project id. A pure-digit value is always an id (mirroring v1's
// ResolveProjectIdentifier middleware); anything else is looked up as a
// case-insensitive identifier and 404s if unknown.
func resolveProjectIdentifier(raw string) (int64, error) {
	if id, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return id, nil
	}
	s := db.NewSession()
	defer s.Close()
	project, err := models.GetProjectSimpleByIdentifier(s, raw)
	if err != nil {
		return 0, translateDomainError(err)
	}
	return project.ID, nil
}
