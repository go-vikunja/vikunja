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
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"xorm.io/xorm"
)

// ProjectDuplicateService handles the duplication of projects and all their related data.
// This service orchestrates the complex process of copying a project, including tasks,
// attachments, labels, assignees, comments, relations, views, and permissions.
type ProjectDuplicateService struct {
	DB             *xorm.Engine
	ProjectService *ProjectService
	TaskService    *TaskService
}

// NewProjectDuplicateService creates a new ProjectDuplicateService.
func NewProjectDuplicateService(db *xorm.Engine) *ProjectDuplicateService {
	return &ProjectDuplicateService{
		DB:             db,
		ProjectService: NewProjectService(db),
		TaskService:    NewTaskService(db),
	}
}

// InitProjectDuplicateService sets up dependency injection for project duplication related model functions.
// This function must be called during test initialization to ensure models can call services.
func InitProjectDuplicateService() {
	// For now, no dependency injection needed for project duplication.
	// This may be expanded later if needed.
}

// Duplicate creates a complete copy of a project and all its related data.
// This includes tasks, attachments, labels, assignees, comments, relations, views,
// kanban data, user/team permissions, and link shares.
//
// The user needs read access to the source project and write access to the parent
// project where the new project will be created.
func (pds *ProjectDuplicateService) Duplicate(s *xorm.Session, projectID int64, parentProjectID int64, u *user.User) (*models.Project, error) {
	// Permission checks: Read access to source project
	canRead, err := pds.ProjectService.HasPermission(s, projectID, u, models.PermissionRead)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, ErrAccessDenied
	}

	// Permission checks: Write access to parent project (if specified)
	if parentProjectID != 0 {
		canCreate, err := pds.ProjectService.HasPermission(s, parentProjectID, u, models.PermissionWrite)
		if err != nil {
			return nil, err
		}
		if !canCreate {
			return nil, ErrAccessDenied
		}
	}

	// Get the source project
	sourceProject, err := models.GetProjectSimpleByID(s, projectID)
	if err != nil {
		return nil, err
	}

	log.Debugf("Duplicating project %d", projectID)

	// Create the new project
	newProject := &models.Project{
		Title:           sourceProject.Title + " - duplicate",
		Description:     sourceProject.Description,
		ParentProjectID: parentProjectID,
		OwnerID:         u.ID,
		Position:        sourceProject.Position,
		HexColor:        sourceProject.HexColor,
		IsFavorite:      false, // Reset favorite status for new project
		IsArchived:      false, // Reset archived status for new project
	}

	// Create the project using ProjectService
	createdProject, err := pds.ProjectService.Create(s, newProject, u)
	if err != nil {
		// If there is no available unique project identifier, reset it and try again
		if models.IsErrProjectIdentifierIsNotUnique(err) {
			newProject.Identifier = ""
			createdProject, err = pds.ProjectService.Create(s, newProject, u)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	log.Debugf("Duplicated project %d into new project %d", projectID, createdProject.ID)

	// Duplicate views first so they exist when tasks are created
	// Create empty task ID map for now
	taskIDMap := make(map[int64]int64)
	err = pds.duplicateProjectViews(s, projectID, createdProject.ID, u, taskIDMap)
	if err != nil {
		return nil, err
	}

	log.Debugf("Duplicated all views from project %d into %d", projectID, createdProject.ID)

	// Now duplicate all tasks and their related data
	taskIDMap, err = pds.duplicateTasksAndRelatedData(s, projectID, createdProject.ID, u)
	if err != nil {
		return nil, err
	}

	log.Debugf("Duplicated all tasks from project %d into %d", projectID, createdProject.ID)

	log.Debugf("Duplicated all views, buckets and positions from project %d into %d", projectID, createdProject.ID)

	// Duplicate project metadata (background, permissions, shares)
	err = pds.duplicateProjectMetadata(s, projectID, createdProject.ID, u)
	if err != nil {
		return nil, err
	}

	log.Debugf("Duplicated all metadata from project %d into %d", projectID, createdProject.ID)

	// Reload the project with full details
	err = createdProject.ReadOne(s, u)
	if err != nil {
		return nil, err
	}

	return createdProject, nil
}

// duplicateTasksAndRelatedData handles the duplication of all tasks and their related data.
// Returns a map of old task ID -> new task ID for use in other duplication functions.
func (pds *ProjectDuplicateService) duplicateTasksAndRelatedData(s *xorm.Session, sourceProjectID int64, targetProjectID int64, u *user.User) (map[int64]int64, error) {
	// Get all tasks from the source project using TaskService
	// Use a large page size to get all tasks at once
	tasks, _, _, err := pds.TaskService.GetAllByProject(s, sourceProjectID, u, 1, 999999, "")
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return make(map[int64]int64), nil
	}

	// This map contains the old task id as key and the new duplicated task id as value.
	// It is used to map old task items to new ones.
	newTaskIDs := make(map[int64]int64, len(tasks))
	oldTaskIDs := make([]int64, 0, len(tasks))

	// Create all tasks using TaskService.Create (proper inter-service communication)
	for _, t := range tasks {
		oldID := t.ID
		t.ID = 0
		t.ProjectID = targetProjectID
		t.UID = "" // Reset UID to generate a new one

		// Clear assignees and bucket data - they will be duplicated separately later
		t.Assignees = nil
		t.BucketID = 0 // Reset bucket ID to use default bucket

		// Use TaskService.CreateWithoutPermissionCheck since we've already verified permissions
		// at the beginning of the duplication process
		createdTask, err := pds.TaskService.CreateWithoutPermissionCheck(s, t, u)
		if err != nil {
			return nil, err
		}

		newTaskIDs[oldID] = createdTask.ID
		oldTaskIDs = append(oldTaskIDs, oldID)
	}

	log.Debugf("Duplicated all tasks from project %d into %d", sourceProjectID, targetProjectID)

	// Duplicate task attachments
	err = pds.duplicateTaskAttachments(s, oldTaskIDs, newTaskIDs, sourceProjectID, targetProjectID, u)
	if err != nil {
		return nil, err
	}

	// Duplicate task labels
	err = pds.duplicateTaskLabels(s, oldTaskIDs, newTaskIDs)
	if err != nil {
		return nil, err
	}

	// Duplicate task assignees
	err = pds.duplicateTaskAssignees(s, oldTaskIDs, newTaskIDs, targetProjectID, u)
	if err != nil {
		return nil, err
	}

	// Duplicate task comments
	err = pds.duplicateTaskComments(s, oldTaskIDs, newTaskIDs)
	if err != nil {
		return nil, err
	}

	// Duplicate task relations
	err = pds.duplicateTaskRelations(s, oldTaskIDs, newTaskIDs)
	if err != nil {
		return nil, err
	}

	return newTaskIDs, nil
}

// duplicateTaskAttachments handles the duplication of task attachments and their underlying files.
func (pds *ProjectDuplicateService) duplicateTaskAttachments(s *xorm.Session, oldTaskIDs []int64, taskIDMap map[int64]int64, sourceProjectID int64, targetProjectID int64, u *user.User) error {
	// Get all attachments for the old tasks by direct query for now
	// TODO: Use TaskService method when attachment handling is properly refactored
	attachments := []*models.TaskAttachment{}
	err := s.In("task_id", oldTaskIDs).Find(&attachments)
	if err != nil {
		return err
	}

	for _, attachment := range attachments {
		oldAttachmentID := attachment.ID
		attachment.ID = 0
		var exists bool
		attachment.TaskID, exists = taskIDMap[attachment.TaskID]
		if !exists {
			log.Debugf("Error duplicating attachment %d from old task %d to new task: Old task <-> new task does not seem to exist.", oldAttachmentID, attachment.TaskID)
			continue
		}

		// Load the file metadata and content
		attachment.File = &files.File{ID: attachment.FileID}
		if err := attachment.File.LoadFileMetaByID(); err != nil {
			if files.IsErrFileDoesNotExist(err) {
				log.Debugf("Not duplicating attachment %d (file %d) because it does not exist from project %d into %d", oldAttachmentID, attachment.FileID, sourceProjectID, targetProjectID)
				continue
			}
			return err
		}
		if err := attachment.File.LoadFileByID(); err != nil {
			return err
		}

		// Create new attachment with duplicated file
		err := attachment.NewAttachment(s, attachment.File.File, attachment.File.Name, attachment.File.Size, u)
		if err != nil {
			return err
		}

		// Close the file handle
		if attachment.File.File != nil {
			_ = attachment.File.File.Close()
		}

		log.Debugf("Duplicated attachment %d into %d from project %d into %d", oldAttachmentID, attachment.ID, sourceProjectID, targetProjectID)
	}

	log.Debugf("Duplicated all attachments from project %d into %d", sourceProjectID, targetProjectID)
	return nil
}

// duplicateTaskLabels handles the duplication of task-label associations.
func (pds *ProjectDuplicateService) duplicateTaskLabels(s *xorm.Session, oldTaskIDs []int64, taskIDMap map[int64]int64) error {
	// Copy label tasks (not the labels themselves, just the associations)
	labelTasks := []*models.LabelTask{}
	err := s.In("task_id", oldTaskIDs).Find(&labelTasks)
	if err != nil {
		return err
	}

	for _, lt := range labelTasks {
		lt.ID = 0
		lt.TaskID = taskIDMap[lt.TaskID]
		if _, err := s.Insert(lt); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated all labels")
	return nil
}

// duplicateTaskAssignees handles the duplication of task assignees with permission checks.
func (pds *ProjectDuplicateService) duplicateTaskAssignees(s *xorm.Session, oldTaskIDs []int64, taskIDMap map[int64]int64, targetProjectID int64, u *user.User) error {
	// Only copy those assignees who have access to the target project
	assignees := []*models.TaskAssginee{}
	err := s.In("task_id", oldTaskIDs).Find(&assignees)
	if err != nil {
		return err
	}

	for _, a := range assignees {
		newTaskID := taskIDMap[a.TaskID]

		// Check if the user being assigned has access to the target project
		assigneeUser := &user.User{ID: a.UserID}
		canRead, err := pds.ProjectService.HasPermission(s, targetProjectID, assigneeUser, models.PermissionRead)
		if err != nil {
			return err
		}
		if !canRead {
			// Skip assignees who don't have access to the target project
			log.Debugf("Skipping assignee %d for task %d because they don't have access to project %d", a.UserID, newTaskID, targetProjectID)
			continue
		}

		// Create new assignee record
		newAssignee := &models.TaskAssginee{
			TaskID: newTaskID,
			UserID: a.UserID,
		}

		if _, err := s.Insert(newAssignee); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated all assignees")
	return nil
}

// duplicateTaskComments handles the duplication of task comments.
func (pds *ProjectDuplicateService) duplicateTaskComments(s *xorm.Session, oldTaskIDs []int64, taskIDMap map[int64]int64) error {
	comments := []*models.TaskComment{}
	err := s.In("task_id", oldTaskIDs).Find(&comments)
	if err != nil {
		return err
	}

	for _, c := range comments {
		c.ID = 0
		c.TaskID = taskIDMap[c.TaskID]
		if _, err := s.Insert(c); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated all comments")
	return nil
}

// duplicateTaskRelations handles the duplication of task relationships, ensuring both tasks exist in the new project.
func (pds *ProjectDuplicateService) duplicateTaskRelations(s *xorm.Session, oldTaskIDs []int64, taskIDMap map[int64]int64) error {
	// Relations in that project
	// Only copy those relations which are between tasks in the same project
	// because we can do that without a lot of hassle
	relations := []*models.TaskRelation{}
	err := s.In("task_id", oldTaskIDs).Find(&relations)
	if err != nil {
		return err
	}

	for _, r := range relations {
		otherTaskID, exists := taskIDMap[r.OtherTaskID]
		if !exists {
			// Skip relations to tasks that don't exist in the target project
			continue
		}
		r.ID = 0
		r.OtherTaskID = otherTaskID
		r.TaskID = taskIDMap[r.TaskID]
		if _, err := s.Insert(r); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated all task relations")
	return nil
}

// duplicateProjectViews handles the duplication of project views and kanban data.
func (pds *ProjectDuplicateService) duplicateProjectViews(s *xorm.Session, sourceProjectID int64, targetProjectID int64, u *user.User, taskIDMap map[int64]int64) error {
	// Duplicate Views
	views := make(map[int64]*models.ProjectView)
	err := s.Where("project_id = ?", sourceProjectID).Find(&views)
	if err != nil {
		return err
	}

	oldViewIDs := []int64{}
	viewMap := make(map[int64]int64)
	for _, view := range views {
		oldID := view.ID
		oldViewIDs = append(oldViewIDs, oldID)

		view.ID = 0
		view.ProjectID = targetProjectID
		err = view.Create(s, u)
		if err != nil {
			return err
		}

		viewMap[oldID] = view.ID
	}

	// Duplicate buckets
	buckets := []*models.Bucket{}
	err = s.In("project_view_id", oldViewIDs).Find(&buckets)
	if err != nil {
		return err
	}

	// Old bucket ID as key, new id as value
	// Used to map the newly created tasks to their new buckets
	bucketMap := make(map[int64]int64)
	oldBucketIDs := []int64{}

	for _, b := range buckets {
		oldBucketID := b.ID
		oldViewID := b.ProjectViewID
		oldBucketIDs = append(oldBucketIDs, oldBucketID)

		b.ID = 0
		b.ProjectID = targetProjectID
		b.ProjectViewID = viewMap[oldViewID]

		err = b.Create(s, u)
		if err != nil {
			return err
		}

		bucketMap[oldBucketID] = b.ID
	}

	// Update view bucket references
	for _, view := range views {
		updated := false
		if view.DefaultBucketID != 0 {
			view.DefaultBucketID = bucketMap[view.DefaultBucketID]
			updated = true
		}
		if view.DoneBucketID != 0 {
			view.DoneBucketID = bucketMap[view.DoneBucketID]
			updated = true
		}

		if updated {
			err = view.Update(s, u)
			if err != nil {
				return err
			}
		}
	}

	// Duplicate task-bucket associations
	oldTaskBuckets := []*models.TaskBucket{}
	err = s.In("bucket_id", oldBucketIDs).Find(&oldTaskBuckets)
	if err != nil {
		return err
	}

	taskBuckets := []*models.TaskBucket{}
	for _, tb := range oldTaskBuckets {
		newTaskID, taskExists := taskIDMap[tb.TaskID]
		if !taskExists {
			continue // Skip if the task wasn't duplicated
		}

		taskBuckets = append(taskBuckets, &models.TaskBucket{
			BucketID:      bucketMap[tb.BucketID],
			TaskID:        newTaskID,
			ProjectViewID: viewMap[tb.ProjectViewID],
		})
	}

	if len(taskBuckets) > 0 {
		_, err = s.Insert(&taskBuckets)
		if err != nil {
			return err
		}
	}

	// Duplicate task positions
	oldTaskPositions := []*models.TaskPosition{}
	err = s.In("project_view_id", oldViewIDs).Find(&oldTaskPositions)
	if err != nil {
		return err
	}

	taskPositions := []*models.TaskPosition{}
	for _, tp := range oldTaskPositions {
		newTaskID, taskExists := taskIDMap[tp.TaskID]
		if !taskExists {
			continue // Skip if the task wasn't duplicated
		}

		taskPositions = append(taskPositions, &models.TaskPosition{
			ProjectViewID: viewMap[tp.ProjectViewID],
			TaskID:        newTaskID,
			Position:      tp.Position,
		})
	}

	if len(taskPositions) > 0 {
		_, err = s.Insert(&taskPositions)
		if err != nil {
			return err
		}
	}

	return nil
}

// duplicateProjectMetadata handles the duplication of project background, permissions, and shares.
func (pds *ProjectDuplicateService) duplicateProjectMetadata(s *xorm.Session, sourceProjectID int64, targetProjectID int64, u *user.User) error {
	// Get the target project to check if it has a background
	targetProject, err := models.GetProjectSimpleByID(s, targetProjectID)
	if err != nil {
		return err
	}

	// Duplicate project background if it exists
	err = pds.duplicateProjectBackground(s, sourceProjectID, targetProject, u)
	if err != nil {
		return err
	}

	// Duplicate user permissions
	err = pds.duplicateUserPermissions(s, sourceProjectID, targetProjectID)
	if err != nil {
		return err
	}

	// Duplicate team permissions
	err = pds.duplicateTeamPermissions(s, sourceProjectID, targetProjectID)
	if err != nil {
		return err
	}

	// Duplicate link shares
	err = pds.duplicateLinkShares(s, sourceProjectID, targetProjectID)
	if err != nil {
		return err
	}

	return nil
}

// duplicateProjectBackground handles the duplication of project background images.
func (pds *ProjectDuplicateService) duplicateProjectBackground(s *xorm.Session, sourceProjectID int64, targetProject *models.Project, u *user.User) error {
	if targetProject.BackgroundFileID == 0 {
		return nil
	}

	log.Debugf("Duplicating background %d from project %d into %d", targetProject.BackgroundFileID, sourceProjectID, targetProject.ID)

	f := &files.File{ID: targetProject.BackgroundFileID}
	err := f.LoadFileMetaByID()
	if err != nil && files.IsErrFileDoesNotExist(err) {
		targetProject.BackgroundFileID = 0
		return nil
	}
	if err != nil {
		return err
	}
	if err := f.LoadFileByID(); err != nil {
		return err
	}
	defer f.File.Close()

	file, err := files.Create(f.File, f.Name, f.Size, u)
	if err != nil {
		return err
	}

	// Get unsplash info if applicable
	up, err := models.GetUnsplashPhotoByFileID(s, targetProject.BackgroundFileID)
	if err != nil && !files.IsErrFileIsNotUnsplashFile(err) {
		return err
	}
	if up != nil {
		up.ID = 0
		up.FileID = file.ID
		if err := up.Save(s); err != nil {
			return err
		}
	}

	if err := models.SetProjectBackground(s, targetProject.ID, file, targetProject.BackgroundBlurHash); err != nil {
		return err
	}

	log.Debugf("Duplicated project background from project %d into %d", sourceProjectID, targetProject.ID)
	return nil
}

// duplicateUserPermissions handles the duplication of user permissions.
func (pds *ProjectDuplicateService) duplicateUserPermissions(s *xorm.Session, sourceProjectID int64, targetProjectID int64) error {
	// To keep it simple(r) we will only copy permissions which are directly used with the project, not the parent
	users := []*models.ProjectUser{}
	err := s.Where("project_id = ?", sourceProjectID).Find(&users)
	if err != nil {
		return err
	}

	for _, u := range users {
		u.ID = 0
		u.ProjectID = targetProjectID
		if _, err := s.Insert(u); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated user shares from project %d into %d", sourceProjectID, targetProjectID)
	return nil
}

// duplicateTeamPermissions handles the duplication of team permissions.
func (pds *ProjectDuplicateService) duplicateTeamPermissions(s *xorm.Session, sourceProjectID int64, targetProjectID int64) error {
	teams := []*models.TeamProject{}
	err := s.Where("project_id = ?", sourceProjectID).Find(&teams)
	if err != nil {
		return err
	}

	for _, t := range teams {
		t.ID = 0
		t.ProjectID = targetProjectID
		if _, err := s.Insert(t); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated team shares from project %d into %d", sourceProjectID, targetProjectID)
	return nil
}

// duplicateLinkShares handles the duplication of link shares with new hashes.
func (pds *ProjectDuplicateService) duplicateLinkShares(s *xorm.Session, sourceProjectID int64, targetProjectID int64) error {
	// Generate new link shares if any are available
	linkShares := []*models.LinkSharing{}
	err := s.Where("project_id = ?", sourceProjectID).Find(&linkShares)
	if err != nil {
		return err
	}

	for _, share := range linkShares {
		share.ID = 0
		share.ProjectID = targetProjectID
		hash, err := utils.CryptoRandomString(40)
		if err != nil {
			return err
		}
		share.Hash = hash
		if _, err := s.Insert(share); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated all link shares from project %d into %d", sourceProjectID, targetProjectID)
	return nil
}
