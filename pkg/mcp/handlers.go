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

package mcp

import (
	"encoding/json"
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/mark3labs/mcp-go/mcp"
)

type MCPServer struct {
	authToken string
}

func NewMCPServer(authToken string) *MCPServer {
	return &MCPServer{authToken: authToken}
}

func (s *MCPServer) authenticate(token string) (*user.User, error) {
	if s.authToken != "" && token != s.authToken {
		return nil, fmt.Errorf("invalid token")
	}

	session := db.NewSession()
	defer session.Close()

	apiToken, err := models.GetTokenFromTokenString(session, token)
	if err != nil {
		if models.IsErrAPITokenInvalid(err) {
			return nil, fmt.Errorf("invalid token")
		}
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if time.Now().After(apiToken.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	u, err := user.GetUserByID(session, apiToken.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return u, nil
}

func (s *MCPServer) HandleGetTasks(params json.RawMessage, authToken string) (*mcp.CallToolResult, error) {
	u, err := s.authenticate(authToken)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var filter struct {
		ProjectID *int64  `json:"project_id"`
		ListID    *int64  `json:"list_id"`
		IsDone    *bool   `json:"is_done"`
		Limit     *int    `json:"limit"`
		Offset    *int    `json:"offset"`
		Search    *string `json:"search"`
	}
	if err := json.Unmarshal(params, &filter); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid params: %v", err)), nil
	}

	session := db.NewSession()
	defer session.Close()

	taskCollection := &models.TaskCollection{
		ProjectID: 0,
	}

	if filter.ProjectID != nil {
		taskCollection.ProjectID = *filter.ProjectID
	}
	if filter.Search != nil {
		taskCollection.Search = *filter.Search
	}

	perPage := 50
	page := 1
	if filter.Limit != nil && *filter.Limit > 0 {
		perPage = *filter.Limit
	}
	if filter.Offset != nil && *filter.Offset > 0 {
		page = *filter.Offset/perPage + 1
	}

	result, _, _, err := taskCollection.ReadAll(session, u, taskCollection.Search, page, perPage)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get tasks: %v", err)), nil
	}

	tasks, ok := result.([]*models.Task)
	if !ok {
		return mcp.NewToolResultText(toJSON(result)), nil
	}

	if filter.IsDone != nil {
		var filtered []*models.Task
		for _, t := range tasks {
			if t.Done == *filter.IsDone {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	return mcp.NewToolResultText(toJSON(tasks)), nil
}

func (s *MCPServer) HandleGetTask(params json.RawMessage, authToken string) (*mcp.CallToolResult, error) {
	u, err := s.authenticate(authToken)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var input struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid params: %v", err)), nil
	}

	session := db.NewSession()
	defer session.Close()

	task, err := models.GetTaskByIDSimple(session, input.ID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get task: %v", err)), nil
	}

	canRead, _, err := task.CanRead(session, u)
	if err != nil || !canRead {
		return mcp.NewToolResultError("permission denied"), nil
	}

	return mcp.NewToolResultText(toJSON(task)), nil
}

func (s *MCPServer) HandleCreateTask(params json.RawMessage, authToken string) (*mcp.CallToolResult, error) {
	u, err := s.authenticate(authToken)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var input struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		ProjectID   int64   `json:"project_id"`
		DueDate     *string `json:"due_date"`
		Priority    *int64  `json:"priority"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid params: %v", err)), nil
	}

	if input.Title == "" || input.ProjectID == 0 {
		return mcp.NewToolResultError("title and project_id are required"), nil
	}

	task := &models.Task{
		Title:       input.Title,
		Description: input.Description,
		ProjectID:   input.ProjectID,
		Priority:    0,
		Done:        false,
	}

	if input.DueDate != nil {
		t, err := time.Parse(time.RFC3339, *input.DueDate)
		if err == nil {
			task.DueDate = t
		}
	}
	if input.Priority != nil {
		task.Priority = *input.Priority
	}

	session := db.NewSession()
	defer session.Close()

	err = task.Create(session, u)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create task: %v", err)), nil
	}

	err = session.Commit()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to commit: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(task)), nil
}

func (s *MCPServer) HandleUpdateTask(params json.RawMessage, authToken string) (*mcp.CallToolResult, error) {
	u, err := s.authenticate(authToken)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var input struct {
		ID          int64   `json:"id"`
		Title       string  `json:"title"`
		Description *string `json:"description"`
		Done        *bool   `json:"done"`
		DueDate     *string `json:"due_date"`
		Priority    *int64  `json:"priority"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid params: %v", err)), nil
	}

	session := db.NewSession()
	defer session.Close()

	task, err := models.GetTaskByIDSimple(session, input.ID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get task: %v", err)), nil
	}

	canWrite, err := task.CanWrite(session, u)
	if err != nil || !canWrite {
		return mcp.NewToolResultError("permission denied"), nil
	}

	if input.Title != "" {
		task.Title = input.Title
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.Done != nil {
		task.Done = *input.Done
	}
	if input.DueDate != nil {
		t, err := time.Parse(time.RFC3339, *input.DueDate)
		if err == nil {
			task.DueDate = t
		}
	}
	if input.Priority != nil {
		task.Priority = *input.Priority
	}

	err = task.Update(session, u)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update task: %v", err)), nil
	}

	err = session.Commit()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to commit: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(task)), nil
}

func (s *MCPServer) HandleDeleteTask(params json.RawMessage, authToken string) (*mcp.CallToolResult, error) {
	u, err := s.authenticate(authToken)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var input struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(params, &input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid params: %v", err)), nil
	}

	session := db.NewSession()
	defer session.Close()

	task, err := models.GetTaskByIDSimple(session, input.ID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get task: %v", err)), nil
	}

	canDelete, err := task.CanDelete(session, u)
	if err != nil || !canDelete {
		return mcp.NewToolResultError("permission denied"), nil
	}

	err = task.Delete(session, u)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete task: %v", err)), nil
	}

	err = session.Commit()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to commit: %v", err)), nil
	}

	return mcp.NewToolResultText(`{"success": true, "message": "Task deleted successfully"}`), nil
}

func (s *MCPServer) HandleGetProjects(params json.RawMessage, authToken string) (*mcp.CallToolResult, error) {
	u, err := s.authenticate(authToken)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var filter struct {
		Limit  *int `json:"limit"`
		Offset *int `json:"offset"`
	}
	if err := json.Unmarshal(params, &filter); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid params: %v", err)), nil
	}

	perPage := 50
	page := 1
	if filter.Limit != nil && *filter.Limit > 0 {
		perPage = *filter.Limit
	}
	if filter.Offset != nil && *filter.Offset > 0 {
		page = *filter.Offset/perPage + 1
	}

	session := db.NewSession()
	defer session.Close()

	project := &models.Project{}
	result, _, _, err := project.ReadAll(session, u, "", page, perPage)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get projects: %v", err)), nil
	}

	projects, ok := result.([]*models.Project)
	if !ok {
		return mcp.NewToolResultText(toJSON(result)), nil
	}

	return mcp.NewToolResultText(toJSON(projects)), nil
}

func (s *MCPServer) HandleGetLists(params json.RawMessage, authToken string) (*mcp.CallToolResult, error) {
	u, err := s.authenticate(authToken)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var filter struct {
		ProjectID int64 `json:"project_id"`
		Limit     *int  `json:"limit"`
		Offset    *int  `json:"offset"`
	}
	if err := json.Unmarshal(params, &filter); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid params: %v", err)), nil
	}

	if filter.ProjectID == 0 {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	perPage := 50
	page := 1
	if filter.Limit != nil && *filter.Limit > 0 {
		perPage = *filter.Limit
	}
	if filter.Offset != nil && *filter.Offset > 0 {
		page = *filter.Offset/perPage + 1
	}

	session := db.NewSession()
	defer session.Close()

	project, err := models.GetProjectSimpleByID(session, filter.ProjectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get project: %v", err)), nil
	}

	canRead, _, err := project.CanRead(session, u)
	if err != nil || !canRead {
		return mcp.NewToolResultError("permission denied"), nil
	}

	projectView := &models.ProjectView{ProjectID: filter.ProjectID}
	result, _, _, err := projectView.ReadAll(session, u, "", page, perPage)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get lists: %v", err)), nil
	}

	return mcp.NewToolResultText(toJSON(result)), nil
}

func (s *MCPServer) HandleGetKanbanBoard(params json.RawMessage, authToken string) (*mcp.CallToolResult, error) {
	u, err := s.authenticate(authToken)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var filter struct {
		ProjectID int64 `json:"project_id"`
	}
	if err := json.Unmarshal(params, &filter); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid params: %v", err)), nil
	}

	if filter.ProjectID == 0 {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	session := db.NewSession()
	defer session.Close()

	project, err := models.GetProjectSimpleByID(session, filter.ProjectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get project: %v", err)), nil
	}

	canRead, _, err := project.CanRead(session, u)
	if err != nil || !canRead {
		return mcp.NewToolResultError("permission denied"), nil
	}

	projectView := &models.ProjectView{ProjectID: filter.ProjectID}
	viewsResult, _, _, err := projectView.ReadAll(session, u, "", 1, 50)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get views: %v", err)), nil
	}

	result := struct {
		ProjectID int64                 `json:"project_id"`
		Title     string                `json:"title"`
		Views     []*models.ProjectView `json:"views"`
	}{
		ProjectID: filter.ProjectID,
		Title:     project.Title,
		Views:     viewsResult.([]*models.ProjectView),
	}

	return mcp.NewToolResultText(toJSON(result)), nil
}

func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.Errorf("Error marshaling JSON: %v", err)
		return "{}"
	}
	return string(b)
}
