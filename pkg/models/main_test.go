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

package models

import (
	"fmt"
	"os"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// mockProjectDuplicateService provides a test implementation for project duplication
// This prevents import cycles while allowing model tests to continue working
type mockProjectDuplicateService struct{}

func (m *mockProjectDuplicateService) Duplicate(s *xorm.Session, projectID int64, parentProjectID int64, u *user.User) (*Project, error) {
	// Simple mock implementation for tests - delegates to old model logic
	// This is what the old ProjectDuplicate.Create did before refactoring
	pd := &ProjectDuplicate{
		ProjectID:       projectID,
		ParentProjectID: parentProjectID,
	}

	// Get the source project
	sourceProject, err := GetProjectSimpleByID(s, projectID)
	if err != nil {
		return nil, err
	}

	// Create new project with duplicate suffix
	pd.Project = &Project{
		Title:           sourceProject.Title + " - duplicate",
		Description:     sourceProject.Description,
		ParentProjectID: parentProjectID,
		OwnerID:         u.ID,
		Position:        sourceProject.Position,
		HexColor:        sourceProject.HexColor,
	}

	// Use getProjectService() to create the project (same as before)
	projectService := getProjectService()
	createdProject, err := projectService.Create(s, pd.Project, u)
	if err != nil {
		// If there is no available unique project identifier, reset it and try again
		if IsErrProjectIdentifierIsNotUnique(err) {
			pd.Project.Identifier = ""
			createdProject, err = projectService.Create(s, pd.Project, u)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return createdProject, nil
}

func (m *mockProjectDuplicateService) CanCreate(s *xorm.Session, projectID int64, parentProjectID int64, a web.Auth) (bool, error) {
	// Simple mock implementation for tests - always return true for tests
	return true, nil
}

// mockProjectViewService provides a test implementation for project views
// This prevents import cycles while allowing model tests to continue working
// Following T-CLEANUP pattern - this mock will be removed in future cleanup tasks
type mockProjectViewService struct{}

func (m *mockProjectViewService) Create(s *xorm.Session, pv *ProjectView, a web.Auth, createBacklogBucket bool, addExistingTasksToView bool) error {
	// Simplified version - just validate and insert the view
	if pv.Filter != nil && pv.Filter.Filter != "" {
		_, err := GetTaskFiltersFromFilterString(pv.Filter.Filter, pv.Filter.FilterTimezone)
		if err != nil {
			return err
		}
	}
	pv.ID = 0
	_, err := s.Insert(pv)
	return err
}

func (m *mockProjectViewService) Update(s *xorm.Session, pv *ProjectView) error {
	if pv.Filter != nil && pv.Filter.Filter != "" {
		_, err := GetTaskFiltersFromFilterString(pv.Filter.Filter, pv.Filter.FilterTimezone)
		if err != nil {
			return err
		}
	}
	_, err := s.ID(pv.ID).Cols("title", "view_kind", "filter", "position", "bucket_configuration_mode", "bucket_configuration", "default_bucket_id", "done_bucket_id").Update(pv)
	return err
}

func (m *mockProjectViewService) Delete(s *xorm.Session, viewID int64, projectID int64) error {
	_, err := s.Where("id = ? AND project_id = ?", viewID, projectID).Delete(&ProjectView{})
	return err
}

func (m *mockProjectViewService) GetAll(s *xorm.Session, projectID int64, a web.Auth) (views []*ProjectView, totalCount int64, err error) {
	views = []*ProjectView{}
	err = s.Where("project_id = ?", projectID).OrderBy("position asc").Find(&views)
	if err != nil {
		return nil, 0, err
	}
	totalCount, err = s.Where("project_id = ?", projectID).Count(&ProjectView{})
	return views, totalCount, err
}

func (m *mockProjectViewService) GetByIDAndProject(s *xorm.Session, viewID, projectID int64) (view *ProjectView, err error) {
	if projectID == FavoritesPseudoProjectID && viewID < 0 {
		for _, v := range FavoritesPseudoProject.Views {
			if v.ID == viewID {
				return v, nil
			}
		}
		return nil, &ErrProjectViewDoesNotExist{ProjectViewID: viewID}
	}
	view = &ProjectView{}
	exists, err := s.Where("id = ? AND project_id = ?", viewID, projectID).NoAutoCondition().Get(view)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ErrProjectViewDoesNotExist{ProjectViewID: viewID}
	}
	return view, nil
}

func (m *mockProjectViewService) GetByID(s *xorm.Session, id int64) (view *ProjectView, err error) {
	view = &ProjectView{}
	exists, err := s.Where("id = ?", id).NoAutoCondition().Get(view)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ErrProjectViewDoesNotExist{ProjectViewID: id}
	}
	return view, nil
}

func (m *mockProjectViewService) CreateDefaultViewsForProject(s *xorm.Session, project *Project, a web.Auth, createBacklogBucket bool, createDefaultListFilter bool) error {
	// Create the four default views
	list := &ProjectView{ProjectID: project.ID, Title: "List", ViewKind: ProjectViewKindList, Position: 100}
	if createDefaultListFilter {
		list.Filter = &TaskCollection{FilterTimezone: "GMT"}
	}

	gantt := &ProjectView{ProjectID: project.ID, Title: "Gantt", ViewKind: ProjectViewKindGantt, Position: 200}
	table := &ProjectView{ProjectID: project.ID, Title: "Table", ViewKind: ProjectViewKindTable, Position: 300}
	kanban := &ProjectView{ProjectID: project.ID, Title: "Kanban", ViewKind: ProjectViewKindKanban, Position: 400, BucketConfigurationMode: BucketConfigurationModeManual}

	for _, view := range []*ProjectView{list, gantt, table, kanban} {
		_, err := s.Insert(view)
		if err != nil {
			return err
		}

		// Create default buckets for Kanban view
		if view.ViewKind == ProjectViewKindKanban && createBacklogBucket {
			buckets := []*Bucket{
				{Title: "To-Do", ProjectViewID: view.ID, Position: 0},
				{Title: "Doing", ProjectViewID: view.ID, Position: 1},
				{Title: "Done", ProjectViewID: view.ID, Position: 2},
			}
			for _, bucket := range buckets {
				_, err := s.Insert(bucket)
				if err != nil {
					return err
				}
			}

			// Set default and done buckets
			view.DefaultBucketID = buckets[0].ID
			view.DoneBucketID = buckets[2].ID
			_, err = s.ID(view.ID).Cols("default_bucket_id", "done_bucket_id").Update(view)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// mockTaskService provides a test implementation for task operations
// This prevents import cycles while allowing model tests to continue working
type mockTaskService struct{}

func (m *mockTaskService) Create(s *xorm.Session, task *Task, u *user.User, updateAssignees bool, setBucket bool) error {
	// Minimal implementation for model tests
	// For proper task creation tests, use service layer tests
	task.CreatedByID = u.ID
	_, err := s.Insert(task)
	return err
}

func (m *mockTaskService) Update(s *xorm.Session, task *Task, u *user.User) (*Task, error) {
	// Basic update for model tests
	// For proper task update tests, use service layer tests
	cols := []string{"title", "description", "done", "due_date", "priority", "repeat_after", "start_date", "end_date", "hex_color", "percent_done", "project_id"}
	_, err := s.ID(task.ID).Cols(cols...).Update(task)
	return task, err
}

func (m *mockTaskService) Delete(s *xorm.Session, task *Task, a web.Auth) error {
	// Basic delete for model tests
	// For proper task delete tests, use service layer tests
	_, err := s.ID(task.ID).Delete(&Task{})
	return err
}

func (m *mockTaskService) GetByID(s *xorm.Session, taskID int64, u *user.User) (*Task, error) {
	// Simple implementation - just fetch the task without expansion
	task := &Task{ID: taskID}
	exists, err := s.Get(task)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTaskDoesNotExist{ID: taskID}
	}
	return task, nil
}

func (m *mockTaskService) GetByIDSimple(s *xorm.Session, taskID int64) (*Task, error) {
	// Simple implementation - just fetch the task without permission checks
	task := &Task{ID: taskID}
	exists, err := s.Get(task)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTaskDoesNotExist{ID: taskID}
	}
	return task, nil
}

func (m *mockTaskService) GetByIDs(s *xorm.Session, ids []int64) ([]*Task, error) {
	tasks := []*Task{}
	err := s.In("id", ids).Find(&tasks)
	return tasks, err
}

// Permission methods (T-PERM-010) - mock implementations for testing
func (m *mockTaskService) CanCreateAssignee(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	task := &Task{ID: taskID}
	return task.CanWrite(s, a)
}

func (m *mockTaskService) CanDeleteAssignee(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	task := &Task{ID: taskID}
	return task.CanWrite(s, a)
}

func (m *mockTaskService) CanCreateRelation(s *xorm.Session, taskID int64, otherTaskID int64, relationKind RelationKind, a web.Auth) (bool, error) {
	task := &Task{ID: taskID}
	canWrite, err := task.CanWrite(s, a)
	if err != nil || !canWrite {
		return canWrite, err
	}
	otherTask := &Task{ID: otherTaskID}
	return otherTask.CanWrite(s, a)
}

func (m *mockTaskService) CanDeleteRelation(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	task := &Task{ID: taskID}
	return task.CanWrite(s, a)
}

func (m *mockTaskService) CanUpdatePosition(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	task := &Task{ID: taskID}
	return task.CanWrite(s, a)
}

// mockBulkTaskService provides a test implementation for bulk task operations
// This prevents import cycles while allowing model tests to continue working
type mockBulkTaskService struct{}

func (m *mockBulkTaskService) GetTasksByIDs(s *xorm.Session, taskIDs []int64) ([]*Task, error) {
	// Validate all IDs are positive
	for _, id := range taskIDs {
		if id < 1 {
			return nil, ErrTaskDoesNotExist{ID: id}
		}
	}

	// Fetch tasks
	tasks := []*Task{}
	err := s.In("id", taskIDs).Find(&tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (m *mockBulkTaskService) CanUpdate(s *xorm.Session, taskIDs []int64, a web.Auth) (bool, error) {
	// Get the tasks
	tasks, err := m.GetTasksByIDs(s, taskIDs)
	if err != nil {
		return false, err
	}

	if len(tasks) == 0 {
		return false, ErrBulkTasksNeedAtLeastOne{}
	}

	// Check if all tasks are in the same project
	firstProjectID := tasks[0].ProjectID
	for _, t := range tasks {
		if t.ProjectID != firstProjectID {
			return false, ErrBulkTasksMustBeInSameProject{
				ShouldBeID: firstProjectID,
				IsID:       t.ProjectID,
			}
		}
	}

	// Check if user has write access to the project
	project := &Project{ID: tasks[0].ProjectID}
	return project.CanWrite(s, a)
}

func (m *mockBulkTaskService) Update(s *xorm.Session, taskIDs []int64, taskUpdate *Task, assignees []*user.User, a web.Auth) error {
	// Get the tasks
	tasks, err := m.GetTasksByIDs(s, taskIDs)
	if err != nil {
		return err
	}

	// NOTE: No validation here - CanUpdate should be called first by the handler
	// The original model's Update method doesn't validate same-project constraint

	// Update each task
	for _, oldTask := range tasks {
		// When a repeating task is marked as done, we update all deadlines and reminders and set it as undone
		UpdateDone(oldTask, taskUpdate)

		// Update the assignees
		if err := oldTask.UpdateTaskAssignees(s, assignees, a); err != nil {
			return err
		}

		// Merge the update into the old task using copier as a simple merge alternative
		if taskUpdate.Title != "" {
			oldTask.Title = taskUpdate.Title
		}
		if taskUpdate.Description != "" {
			oldTask.Description = taskUpdate.Description
		}
		oldTask.Done = taskUpdate.Done
		if !taskUpdate.DueDate.IsZero() {
			oldTask.DueDate = taskUpdate.DueDate
		}
		if len(taskUpdate.Reminders) > 0 {
			oldTask.Reminders = taskUpdate.Reminders
		}
		if taskUpdate.RepeatAfter != 0 {
			oldTask.RepeatAfter = taskUpdate.RepeatAfter
		}
		if taskUpdate.Priority != 0 {
			oldTask.Priority = taskUpdate.Priority
		}
		if !taskUpdate.StartDate.IsZero() {
			oldTask.StartDate = taskUpdate.StartDate
		}
		if !taskUpdate.EndDate.IsZero() {
			oldTask.EndDate = taskUpdate.EndDate
		}

		// And because a false is considered to be a null value, we need to explicitly check that case here.
		if !taskUpdate.Done {
			oldTask.Done = false
		}

		// Save the updated task
		_, err = s.ID(oldTask.ID).
			Cols("title",
				"description",
				"done",
				"due_date",
				"reminders",
				"repeat_after",
				"priority",
				"start_date",
				"end_date").
			Update(oldTask)
		if err != nil {
			return err
		}
	}

	return nil
}

// mockProjectService provides a test implementation that uses direct logic
// This prevents import cycles while allowing model tests to continue working
// Updated to not call model helper functions per T011A-PART2 requirements
type mockProjectService struct{}

func (m *mockProjectService) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int, isArchived bool, expand ProjectExpandable) (projects []*Project, resultCount int, totalItems int64, err error) {
	// Replicate the core logic without calling model helpers

	// Check if we're dealing with a share auth
	shareAuth, is := a.(*LinkSharing)
	if is {
		project, err := GetProjectSimpleByID(s, shareAuth.ProjectID)
		if err != nil {
			return nil, 0, 0, err
		}
		projects := []*Project{project}

		// Add details manually for share auth
		if AddProjectDetailsFunc != nil {
			err = AddProjectDetailsFunc(s, projects, a)
		}
		if err == nil && len(projects) > 0 {
			projects[0].ParentProjectID = 0
		}
		return projects, 0, 0, err
	}

	doer, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}

	// Get raw projects using the low-level function
	prs, resultCount, totalItems, err := getRawProjectsForUser(
		s,
		&ProjectOptions{
			Search:      search,
			User:        doer,
			Page:        page,
			PerPage:     perPage,
			GetArchived: isArchived,
		})
	if err != nil {
		return nil, 0, 0, err
	}

	// Add saved filters
	savedFiltersProject, err := getSavedFilterProjects(s, doer, search)
	if err != nil {
		return nil, 0, 0, err
	}
	if len(savedFiltersProject) > 0 {
		prs = append(prs, savedFiltersProject...)
	}

	// Add project details using the function variable if set
	if AddProjectDetailsFunc != nil {
		err = AddProjectDetailsFunc(s, prs, a)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	// Handle permission expansion
	if expand == ProjectExpandableRights {
		err = AddMaxPermissionToProjects(s, prs, doer)
		if err != nil {
			return nil, 0, 0, err
		}
	} else {
		for _, pr := range prs {
			pr.MaxPermission = PermissionUnknown
		}
	}

	return prs, resultCount, totalItems, err
}

func (m *mockProjectService) Create(s *xorm.Session, project *Project, u *user.User) (*Project, error) {
	// Validate that the user exists in the database
	// This is crucial because the user might be a test stub or invalid reference
	_, err := user.GetUserByID(s, u.ID)
	if err != nil {
		return nil, err
	}

	// Replicate the core logic without calling model helper CreateProject

	err = project.CheckIsArchived(s)
	if err != nil {
		return nil, err
	}

	project.ID = 0
	project.OwnerID = u.ID
	project.Owner = u

	err = project.validate(s, project)
	if err != nil {
		return nil, err
	}

	project.HexColor = utils.NormalizeHex(project.HexColor)

	_, err = s.Insert(project)
	if err != nil {
		return nil, err
	}

	project.Position = CalculateDefaultPosition(project.ID, project.Position)
	_, err = s.Where("id = ?", project.ID).Update(project)
	if err != nil {
		return nil, err
	}

	if project.IsFavorite {
		if err := AddToFavorites(s, project.ID, u, FavoriteKindProject); err != nil {
			return nil, err
		}
	}

	// Create default views for tests
	err = CreateDefaultViewsForProject(s, project, u, true, true)
	if err != nil {
		return nil, err
	}

	err = events.Dispatch(&ProjectCreatedEvent{
		Project: project,
		Doer:    u,
	})
	if err != nil {
		return nil, err
	}

	// Return full project with details
	fullProject, err := GetProjectSimpleByID(s, project.ID)
	if err != nil {
		return nil, err
	}

	err = fullProject.ReadOne(s, u)
	if err != nil {
		return nil, err
	}

	return fullProject, nil
}

func (m *mockProjectService) Delete(s *xorm.Session, projectID int64, a web.Auth) error {
	// Replicate the core delete logic for tests

	// Load the project
	project, err := GetProjectSimpleByID(s, projectID)
	if err != nil {
		return err
	}

	// Check if this is a default project
	isDefaultProject, err := project.IsDefaultProject(s)
	if err != nil {
		return err
	}

	// No one can delete a default project (not even the owner)
	if isDefaultProject {
		return &ErrCannotDeleteDefaultProject{ProjectID: project.ID}
	}

	// Permission check - simplified for tests
	// Check if auth is a link share
	shareAuth, isShare := a.(*LinkSharing)
	if isShare {
		// Link shares can only delete if they have admin permission and it's their project
		if !(project.ID == shareAuth.ProjectID && shareAuth.Permission == PermissionAdmin) {
			return ErrGenericForbidden{}
		}
	} else {
		// Get user for regular auth
		doerUser, err := GetUserOrLinkShareUser(s, a)
		if err != nil {
			return err
		}

		// Owner has full permissions
		if project.OwnerID != doerUser.ID {
			// For non-owners, check if they have admin rights
			can, err := project.CanWrite(s, a)
			if err != nil {
				return err
			}
			if !can {
				return ErrGenericForbidden{}
			}
		}
	}

	// Delete all tasks on that project
	tasks := []*Task{}
	err = s.Where("project_id = ?", project.ID).Find(&tasks)
	if err != nil {
		return err
	}

	// Get user for task deletion
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		err = task.Delete(s, u)
		if err != nil {
			return err
		}
	}

	// Delete background file if exists
	if project.BackgroundFileID != 0 {
		_, err := s.Where("id = ?", project.BackgroundFileID).Delete(&files.File{})
		if err != nil {
			return err
		}
	}

	// Delete related project entities
	views := []*ProjectView{}
	err = s.Where("project_id = ?", project.ID).Find(&views)
	if err != nil {
		return err
	}

	viewIDs := []int64{}
	for _, v := range views {
		viewIDs = append(viewIDs, v.ID)
	}

	if len(viewIDs) > 0 {
		// Delete buckets associated with these views
		_, err = s.In("project_view_id", viewIDs).Delete(&Bucket{})
		if err != nil {
			return err
		}

		// Delete the views themselves
		_, err = s.In("id", viewIDs).Delete(&ProjectView{})
		if err != nil {
			return err
		}
	}

	// Remove from favorites
	err = RemoveFromFavorite(s, project.ID, u, FavoriteKindProject)
	if err != nil {
		return err
	}

	// Delete link sharing
	_, err = s.Where("project_id = ?", project.ID).Delete(&LinkSharing{})
	if err != nil {
		return err
	}

	// Delete project users
	_, err = s.Where("project_id = ?", project.ID).Delete(&ProjectUser{})
	if err != nil {
		return err
	}

	// Delete team projects
	_, err = s.Where("project_id = ?", project.ID).Delete(&TeamProject{})
	if err != nil {
		return err
	}

	// Delete the project itself
	_, err = s.ID(project.ID).Delete(&Project{})
	if err != nil {
		return err
	}

	// Dispatch project deleted event
	err = events.Dispatch(&ProjectDeletedEvent{
		Project: project,
		Doer:    u,
	})
	if err != nil {
		return err
	}

	// Delete child projects recursively
	childProjects := []*Project{}
	err = s.Where("parent_project_id = ?", project.ID).Find(&childProjects)
	if err != nil {
		return err
	}

	for _, child := range childProjects {
		err = m.Delete(s, child.ID, u)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *mockProjectService) DeleteForce(s *xorm.Session, projectID int64, a web.Auth) error {
	// DeleteForce is the same as Delete but allows deleting default projects
	// Load the project
	project, err := GetProjectSimpleByID(s, projectID)
	if err != nil {
		return err
	}

	// Check if this is a default project
	isDefaultProject, err := project.IsDefaultProject(s)
	if err != nil {
		return err
	}

	// If we're deleting a default project, remove it as default first
	if isDefaultProject {
		_, err = s.Where("default_project_id = ?", project.ID).
			Cols("default_project_id").
			Update(&user.User{DefaultProjectID: 0})
		if err != nil {
			return err
		}
	}

	// Permission check - simplified for tests
	// Check if auth is a link share
	shareAuth, isShare := a.(*LinkSharing)
	if isShare {
		// Link shares can only delete if they have admin permission and it's their project
		if !(project.ID == shareAuth.ProjectID && shareAuth.Permission == PermissionAdmin) {
			return ErrGenericForbidden{}
		}
	} else {
		// Get user for regular auth
		doerUser, err := GetUserOrLinkShareUser(s, a)
		if err != nil {
			return err
		}

		// Owner has full permissions
		if project.OwnerID != doerUser.ID {
			// For non-owners, check if they have admin rights
			can, err := project.CanWrite(s, a)
			if err != nil {
				return err
			}
			if !can {
				return ErrGenericForbidden{}
			}
		}
	}

	// Delete all tasks on that project
	tasks := []*Task{}
	err = s.Where("project_id = ?", project.ID).Find(&tasks)
	if err != nil {
		return err
	}

	// Get user for task deletion
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		err = task.Delete(s, u)
		if err != nil {
			return err
		}
	}

	// Delete background file if exists
	if project.BackgroundFileID != 0 {
		_, err := s.Where("id = ?", project.BackgroundFileID).Delete(&files.File{})
		if err != nil {
			return err
		}
	}

	// Delete related project entities
	views := []*ProjectView{}
	err = s.Where("project_id = ?", project.ID).Find(&views)
	if err != nil {
		return err
	}

	viewIDs := []int64{}
	for _, v := range views {
		viewIDs = append(viewIDs, v.ID)
	}

	if len(viewIDs) > 0 {
		// Delete buckets associated with these views
		_, err = s.In("project_view_id", viewIDs).Delete(&Bucket{})
		if err != nil {
			return err
		}

		// Delete the views themselves
		_, err = s.In("id", viewIDs).Delete(&ProjectView{})
		if err != nil {
			return err
		}
	}

	// Remove from favorites
	err = RemoveFromFavorite(s, project.ID, u, FavoriteKindProject)
	if err != nil {
		return err
	}

	// Delete link sharing
	_, err = s.Where("project_id = ?", project.ID).Delete(&LinkSharing{})
	if err != nil {
		return err
	}

	// Delete project users
	_, err = s.Where("project_id = ?", project.ID).Delete(&ProjectUser{})
	if err != nil {
		return err
	}

	// Delete team projects
	_, err = s.Where("project_id = ?", project.ID).Delete(&TeamProject{})
	if err != nil {
		return err
	}

	// Delete the project itself
	_, err = s.ID(project.ID).Delete(&Project{})
	if err != nil {
		return err
	}

	// Dispatch project deleted event
	err = events.Dispatch(&ProjectDeletedEvent{
		Project: project,
		Doer:    u,
	})
	if err != nil {
		return err
	}

	// Delete child projects recursively
	childProjects := []*Project{}
	err = s.Where("parent_project_id = ?", project.ID).Find(&childProjects)
	if err != nil {
		return err
	}

	for _, child := range childProjects {
		err = m.DeleteForce(s, child.ID, a)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *mockProjectService) GetByIDSimple(s *xorm.Session, projectID int64) (*Project, error) {
	return GetProjectSimpleByID(s, projectID)
}

func (m *mockProjectService) GetByIDs(s *xorm.Session, projectIDs []int64) ([]*Project, error) {
	return GetProjectsByIDs(s, projectIDs)
}

func (m *mockProjectService) GetMapByIDs(s *xorm.Session, projectIDs []int64) (map[int64]*Project, error) {
	return GetProjectsMapByIDs(s, projectIDs)
}

func setupTime() {
	var err error
	loc, err := time.LoadLocation("GMT")
	if err != nil {
		fmt.Printf("Error setting up time: %s", err)
		os.Exit(1)
	}
	testCreatedTime, err = time.ParseInLocation(time.RFC3339Nano, "2018-12-01T15:13:12.0+00:00", loc)
	if err != nil {
		fmt.Printf("Error setting up time: %s", err)
		os.Exit(1)
	}
	testCreatedTime = testCreatedTime.In(loc)
	testUpdatedTime, err = time.ParseInLocation(time.RFC3339Nano, "2018-12-02T15:13:12.0+00:00", loc)
	if err != nil {
		fmt.Printf("Error setting up time: %s", err)
		os.Exit(1)
	}
	testUpdatedTime = testUpdatedTime.In(loc)
}

// mockLabelTaskService provides a test implementation for label-task operations
// This prevents import cycles while allowing model tests to continue working
type mockLabelTaskService struct{}

func (m *mockLabelTaskService) AddLabelToTask(s *xorm.Session, labelID, taskID int64, auth web.Auth) error {
	// Check if the label is already added
	exists, err := s.Exist(&LabelTask{LabelID: labelID, TaskID: taskID})
	if err != nil {
		return err
	}
	if exists {
		return ErrLabelIsAlreadyOnTask{labelID, taskID}
	}

	// Create the label task with ID=0 to let the database auto-increment
	lt := &LabelTask{ID: 0, LabelID: labelID, TaskID: taskID}
	_, err = s.Insert(lt)
	if err != nil {
		return err
	}

	err = triggerTaskUpdatedEventForTaskID(s, auth, taskID)
	if err != nil {
		return err
	}

	return updateProjectByTaskID(s, taskID)
}

func (m *mockLabelTaskService) RemoveLabelFromTask(s *xorm.Session, labelID, taskID int64, auth web.Auth) error {
	_, err := s.Delete(&LabelTask{LabelID: labelID, TaskID: taskID})
	if err != nil {
		return err
	}

	return triggerTaskUpdatedEventForTaskID(s, auth, taskID)
}

func (m *mockLabelTaskService) UpdateTaskLabels(s *xorm.Session, taskID int64, newLabels []*Label, auth web.Auth) error {
	// Get current task with labels
	task, err := GetTaskByIDSimple(s, taskID)
	if err != nil {
		return err
	}

	// Get current labels
	currentLabels, _, _, err := m.GetLabelsByTaskIDs(s, &LabelByTaskIDsOptions{
		TaskIDs: []int64{taskID},
	})
	if err != nil {
		return err
	}

	task.Labels = make([]*Label, 0, len(currentLabels))
	for i := range currentLabels {
		task.Labels = append(task.Labels, &currentLabels[i].Label)
	}

	// If we don't have any new labels, delete everything right away
	if len(newLabels) == 0 && len(task.Labels) > 0 {
		_, err = s.Where("task_id = ?", taskID).Delete(LabelTask{})
		return err
	}

	// If we didn't change anything (from 0 to zero) don't do anything
	if len(newLabels) == 0 && len(task.Labels) == 0 {
		return nil
	}

	// Make a hashmap of the new labels for easier comparison
	newLabelsMap := make(map[int64]*Label, len(newLabels))
	for _, newLabel := range newLabels {
		newLabelsMap[newLabel.ID] = newLabel
	}

	// Get old labels to delete
	var labelsToDelete []int64
	oldLabels := make(map[int64]*Label, len(task.Labels))
	for _, oldLabel := range task.Labels {
		if newLabelsMap[oldLabel.ID] == nil {
			// Label not in new list, mark for deletion
			labelsToDelete = append(labelsToDelete, oldLabel.ID)
		}
		oldLabels[oldLabel.ID] = oldLabel
	}

	// Delete all labels not passed
	if len(labelsToDelete) > 0 {
		_, err = s.In("label_id", labelsToDelete).
			And("task_id = ?", taskID).
			Delete(LabelTask{})
		if err != nil {
			return err
		}
	}

	// Loop through our labels and add them
	for _, l := range newLabels {
		// Check if the label is already added on the task and only add it if not
		if oldLabels[l.ID] != nil {
			continue
		}

		// Add the new label
		// Note: Permission check removed for test mock simplicity
		// In production, LabelTaskService performs proper permission checks

		// Insert it
		_, err = s.Insert(&LabelTask{
			LabelID: l.ID,
			TaskID:  taskID,
		})
		if err != nil {
			return err
		}
	}

	err = triggerTaskUpdatedEventForTaskID(s, auth, taskID)
	if err != nil {
		return err
	}

	err = UpdateProjectLastUpdated(s, &Project{ID: task.ProjectID})
	return err
}

func (m *mockLabelTaskService) GetLabelsByTaskIDs(s *xorm.Session, opts *LabelByTaskIDsOptions) ([]*LabelWithTaskID, int, int64, error) {
	// This is a simplified implementation for tests
	// Check if the user has the permission to see the task (if single task)
	if len(opts.TaskIDs) == 1 && opts.User != nil {
		task := Task{ID: opts.TaskIDs[0]}
		canRead, _, err := task.CanRead(s, opts.User)
		if err != nil {
			return nil, 0, 0, err
		}
		if !canRead {
			return nil, 0, 0, ErrNoPermissionToSeeTask{opts.TaskIDs[0], opts.User.GetID()}
		}
	}

	// Get labels for the task IDs
	var labels []*LabelWithTaskID
	query := s.Table("labels").
		Select("labels.*, label_tasks.task_id").
		Join("LEFT", "label_tasks", "label_tasks.label_id = labels.id").
		In("label_tasks.task_id", opts.TaskIDs).
		OrderBy("labels.id ASC")

	if len(opts.Search) > 0 && opts.Search[0] != "" {
		query = query.Where("labels.title LIKE ?", "%"+opts.Search[0]+"%")
	}

	err := query.Find(&labels)
	if err != nil {
		return nil, 0, 0, err
	}

	// Get all created by users
	var userids []int64
	for _, l := range labels {
		userids = append(userids, l.CreatedByID)
	}
	users := make(map[int64]*user.User)
	if len(userids) > 0 {
		err = s.In("id", userids).Find(&users)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	// Obfuscate all user emails
	for _, u := range users {
		u.Email = ""
	}

	// Put it all together
	for in, l := range labels {
		if createdBy, has := users[l.CreatedByID]; has {
			labels[in].CreatedBy = createdBy
		}
	}

	return labels, len(labels), int64(len(labels)), nil
}

// Inline helper functions for test initialization (used by CRUD function pointers)
// These support deprecated model CRUD methods that delegate to services
func getSavedFilterSimpleByIDForTest(s *xorm.Session, id int64) (*SavedFilter, error) {
	sf := &SavedFilter{}
	exists, err := s.Where("id = ?", id).Get(sf)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ErrSavedFilterDoesNotExist{SavedFilterID: id}
	}
	return sf, nil
}

func getLinkSharesByIDsForTest(s *xorm.Session, ids []int64) (map[int64]*LinkSharing, error) {
	shares := make(map[int64]*LinkSharing)
	if len(ids) == 0 {
		return shares, nil
	}
	var shareList []*LinkSharing
	err := s.In("id", ids).Find(&shareList)
	if err != nil {
		return nil, err
	}
	for _, share := range shareList {
		shares[share.ID] = share
	}
	return shares, nil
}

func getProjectViewByIDAndProjectForTest(s *xorm.Session, viewID, projectID int64) (*ProjectView, error) {
	pv := &ProjectView{}
	exists, err := s.Where("id = ? AND project_id = ?", viewID, projectID).Get(pv)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ErrProjectViewDoesNotExist{ProjectViewID: viewID}
	}
	return pv, nil
}

func getProjectViewByIDForTest(s *xorm.Session, id int64) (*ProjectView, error) {
	pv := &ProjectView{}
	exists, err := s.Where("id = ?", id).Get(pv)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ErrProjectViewDoesNotExist{ProjectViewID: id}
	}
	return pv, nil
}

func TestMain(m *testing.M) {
	// T-PERM-016B: Simplified TestMain for pure structure tests
	// Model tests are now pure structure tests with NO database dependencies
	// All CRUD and permission tests have been moved to service layer

	setupTime()

	// Initialize logger for tests (minimal setup)
	log.InitLogger()

	// Set default config (required by some structure validation)
	config.InitDefaultConfig()

	// Initialize i18n for error messages
	i18n.Init()

	// Note: No DB setup, no fixtures loading, no service mocks needed
	// Structure tests only validate data structures, JSON marshaling, etc.

	os.Exit(m.Run())
}
