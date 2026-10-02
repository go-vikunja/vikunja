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
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/danielgtaylor/huma/v2"
)

// assignTaskRequest is what an integration sends to create a task for somebody.
type assignTaskRequest struct {
	AssigneeEmail string    `json:"assignee_email" minLength:"3" maxLength:"250" doc:"The email address of the user the task is assigned to. Compared case-insensitively. Must belong to exactly one active user."`
	Title         string    `json:"title" minLength:"1" maxLength:"250" doc:"The task title."`
	Description   string    `json:"description,omitempty" doc:"The task description. HTML by default, or Markdown with ?format=markdown."`
	DueDate       time.Time `json:"due_date,omitempty" doc:"When the task is due (RFC 3339). Omit for no due date."`
	Priority      int64     `json:"priority,omitempty" minimum:"0" maximum:"5" doc:"The task priority, 0 (none) to 5."`
	ProjectID     int64     `json:"project_id,omitempty" minimum:"0" doc:"The project to create the task in. The token's bot needs write access to it, and the assignee needs to be able to read it. Omit to put the task into the assignee's inbox."`
	ExternalID    string    `json:"external_id,omitempty" maxLength:"250" doc:"Your own reference for this task, for example a ticket number. Sending the same external_id again from the same bot returns the task that already exists instead of creating a second one (200 instead of 201)."`
	AssignorEmail *string   `json:"assignor_email,omitempty" maxLength:"250" nullable:"true" doc:"The email address of the user who assigns the task (shown as the assignor, stored as the task creator). Omit or send null to default to the owner of project_id, or to the assignee when no project is given. Must belong to exactly one active user; the assignor needs no access to the project."`
}

type assignTaskTask struct {
	ID         int64      `json:"id" doc:"The task id."`
	ProjectID  int64      `json:"project_id" doc:"The id of the project the task is in. For an inbox task this is the assignee's inbox."`
	Title      string     `json:"title" doc:"The task title."`
	Identifier string     `json:"identifier" doc:"The textual task identifier, for example \"PROJ-12\"."`
	DueDate    *time.Time `json:"due_date,omitempty" doc:"When the task is due."`
	Priority   int64      `json:"priority" doc:"The task priority."`
	URL        string     `json:"url,omitempty" doc:"A link to the task in the web app. Empty if service.publicurl is not configured."`
}

type assignTaskAssignee struct {
	ID       int64  `json:"id" doc:"The user id."`
	Username string `json:"username" doc:"The username."`
	Name     string `json:"name" doc:"The display name."`
}

type assignTaskResponse struct {
	Created  bool               `json:"created" doc:"True if a new task was created, false if the external_id was seen before and the existing task is returned."`
	Task     assignTaskTask     `json:"task" doc:"The task."`
	Assignee assignTaskAssignee `json:"assignee" doc:"The user the task is assigned to. The email address is not returned."`
	Assignor assignTaskAssignee `json:"assignor" doc:"The user who assigned the task (the task creator). The email address is not returned."`
}

type assignTaskOutput struct {
	Status int
	Body   assignTaskResponse
}

func init() { AddRouteRegistrar(RegisterTaskAssignRoutes) }

// RegisterTaskAssignRoutes registers the endpoint integrations use to create a task for somebody.
func RegisterTaskAssignRoutes(api huma.API) {
	Register(api, huma.Operation{
		OperationID: "tasks-assign",
		Summary:     "Create a task and assign it by email",
		Description: "For integrations. Creates a task and assigns it to the user with the given email address in one call. Without project_id the task goes into the assignee's inbox, which the caller cannot read, edit or delete afterwards. " +
			"Only bot users may call this, with an API token that has the tasks.assign permission; a normal user's token gets 403. " +
			"The assignee does not need to have signed in before, but must be an active user and the only one with that address (404, 409). " +
			"The assignor (stored as the task creator) is assignor_email, else the owner of project_id, else the assignee. " +
			"Send an external_id to make retries safe: the same external_id from the same bot returns the existing task with 200. " +
			"The assignee is notified like for any assignment, with the bot as the doer, so give the bot a meaningful name.",
		Method: http.MethodPost,
		Path:   "/tasks/assign",
		Tags:   []string{"tasks"},
	}, tasksAssign)
}

func tasksAssign(ctx context.Context, in *struct {
	Format string `query:"format" enum:"html,markdown" doc:"How the description is exchanged. See the API description."`
	Body   assignTaskRequest
}) (*assignTaskOutput, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	bot, err := user.GetFromAuth(a)
	if err != nil {
		return nil, translateDomainError(models.ErrGenericForbidden{})
	}

	description := in.Body.Description
	err = convertToHTML(ctx, &description)
	if err != nil {
		return nil, translateDomainError(err)
	}

	input := &models.AssignTaskByEmailInput{
		AssigneeEmail: in.Body.AssigneeEmail,
		Title:         in.Body.Title,
		Description:   description,
		DueDate:       in.Body.DueDate,
		Priority:      in.Body.Priority,
		ProjectID:     in.Body.ProjectID,
		ExternalID:    in.Body.ExternalID,
	}
	if in.Body.AssignorEmail != nil {
		input.AssignorEmail = *in.Body.AssignorEmail
	}

	result, err := runAssignTask(ctx, bot, input)
	if err != nil {
		return nil, translateDomainError(err)
	}

	out := &assignTaskOutput{Status: http.StatusCreated}
	if !result.Created {
		out.Status = http.StatusOK
	}
	out.Body = assignTaskResponse{
		Created: result.Created,
		Task: assignTaskTask{
			ID:         result.Task.ID,
			ProjectID:  result.Task.ProjectID,
			Title:      result.Task.Title,
			Identifier: result.Task.Identifier,
			Priority:   result.Task.Priority,
			URL:        taskURL(result.Task.ID),
		},
		Assignee: assignTaskAssignee{
			ID:       result.Assignee.ID,
			Username: result.Assignee.Username,
			Name:     result.Assignee.Name,
		},
	}
	if result.Assignor != nil {
		out.Body.Assignor = assignTaskAssignee{
			ID:       result.Assignor.ID,
			Username: result.Assignor.Username,
			Name:     result.Assignor.Name,
		}
	}
	if !result.Task.DueDate.IsZero() {
		due := result.Task.DueDate
		out.Body.Task.DueDate = &due
	}
	return out, nil
}

// runAssignTask runs the model in its own transaction. When an external id is used and the call
// fails with something that is not a domain error, it is tried once more: two concurrent calls with
// the same external id race on the unique key, and the loser then finds the winner's task.
func runAssignTask(ctx context.Context, bot *user.User, input *models.AssignTaskByEmailInput) (*models.AssignTaskByEmailResult, error) {
	result, err := assignTaskOnce(ctx, bot, input)
	if err == nil || input.ExternalID == "" {
		return result, err
	}
	var domain web.HTTPErrorProcessor
	if errors.As(err, &domain) {
		return nil, err
	}
	return assignTaskOnce(ctx, bot, input)
}

func assignTaskOnce(ctx context.Context, bot *user.User, input *models.AssignTaskByEmailInput) (*models.AssignTaskByEmailResult, error) {
	s := db.NewSession()
	defer s.Close()

	result, err := models.AssignTaskByEmail(s, bot, input)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}
	err = s.Commit()
	if err != nil {
		events.CleanupPending(s)
		return nil, err
	}
	events.DispatchPending(ctx, s)
	return result, nil
}

func taskURL(id int64) string {
	base := strings.TrimRight(config.ServicePublicURL.GetString(), "/")
	if base == "" {
		return ""
	}
	return base + "/tasks/" + strconv.FormatInt(id, 10)
}
