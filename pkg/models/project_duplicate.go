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
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// ProjectDuplicate holds everything needed to duplicate a project
type ProjectDuplicate struct {
	// The project id of the project to duplicate
	ProjectID int64 `json:"-" param:"projectid"`
	// The target parent project
	ParentProjectID int64 `json:"parent_project_id,omitempty"`

	// The copied project
	Project *Project `json:"duplicated_project,omitempty"`

	web.Permissions `json:"-"`
	web.CRUDable    `json:"-"`
}

// CanCreate checks if a user has the permission to duplicate a project
func (pd *ProjectDuplicate) CanCreate(s *xorm.Session, a web.Auth) (canCreate bool, err error) {
	// Project Exists + user has read access to project
	pd.Project = &Project{ID: pd.ProjectID}
	canRead, _, err := pd.Project.CanRead(s, a)
	if err != nil || !canRead {
		return canRead, err
	}

	if pd.ParentProjectID == 0 { // no parent project
		return canRead, err
	}

	// Parent project exists + user has write access to is (-> can create new projects)
	parent := &Project{ID: pd.ParentProjectID}
	return parent.CanCreate(s, a)
}

// Create duplicates a project
// @Summary Duplicate an existing project
// @Description Copies the project, tasks, files, kanban data, assignees, comments, attachments, labels, relations, backgrounds, user/team permissions and link shares from one project to a new one. The user needs read access in the project and write access in the parent of the new project.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param projectID path int true "The project ID to duplicate"
// @Param project body models.ProjectDuplicate true "The target parent project which should hold the copied project."
// @Success 201 {object} models.ProjectDuplicate "The created project."
// @Failure 400 {object} web.HTTPError "Invalid project duplicate object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project or its parent."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{projectID}/duplicate [put]
//
//nolint:gocyclo
func (pd *ProjectDuplicate) Create(s *xorm.Session, doer web.Auth) (err error) {

	log.Debugf("Duplicating project %d", pd.ProjectID)

	pd.Project.ID = 0
	pd.Project.Identifier = "" // Reset the identifier to trigger regenerating a new one
	pd.Project.ParentProjectID = pd.ParentProjectID
	// Set the owner to the current user
	pd.Project.OwnerID = doer.GetID()
	pd.Project.Title += " - duplicate"
	err = CreateProject(s, pd.Project, doer, false, false)
	if err != nil {
		// If there is no available unique project identifier, just reset it.
		if IsErrProjectIdentifierIsNotUnique(err) {
			pd.Project.Identifier = ""
		} else {
			return err
		}
	}

	log.Debugf("Duplicated project %d into new project %d", pd.ProjectID, pd.Project.ID)

	newTaskIDs, err := duplicateTasks(s, doer, pd)
	if err != nil {
		return
	}

	log.Debugf("Duplicated all tasks from project %d into %d", pd.ProjectID, pd.Project.ID)

	err = duplicateViews(s, pd, doer, newTaskIDs)
	if err != nil {
		return
	}

	log.Debugf("Duplicated all views, buckets and positions from project %d into %d", pd.ProjectID, pd.Project.ID)

	err = duplicateProjectBackground(s, pd, doer)
	if err != nil {
		return
	}

	// Permissions / Shares
	// To keep it simple(r) we will only copy permissions which are directly used with the project, not the parent
	users := []*ProjectUser{}
	err = s.Where("project_id = ?", pd.ProjectID).Find(&users)
	if err != nil {
		return
	}
	for _, u := range users {
		u.ID = 0
		u.ProjectID = pd.Project.ID
		if _, err := s.Insert(u); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated user shares from project %d into %d", pd.ProjectID, pd.Project.ID)

	teams := []*TeamProject{}
	err = s.Where("project_id = ?", pd.ProjectID).Find(&teams)
	if err != nil {
		return
	}
	for _, t := range teams {
		t.ID = 0
		t.ProjectID = pd.Project.ID
		if _, err := s.Insert(t); err != nil {
			return err
		}
	}

	// Generate new link shares if any are available
	linkShares := []*LinkSharing{}
	err = s.Where("project_id = ?", pd.ProjectID).Find(&linkShares)
	if err != nil {
		return
	}
	for _, share := range linkShares {
		share.ID = 0
		share.ProjectID = pd.Project.ID
		hash, err := utils.CryptoRandomString(40)
		if err != nil {
			return err
		}
		share.Hash = hash
		if _, err := s.Insert(share); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated all link shares from project %d into %d", pd.ProjectID, pd.Project.ID)

	err = pd.Project.ReadOne(s, doer)
	return
}

func duplicateViews(s *xorm.Session, pd *ProjectDuplicate, doer web.Auth, taskMap map[int64]int64) (err error) {
	// Duplicate Views
	views := make(map[int64]*ProjectView)
	err = s.Where("project_id = ?", pd.ProjectID).Find(&views)
	if err != nil {
		return
	}

	oldViewIDs := []int64{}
	viewMap := make(map[int64]int64)
	for _, view := range views {
		oldID := view.ID
		oldViewIDs = append(oldViewIDs, oldID)

		view.ID = 0
		view.ProjectID = pd.Project.ID
		err = createProjectView(s, view, doer, false, false)
		if err != nil {
			return
		}

		viewMap[oldID] = view.ID
	}

	buckets := []*Bucket{}
	err = s.In("project_view_id", oldViewIDs).Find(&buckets)
	if err != nil {
		return
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
		b.ProjectID = pd.Project.ID
		b.ProjectViewID = viewMap[oldViewID]

		err = b.Create(s, doer)
		if err != nil {
			return err
		}

		bucketMap[oldBucketID] = b.ID
	}

	for _, view := range views {
		if view.DefaultBucketID != 0 {
			view.DefaultBucketID = bucketMap[view.DefaultBucketID]
		}
		if view.DoneBucketID != 0 {
			view.DoneBucketID = bucketMap[view.DoneBucketID]
		}

		if view.DefaultBucketID != 0 || view.DoneBucketID != 0 {
			err = view.Update(s, doer)
			if err != nil {
				return err
			}
		}
	}

	oldTaskBuckets := []*TaskBucket{}
	err = s.In("bucket_id", oldBucketIDs).Find(&oldTaskBuckets)
	if err != nil {
		return err
	}

	taskBuckets := []*TaskBucket{}
	for _, tb := range oldTaskBuckets {
		taskBuckets = append(taskBuckets, &TaskBucket{
			BucketID:      bucketMap[tb.BucketID],
			TaskID:        taskMap[tb.TaskID],
			ProjectViewID: viewMap[tb.ProjectViewID],
		})
	}

	if len(taskBuckets) > 0 {
		_, err = s.Insert(&taskBuckets)
		if err != nil {
			return err
		}
	}

	oldTaskPositions := []*TaskPosition{}
	err = s.In("project_view_id", oldViewIDs).Find(&oldTaskPositions)
	if err != nil {
		return
	}

	taskPositions := []*TaskPosition{}
	for _, tp := range oldTaskPositions {
		taskPositions = append(taskPositions, &TaskPosition{
			ProjectViewID: viewMap[tp.ProjectViewID],
			TaskID:        taskMap[tp.TaskID],
			Position:      tp.Position,
		})
	}

	if len(taskPositions) > 0 {
		_, err = s.Insert(&taskPositions)
	}
	return
}

func duplicateProjectBackground(s *xorm.Session, pd *ProjectDuplicate, doer web.Auth) (err error) {
	if pd.Project.BackgroundFileID == 0 {
		return
	}

	log.Debugf("Duplicating background %d from project %d into %d", pd.Project.BackgroundFileID, pd.ProjectID, pd.Project.ID)

	f := &files.File{ID: pd.Project.BackgroundFileID}
	err = f.LoadFileMetaByID()
	if err != nil && files.IsErrFileDoesNotExist(err) {
		pd.Project.BackgroundFileID = 0
		return nil
	}
	if err != nil {
		return err
	}
	if err := f.LoadFileByID(); err != nil {
		return err
	}
	defer f.File.Close()

	file, err := files.Create(f.File, f.Name, f.Size, doer)
	if err != nil {
		return err
	}

	// Get unsplash info if applicable
	up, err := GetUnsplashPhotoByFileID(s, pd.Project.BackgroundFileID)
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

	if err := SetProjectBackground(s, pd.Project.ID, file, pd.Project.BackgroundBlurHash); err != nil {
		return err
	}

	log.Debugf("Duplicated project background from project %d into %d", pd.ProjectID, pd.Project.ID)

	return
}

func duplicateTasks(s *xorm.Session, doer web.Auth, ld *ProjectDuplicate) (newTaskIDs map[int64]int64, err error) {
	// Get all tasks + all task details
	tasks, _, _, err := getTasksForProjects(s, []*Project{{ID: ld.ProjectID}}, doer, &taskSearchOptions{}, nil)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return
	}

	// This map contains the old task id as key and the new duplicated task id as value.
	// It is used to map old task items to new ones.
	newTaskIDs = make(map[int64]int64, len(tasks))
	// Create + update all tasks (includes reminders)
	oldTaskIDs := make([]int64, 0, len(tasks))
	for _, t := range tasks {
		oldID := t.ID
		t.ID = 0
		t.ProjectID = ld.Project.ID
		t.UID = ""
		err = createTask(s, t, doer, false, false)
		if err != nil {
			return nil, err
		}
		newTaskIDs[oldID] = t.ID
		oldTaskIDs = append(oldTaskIDs, oldID)
	}

	log.Debugf("Duplicated all tasks from project %d into %d", ld.ProjectID, ld.Project.ID)

	// Save all attachments
	// We also duplicate all underlying files since they could be modified in one project which would result in
	// file changes in the other project which is not something we want.
	attachments, err := getTaskAttachmentsByTaskIDs(s, oldTaskIDs)
	if err != nil {
		return nil, err
	}

	for _, attachment := range attachments {
		oldAttachmentID := attachment.ID
		attachment.ID = 0
		var exists bool
		attachment.TaskID, exists = newTaskIDs[attachment.TaskID]
		if !exists {
			log.Debugf("Error duplicating attachment %d from old task %d to new task: Old task <-> new task does not seem to exist.", oldAttachmentID, attachment.TaskID)
			continue
		}
		attachment.File = &files.File{ID: attachment.FileID}
		if err := attachment.File.LoadFileMetaByID(); err != nil {
			if files.IsErrFileDoesNotExist(err) {
				log.Debugf("Not duplicating attachment %d (file %d) because it does not exist from project %d into %d", oldAttachmentID, attachment.FileID, ld.ProjectID, ld.Project.ID)
				continue
			}
			return nil, err
		}
		if err := attachment.File.LoadFileByID(); err != nil {
			return nil, err
		}

		err := attachment.NewAttachment(s, attachment.File.File, attachment.File.Name, attachment.File.Size, doer)
		if err != nil {
			return nil, err
		}

		if attachment.File.File != nil {
			_ = attachment.File.File.Close()
		}

		log.Debugf("Duplicated attachment %d into %d from project %d into %d", oldAttachmentID, attachment.ID, ld.ProjectID, ld.Project.ID)
	}

	log.Debugf("Duplicated all attachments from project %d into %d", ld.ProjectID, ld.Project.ID)

	// Copy label tasks (not the labels)
	labelTasks := []*LabelTask{}
	err = s.In("task_id", oldTaskIDs).Find(&labelTasks)
	if err != nil {
		return
	}

	for _, lt := range labelTasks {
		lt.ID = 0
		lt.TaskID = newTaskIDs[lt.TaskID]
		if _, err := s.Insert(lt); err != nil {
			return nil, err
		}
	}

	log.Debugf("Duplicated all labels from project %d into %d", ld.ProjectID, ld.Project.ID)

	// Assignees
	// Only copy those assignees who have access to the task
	assignees := []*TaskAssginee{}
	err = s.In("task_id", oldTaskIDs).Find(&assignees)
	if err != nil {
		return
	}
	for _, a := range assignees {
		t := &Task{
			ID:        newTaskIDs[a.TaskID],
			ProjectID: ld.Project.ID,
		}
		if err := t.addNewAssigneeByID(s, a.UserID, ld.Project, doer); err != nil {
			if IsErrUserDoesNotHaveAccessToProject(err) {
				continue
			}
			return nil, err
		}
	}

	log.Debugf("Duplicated all assignees from project %d into %d", ld.ProjectID, ld.Project.ID)

	// Comments
	comments := []*TaskComment{}
	err = s.In("task_id", oldTaskIDs).Find(&comments)
	if err != nil {
		return
	}
	for _, c := range comments {
		c.ID = 0
		c.TaskID = newTaskIDs[c.TaskID]
		if _, err := s.Insert(c); err != nil {
			return nil, err
		}
	}

	log.Debugf("Duplicated all comments from project %d into %d", ld.ProjectID, ld.Project.ID)

	// Relations in that project
	// Low-Effort: Only copy those relations which are between tasks in the same project
	// because we can do that without a lot of hassle
	relations := []*TaskRelation{}
	err = s.In("task_id", oldTaskIDs).Find(&relations)
	if err != nil {
		return
	}
	for _, r := range relations {
		otherTaskID, exists := newTaskIDs[r.OtherTaskID]
		if !exists {
			continue
		}
		r.ID = 0
		r.OtherTaskID = otherTaskID
		r.TaskID = newTaskIDs[r.TaskID]
		if _, err := s.Insert(r); err != nil {
			return nil, err
		}
	}

	log.Debugf("Duplicated all task relations from project %d into %d", ld.ProjectID, ld.Project.ID)

	return
}
