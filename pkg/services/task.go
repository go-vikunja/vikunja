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

package services

import (
	"fmt"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// TaskService represents a service for managing tasks.
type TaskService struct {
	DB              *xorm.Engine
	FavoriteService *FavoriteService
}

// NewTaskService creates a new TaskService.
func NewTaskService(db *xorm.Engine) *TaskService {
	return &TaskService{
		DB:              db,
		FavoriteService: NewFavoriteService(db),
	}
}

// Wire models.AddMoreInfoToTasksFunc to the service implementation via dependency inversion
// InitTaskService sets up dependency injection for task-related model functions.
// This function must be called during test initialization to ensure models can call services.
func InitTaskService() {
	models.AddMoreInfoToTasksFunc = func(s *xorm.Session, taskMap map[int64]*models.Task, a web.Auth, view *models.ProjectView, expand []models.TaskCollectionExpandable) error {
		return NewTaskService(nil).AddDetailsToTasks(s, taskMap, a, view, expand)
	}

	models.GetUsersOrLinkSharesFromIDsFunc = func(s *xorm.Session, ids []int64) (map[int64]*user.User, error) {
		return NewTaskService(nil).getUsersOrLinkSharesFromIDs(s, ids)
	}
}

// GetByID gets a single task by its ID, checking permissions.
func (ts *TaskService) GetByID(s *xorm.Session, taskID int64, u *user.User) (*models.Task, error) {
	// Use a simple model function to get the raw data
	task := new(models.Task)
	has, err := s.ID(taskID).Get(task)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, models.ErrTaskDoesNotExist{ID: taskID}
	}

	// Permission Check: The TaskService asks the ProjectService for a decision.
	projectService := NewProjectService(ts.DB)
	can, err := projectService.HasPermission(s, task.ProjectID, u, models.PermissionRead)
	if err != nil {
		return nil, fmt.Errorf("checking project read permission: %w", err)
	}
	if !can {
		return nil, ErrAccessDenied
	}

	// Add details to the task
	taskMap := map[int64]*models.Task{task.ID: task}
	err = ts.AddDetailsToTasks(s, taskMap, u, nil, nil)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetAllByProject gets all tasks for a project with pagination and filtering
func (ts *TaskService) GetAllByProject(s *xorm.Session, projectID int64, u *user.User, page int, perPage int, search string) ([]*models.Task, int, int64, error) {
	// Permission Check: Use ProjectService for proper inter-service communication
	projectService := NewProjectService(ts.DB)
	canRead, err := projectService.HasPermission(s, projectID, u, models.PermissionRead)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrAccessDenied
	}

	// Calculate offset for pagination
	offset := (page - 1) * perPage

	// Query tasks directly from the database
	var tasks []*models.Task
	query := s.Where("project_id = ?", projectID)

	// Add search filter if provided
	if search != "" {
		query = query.And(builder.Or(
			builder.Like{"title", "%" + search + "%"},
			builder.Like{"description", "%" + search + "%"},
		))
	}

	// Get total count for pagination
	totalCount, err := query.Count(&models.Task{})
	if err != nil {
		return nil, 0, 0, err
	}

	// Get the actual tasks with pagination
	err = query.
		OrderBy("id ASC").
		Limit(perPage, offset).
		Find(&tasks)
	if err != nil {
		return nil, 0, 0, err
	}

	return tasks, len(tasks), totalCount, nil
}

// Update updates a task.
func (ts *TaskService) Update(s *xorm.Session, task *models.Task, u *user.User) (*models.Task, error) {
	can, err := ts.Can(s, task, u).Write()
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrAccessDenied
	}

	// The old logic used task.Update which did a lot of things.
	// We need to replicate that logic here.
	// For now, we'll just do a simple update.
	if _, err := s.ID(task.ID).AllCols().Update(task); err != nil {
		return nil, err
	}
	return task, nil
}

// Delete deletes a task.
func (ts *TaskService) Delete(s *xorm.Session, task *models.Task, a web.Auth) error {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	can, err := ts.canWriteTask(s, task.ID, u)
	if err != nil {
		return err
	}
	if !can {
		return ErrAccessDenied
	}

	t, err := models.GetTaskByIDSimple(s, task.ID)
	if err != nil {
		return err
	}

	// duplicate the task for the event
	fullTask := &models.Task{ID: task.ID}
	err = fullTask.ReadOne(s, a)
	if err != nil {
		return err
	}

	// Delete assignees
	if _, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskAssginee{}); err != nil {
		return err
	}

	// Delete Favorites using the service
	err = ts.FavoriteService.RemoveFromFavorite(s, task.ID, a, models.FavoriteKindTask)
	if err != nil {
		return err
	}

	// Delete label associations
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.LabelTask{})
	if err != nil {
		return err
	}

	// Delete task attachments
	attachments, err := ts.getTaskAttachmentsByTaskIDs(s, []int64{task.ID})
	if err != nil {
		return err
	}
	for _, attachment := range attachments {
		// Using the attachment delete method here because that takes care of removing all files properly
		err = attachment.Delete(s, a)
		if err != nil && !models.IsErrTaskAttachmentDoesNotExist(err) {
			return err
		}
	}

	// Delete all comments
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskComment{})
	if err != nil {
		return err
	}

	// Delete all relations
	_, err = s.Where("task_id = ? OR other_task_id = ?", task.ID, task.ID).Delete(&models.TaskRelation{})
	if err != nil {
		return err
	}

	// Delete all reminders
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskReminder{})
	if err != nil {
		return err
	}

	// Delete all positions
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskPosition{})
	if err != nil {
		return err
	}

	// Delete all bucket relations
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskBucket{})
	if err != nil {
		return err
	}

	// Actually delete the task
	_, err = s.ID(task.ID).Delete(&models.Task{})
	if err != nil {
		return err
	}

	doer, _ := user.GetFromAuth(a)
	err = events.Dispatch(&models.TaskDeletedEvent{
		Task: fullTask,
		Doer: doer,
	})
	if err != nil {
		return err
	}

	err = ts.updateProjectLastUpdated(s, t.ProjectID)
	return err
}

// TaskPermissions represents the permissions for a task.
type TaskPermissions struct {
	s    *xorm.Session
	task *models.Task
	user *user.User
	ts   *TaskService
}

// Can returns a new TaskPermissions struct.
func (ts *TaskService) Can(s *xorm.Session, task *models.Task, u *user.User) *TaskPermissions {
	return &TaskPermissions{s: s, task: task, user: u, ts: ts}
}

// Read checks if the user can read the task.
// This implements the "Move Logic, Don't Expose It" principle by moving permission logic from models to services.
func (tp *TaskPermissions) Read() (bool, error) {
	if tp.user == nil {
		return false, nil
	}

	// Use ProjectService for permission checking instead of calling model methods
	projectService := NewProjectService(tp.ts.DB)
	return projectService.HasPermission(tp.s, tp.task.ProjectID, tp.user, models.PermissionRead)
}

// Write checks if the user can write to the task.
// This implements the "Move Logic, Don't Expose It" principle by moving permission logic from models to services.
func (tp *TaskPermissions) Write() (bool, error) {
	if tp.user == nil {
		return false, nil
	}

	// Use ProjectService for permission checking instead of calling model methods
	projectService := NewProjectService(tp.ts.DB)
	return projectService.HasPermission(tp.s, tp.task.ProjectID, tp.user, models.PermissionWrite)
}

func (ts *TaskService) addDetailsToTasks(s *xorm.Session, tasks []*models.Task, u *user.User) error {
	if len(tasks) == 0 {
		return nil
	}

	taskMap := make(map[int64]*models.Task, len(tasks))
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	return ts.AddDetailsToTasks(s, taskMap, u, nil, nil)
}

// AddDetailsToTasks adds more info to tasks, like assignees, labels, etc.
// This is the service layer implementation of what was previously models.AddMoreInfoToTasks.
func (ts *TaskService) AddDetailsToTasks(s *xorm.Session, taskMap map[int64]*models.Task, a web.Auth, view *models.ProjectView, expand []models.TaskCollectionExpandable) error {
	if len(taskMap) == 0 {
		return nil
	}

	// Get all users & task ids and put them into the array
	var userIDs []int64
	var taskIDs []int64
	var projectIDs []int64
	for _, task := range taskMap {
		taskIDs = append(taskIDs, task.ID)
		if task.CreatedByID != 0 {
			userIDs = append(userIDs, task.CreatedByID)
		}
		projectIDs = append(projectIDs, task.ProjectID)
	}

	// Add assignees
	err := ts.addAssigneesToTasks(s, taskIDs, taskMap)
	if err != nil {
		return err
	}

	// Add labels
	err = ts.addLabelsToTasks(s, taskIDs, taskMap)
	if err != nil {
		return err
	}

	// Get users for CreatedBy field
	users, err := ts.getUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return err
	}

	// Get task reminders
	taskReminders, err := ts.getTaskReminderMap(s, taskIDs)
	if err != nil {
		return err
	}

	// Get favorites if auth is provided
	var taskFavorites map[int64]bool
	if a != nil {
		taskFavorites, err = ts.getFavorites(s, taskIDs, a, models.FavoriteKindTask)
		if err != nil {
			return err
		}
	}

	// Get all projects for identifiers
	projects, err := models.GetProjectsMapByIDs(s, projectIDs)
	if err != nil {
		return err
	}

	// Add all objects to their tasks
	for _, task := range taskMap {
		// Make created by user objects
		if createdBy, has := users[task.CreatedByID]; has {
			task.CreatedBy = createdBy
		}

		// Add the reminders
		task.Reminders = taskReminders[task.ID]

		// Prepare the subtasks
		task.RelatedTasks = make(models.RelatedTaskMap)

		// Build the task identifier from the project identifier and task index
		if project, exists := projects[task.ProjectID]; exists {
			if project.Identifier == "" {
				task.Identifier = "#" + strconv.FormatInt(task.Index, 10)
			} else {
				task.Identifier = project.Identifier + "-" + strconv.FormatInt(task.Index, 10)
			}
		}

		// Set favorite status
		if taskFavorites != nil {
			task.IsFavorite = taskFavorites[task.ID]
		}
	}

	return nil
}

// Helper methods moved from models package

func (ts *TaskService) addAssigneesToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	taskAssignees := []*models.TaskAssigneeWithUser{}
	err := s.Table("task_assignees").
		Select("task_id, users.*").
		In("task_id", taskIDs).
		Join("INNER", "users", "task_assignees.user_id = users.id").
		Find(&taskAssignees)
	if err != nil {
		return err
	}

	// Put the assignees in the task map
	for i, a := range taskAssignees {
		if a != nil {
			a.Email = "" // Obfuscate the email
			taskMap[a.TaskID].Assignees = append(taskMap[a.TaskID].Assignees, &taskAssignees[i].User)
		}
	}

	return nil
}

func (ts *TaskService) addLabelsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	labels, _, _, err := models.GetLabelsByTaskIDs(s, &models.LabelByTaskIDsOptions{
		TaskIDs: taskIDs,
		Page:    -1,
	})
	if err != nil {
		return err
	}
	for i, l := range labels {
		if l != nil {
			taskMap[l.TaskID].Labels = append(taskMap[l.TaskID].Labels, &labels[i].Label)
		}
	}

	return nil
}

func (ts *TaskService) getTaskReminderMap(s *xorm.Session, taskIDs []int64) (map[int64][]*models.TaskReminder, error) {
	reminders := []*models.TaskReminder{}
	err := s.In("task_id", taskIDs).
		OrderBy("reminder asc").
		Find(&reminders)
	if err != nil {
		return nil, err
	}

	reminderMap := make(map[int64][]*models.TaskReminder)
	for _, reminder := range reminders {
		reminderMap[reminder.TaskID] = append(reminderMap[reminder.TaskID], reminder)
	}

	return reminderMap, nil
}

func (ts *TaskService) getFavorites(s *xorm.Session, entityIDs []int64, a web.Auth, kind models.FavoriteKind) (map[int64]bool, error) {
	favorites := make(map[int64]bool)
	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return favorites, nil
	}

	favs := []*models.Favorite{}
	err = s.Where(builder.And(
		builder.Eq{"user_id": u.ID},
		builder.Eq{"kind": kind},
		builder.In("entity_id", entityIDs),
	)).
		Find(&favs)

	for _, fav := range favs {
		favorites[fav.EntityID] = true
	}
	return favorites, err
}

func (ts *TaskService) canWriteTask(s *xorm.Session, taskID int64, u *user.User) (bool, error) {
	project, err := models.GetProjectSimpleByTaskID(s, taskID)
	if err != nil {
		if models.IsErrProjectDoesNotExist(err) {
			return false, nil
		}
		return false, err
	}

	// Check project permissions using ProjectService
	projectService := NewProjectService(ts.DB)
	return projectService.HasPermission(s, project.ID, u, models.PermissionWrite)
}

// getTaskAttachmentsByTaskIDs gets task attachments with full details
func (ts *TaskService) getTaskAttachmentsByTaskIDs(s *xorm.Session, taskIDs []int64) (attachments []*models.TaskAttachment, err error) {
	attachments = []*models.TaskAttachment{}
	err = s.
		In("task_id", taskIDs).
		Find(&attachments)
	if err != nil {
		return
	}

	if len(attachments) == 0 {
		return
	}

	fileIDs := []int64{}
	userIDs := []int64{}
	for _, a := range attachments {
		userIDs = append(userIDs, a.CreatedByID)
		fileIDs = append(fileIDs, a.FileID)
	}

	// Get all files
	fs := make(map[int64]*files.File)
	err = s.In("id", fileIDs).Find(&fs)
	if err != nil {
		return
	}

	users, err := ts.getUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return nil, err
	}

	// Obfuscate all user emails
	for _, u := range users {
		u.Email = ""
	}

	for _, a := range attachments {
		if createdBy, has := users[a.CreatedByID]; has {
			a.CreatedBy = createdBy
		}
		a.File = fs[a.FileID]
	}

	return
}

// updateProjectLastUpdated updates the last updated timestamp of a project
func (ts *TaskService) updateProjectLastUpdated(s *xorm.Session, projectID int64) error {
	project := &models.Project{
		ID:      projectID,
		Updated: time.Now(),
	}
	_, err := s.ID(projectID).Cols("updated").Update(project)
	return err
}

// getUsersOrLinkSharesFromIDs gets users and link shares from their IDs.
func (ts *TaskService) getUsersOrLinkSharesFromIDs(s *xorm.Session, ids []int64) (users map[int64]*user.User, err error) {
	users = make(map[int64]*user.User)
	var userIDs []int64
	var linkShareIDs []int64
	for _, id := range ids {
		if id < 0 {
			linkShareIDs = append(linkShareIDs, id*-1)
			continue
		}

		userIDs = append(userIDs, id)
	}

	if len(userIDs) > 0 {
		users, err = user.GetUsersByIDs(s, userIDs)
		if err != nil {
			return
		}
	}

	if len(linkShareIDs) == 0 {
		return
	}

	shares, err := models.GetLinkSharesByIDs(s, linkShareIDs)
	if err != nil {
		return nil, err
	}

	for _, share := range shares {
		users[share.ID*-1] = ts.toUser(share)
	}

	return
}

func (ts *TaskService) toUser(share *models.LinkSharing) *user.User {
	suffix := "Link Share"
	if share.Name != "" {
		suffix = " (" + suffix + ")"
	}

	username := "link-share-" + strconv.FormatInt(share.ID, 10)

	return &user.User{
		ID:       ts.getUserID(share),
		Name:     share.Name + suffix,
		Username: username,
		Created:  share.Created,
		Updated:  share.Updated,
	}
}

func (ts *TaskService) getUserID(share *models.LinkSharing) int64 {
	return share.ID * -1
}
