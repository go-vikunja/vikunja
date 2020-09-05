// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/web"
)

// ListDuplicate holds everything needed to duplicate a list
type ListDuplicate struct {
	// The list id of the list to duplicate
	ListID int64 `json:"-" param:"listid"`
	// The target namespace ID
	NamespaceID int64 `json:"namespace_id,omitempty"`

	// The copied list
	List *List `json:",omitempty"`

	web.Rights   `json:"-"`
	web.CRUDable `json:"-"`
}

// CanCreate checks if a user has the right to duplicate a list
func (ld *ListDuplicate) CanCreate(a web.Auth) (canCreate bool, err error) {
	// List Exists + user has read access to list
	ld.List = &List{ID: ld.ListID}
	canRead, _, err := ld.List.CanRead(a)
	if err != nil || !canRead {
		return canRead, err
	}

	// Namespace exists + user has write access to is (-> can create new lists)
	ld.List.NamespaceID = ld.NamespaceID
	return ld.List.CanCreate(a)
}

// Create duplicates a list
// @Summary Duplicate an existing list
// @Description Copies the list, tasks, files, kanban data, assignees, comments, attachments, lables, relations, backgrounds, user/team rights and link shares from one list to a new namespace. The user needs read access in the list and write access in the namespace of the new list.
// @tags list
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param listID path int true "The list ID to duplicate"
// @Param list body models.ListDuplicate true "The target namespace which should hold the copied list."
// @Success 200 {object} models.ListDuplicate "The created list."
// @Failure 400 {object} web.HTTPError "Invalid list duplicate object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the list or namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/duplicate [put]
func (ld *ListDuplicate) Create(a web.Auth) (err error) {

	log.Debugf("Duplicating list %d", ld.ListID)

	ld.List.ID = 0
	ld.List.Identifier = "" // Reset the identifier to trigger regenerating a new one
	// Set the owner to the current user
	ld.List.OwnerID = a.GetID()
	if err := CreateOrUpdateList(ld.List); err != nil {
		// If there is no available unique list identifier, just reset it.
		if IsErrListIdentifierIsNotUnique(err) {
			ld.List.Identifier = ""
		} else {
			return err
		}
	}

	log.Debugf("Duplicated list %d into new list %d", ld.ListID, ld.List.ID)

	// Duplicate kanban buckets
	// Old bucket ID as key, new id as value
	// Used to map the newly created tasks to their new buckets
	bucketMap := make(map[int64]int64)
	buckets := []*Bucket{}
	err = x.Where("list_id = ?", ld.ListID).Find(&buckets)
	if err != nil {
		return
	}
	for _, b := range buckets {
		oldID := b.ID
		b.ID = 0
		b.ListID = ld.List.ID
		if err := b.Create(a); err != nil {
			return err
		}
		bucketMap[oldID] = b.ID
	}

	log.Debugf("Duplicated all buckets from list %d into %d", ld.ListID, ld.List.ID)

	// Get all tasks + all task details
	tasks, _, _, err := getTasksForLists([]*List{{ID: ld.ListID}}, a, &taskOptions{})
	if err != nil {
		return err
	}

	// This map contains the old task id as key and the new duplicated task id as value.
	// It is used to map old task items to new ones.
	taskMap := make(map[int64]int64)
	// Create + update all tasks (includes reminders)
	oldTaskIDs := make([]int64, len(tasks))
	for _, t := range tasks {
		oldID := t.ID
		t.ID = 0
		t.ListID = ld.List.ID
		t.BucketID = bucketMap[t.BucketID]
		t.UID = ""
		s := x.NewSession()
		err := createTask(s, t, a, false)
		if err != nil {
			_ = s.Rollback()
			return err
		}
		taskMap[oldID] = t.ID
		oldTaskIDs = append(oldTaskIDs, oldID)
	}

	log.Debugf("Duplicated all tasks from list %d into %d", ld.ListID, ld.List.ID)

	// Save all attachments
	// We also duplicate all underlying files since they could be modified in one list which would result in
	// file changes in the other list which is not something we want.
	attachments, err := getTaskAttachmentsByTaskIDs(oldTaskIDs)
	if err != nil {
		return err
	}

	for _, attachment := range attachments {
		oldAttachmentID := attachment.ID
		attachment.ID = 0
		var exists bool
		attachment.TaskID, exists = taskMap[attachment.TaskID]
		if !exists {
			log.Debugf("Error duplicating attachment %d from old task %d to new task: Old task <-> new task does not seem to exist.", oldAttachmentID, attachment.TaskID)
			continue
		}
		attachment.File = &files.File{ID: attachment.FileID}
		if err := attachment.File.LoadFileMetaByID(); err != nil {
			if files.IsErrFileDoesNotExist(err) {
				log.Debugf("Not duplicating attachment %d (file %d) because it does not exist from list %d into %d", oldAttachmentID, attachment.FileID, ld.ListID, ld.List.ID)
				continue
			}
			return err
		}
		if err := attachment.File.LoadFileByID(); err != nil {
			return err
		}

		err := attachment.NewAttachment(attachment.File.File, attachment.File.Name, attachment.File.Size, a)
		if err != nil {
			return err
		}

		if attachment.File.File != nil {
			_ = attachment.File.File.Close()
		}

		log.Debugf("Duplicated attachment %d into %d from list %d into %d", oldAttachmentID, attachment.ID, ld.ListID, ld.List.ID)
	}

	log.Debugf("Duplicated all attachments from list %d into %d", ld.ListID, ld.List.ID)

	// Copy label tasks (not the labels)
	labelTasks := []*LabelTask{}
	err = x.In("task_id", oldTaskIDs).Find(&labelTasks)
	if err != nil {
		return
	}

	for _, lt := range labelTasks {
		lt.ID = 0
		lt.TaskID = taskMap[lt.TaskID]
		if _, err := x.Insert(lt); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated all labels from list %d into %d", ld.ListID, ld.List.ID)

	// Assignees
	// Only copy those assignees who have access to the task
	assignees := []*TaskAssginee{}
	err = x.In("task_id", oldTaskIDs).Find(&assignees)
	if err != nil {
		return
	}
	for _, a := range assignees {
		t := &Task{
			ID:     taskMap[a.TaskID],
			ListID: ld.List.ID,
		}
		if err := t.addNewAssigneeByID(a.UserID, ld.List); err != nil {
			if IsErrUserDoesNotHaveAccessToList(err) {
				continue
			}
			return err
		}
	}

	log.Debugf("Duplicated all assignees from list %d into %d", ld.ListID, ld.List.ID)

	// Comments
	comments := []*TaskComment{}
	err = x.In("task_id", oldTaskIDs).Find(&comments)
	if err != nil {
		return
	}
	for _, c := range comments {
		c.ID = 0
		c.TaskID = taskMap[c.TaskID]
		if _, err := x.Insert(c); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated all comments from list %d into %d", ld.ListID, ld.List.ID)

	// Relations in that list
	// Low-Effort: Only copy those relations which are between tasks in the same list
	// because we can do that without a lot of hassle
	relations := []*TaskRelation{}
	err = x.In("task_id", oldTaskIDs).Find(&relations)
	if err != nil {
		return
	}
	for _, r := range relations {
		otherTaskID, exists := taskMap[r.OtherTaskID]
		if !exists {
			continue
		}
		r.ID = 0
		r.OtherTaskID = otherTaskID
		r.TaskID = taskMap[r.TaskID]
		if _, err := x.Insert(r); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated all task relations from list %d into %d", ld.ListID, ld.List.ID)

	// Background files + unsplash info
	if ld.List.BackgroundFileID != 0 {

		log.Debugf("Duplicating background %d from list %d into %d", ld.List.BackgroundFileID, ld.ListID, ld.List.ID)

		f := &files.File{ID: ld.List.BackgroundFileID}
		if err := f.LoadFileMetaByID(); err != nil {
			return err
		}
		if err := f.LoadFileByID(); err != nil {
			return err
		}
		defer f.File.Close()

		file, err := files.Create(f.File, f.Name, f.Size, a)
		if err != nil {
			return err
		}

		// Get unsplash info if applicable
		up, err := GetUnsplashPhotoByFileID(ld.List.BackgroundFileID)
		if err != nil && files.IsErrFileIsNotUnsplashFile(err) {
			return err
		}
		if up != nil {
			up.ID = 0
			up.FileID = file.ID
			if err := up.Save(); err != nil {
				return err
			}
		}

		if err := SetListBackground(ld.List.ID, file); err != nil {
			return err
		}

		log.Debugf("Duplicated list background from list %d into %d", ld.ListID, ld.List.ID)
	}

	// Rights / Shares
	// To keep it simple(r) we will only copy rights which are directly used with the list, no namespace changes.
	users := []*ListUser{}
	err = x.Where("list_id = ?", ld.ListID).Find(&users)
	if err != nil {
		return
	}
	for _, u := range users {
		u.ID = 0
		u.ListID = ld.List.ID
		if _, err := x.Insert(u); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated user shares from list %d into %d", ld.ListID, ld.List.ID)

	teams := []*TeamList{}
	err = x.Where("list_id = ?", ld.ListID).Find(&teams)
	if err != nil {
		return
	}
	for _, t := range teams {
		t.ID = 0
		t.ListID = ld.List.ID
		if _, err := x.Insert(t); err != nil {
			return err
		}
	}

	// Generate new link shares if any are available
	linkShares := []*LinkSharing{}
	err = x.Where("list_id = ?", ld.ListID).Find(&linkShares)
	if err != nil {
		return
	}
	for _, share := range linkShares {
		share.ID = 0
		share.ListID = ld.List.ID
		share.Hash = utils.MakeRandomString(40)
		if _, err := x.Insert(share); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated all link shares from list %d into %d", ld.ListID, ld.List.ID)

	return
}
