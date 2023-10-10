// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package migration

import (
	"bytes"
	"io"

	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/background/handler"
	"code.vikunja.io/api/pkg/user"
)

// InsertFromStructure takes a fully nested Vikunja data structure and a user and then creates everything for this user
// (Projects, tasks, etc. Even attachments and relations.)
func InsertFromStructure(str []*models.ProjectWithTasksAndBuckets, user *user.User) (err error) {
	s := db.NewSession()
	defer s.Close()

	err = insertFromStructure(s, str, user)
	if err != nil {
		log.Errorf("[creating structure] Error while creating structure: %s", err.Error())
		_ = s.Rollback()
		return err
	}

	return s.Commit()
}

func insertFromStructure(s *xorm.Session, str []*models.ProjectWithTasksAndBuckets, user *user.User) (err error) {

	log.Debugf("[creating structure] Creating %d projects", len(str))

	labels := make(map[string]*models.Label)
	archivedProjects := []int64{}

	childRelations := make(map[int64][]int64)          // old id is the key, slice of old children ids
	projectsByOldID := make(map[int64]*models.Project) // old id is the key
	// Create all projects
	for i, p := range str {
		oldID := p.ID

		if p.ParentProjectID != 0 {
			childRelations[p.ParentProjectID] = append(childRelations[p.ParentProjectID], oldID)
		}

		p.ID = 0
		err = createProject(s, p, &archivedProjects, labels, user)
		if err != nil {
			return err
		}
		projectsByOldID[oldID] = &str[i].Project
	}

	// parent / child relations
	for parentID, children := range childRelations {
		parent, has := projectsByOldID[parentID]
		if !has {
			log.Debugf("[creating structure] could not find parentID project with old id %d", parentID)
			continue
		}
		for _, childID := range children {
			child, has := projectsByOldID[childID]
			if !has {
				log.Debugf("[creating structure] could not find child project with old id %d for parent project with old id %d", childID, parentID)
				continue
			}

			child.ParentProjectID = parent.ID
			err = child.Update(s, user)
			if err != nil {
				return err
			}
		}
	}

	if len(archivedProjects) > 0 {
		_, err = s.
			Cols("is_archived").
			In("id", archivedProjects).
			Update(&models.Project{IsArchived: true})
		if err != nil {
			return err
		}
	}

	log.Debugf("[creating structure] Done inserting new task structure")

	return nil
}

func createProject(s *xorm.Session, project *models.ProjectWithTasksAndBuckets, archivedProjectIDs *[]int64, labels map[string]*models.Label, user *user.User) (err error) {
	err = createProjectWithEverything(s, project, archivedProjectIDs, labels, user)
	if err != nil {
		return err
	}

	log.Debugf("[creating structure] Created project %d", project.ID)

	return
}

func createProjectWithEverything(s *xorm.Session, project *models.ProjectWithTasksAndBuckets, archivedProjects *[]int64, labels map[string]*models.Label, user *user.User) (err error) {
	// The tasks and bucket slices are going to be reset during the creation of the project, so we rescue it here
	// to be able to still loop over them aftere the project was created.
	tasks := project.Tasks
	originalBuckets := project.Buckets
	originalBackgroundInformation := project.BackgroundInformation
	needsDefaultBucket := false

	// Saving the archived status to archive the project again after creating it
	var wasArchived bool
	if project.IsArchived {
		wasArchived = true
		project.IsArchived = false
	}

	project.ID = 0
	err = project.Create(s, user)
	if err != nil && models.IsErrProjectIdentifierIsNotUnique(err) {
		project.Identifier = ""
		err = project.Create(s, user)
	}
	if err != nil {
		return
	}

	if wasArchived {
		*archivedProjects = append(*archivedProjects, project.ID)
	}

	log.Debugf("[creating structure] Created project %d", project.ID)

	bf, is := originalBackgroundInformation.(*bytes.Buffer)
	if is {

		backgroundFile := bytes.NewReader(bf.Bytes())

		log.Debugf("[creating structure] Creating a background file for project %d", project.ID)

		err = handler.SaveBackgroundFile(s, user, &project.Project, backgroundFile, "", uint64(backgroundFile.Len()))
		if err != nil {
			return err
		}

		log.Debugf("[creating structure] Created a background file for project %d", project.ID)
	}

	// Create all buckets
	buckets := make(map[int64]*models.Bucket) // old bucket id is the key
	if len(project.Buckets) > 0 {
		log.Debugf("[creating structure] Creating %d buckets", len(project.Buckets))
	}
	for _, bucket := range originalBuckets {
		oldID := bucket.ID
		bucket.ID = 0 // We want a new id
		bucket.ProjectID = project.ID
		err = bucket.Create(s, user)
		if err != nil {
			return
		}
		buckets[oldID] = bucket
		log.Debugf("[creating structure] Created bucket %d, old ID was %d", bucket.ID, oldID)
	}

	log.Debugf("[creating structure] Creating %d tasks", len(tasks))

	setBucketOrDefault := func(task *models.Task) {
		bucket, exists := buckets[task.BucketID]
		if exists {
			task.BucketID = bucket.ID
		} else if task.BucketID > 0 {
			log.Debugf("[creating structure] No bucket created for original bucket id %d", task.BucketID)
			task.BucketID = 0
		}
		if !exists || task.BucketID == 0 {
			needsDefaultBucket = true
		}
	}

	tasksByOldID := make(map[int64]*models.TaskWithComments, len(tasks))
	// Create all tasks
	for i, t := range tasks {
		setBucketOrDefault(&tasks[i].Task)

		oldid := t.ID
		t.ProjectID = project.ID
		err = t.Create(s, user)
		if err != nil {
			return
		}
		tasksByOldID[oldid] = t

		log.Debugf("[creating structure] Created task %d", t.ID)
		if len(t.RelatedTasks) > 0 {
			log.Debugf("[creating structure] Creating %d related task kinds", len(t.RelatedTasks))
		}

		// Create all relation for each task
		for kind, tasks := range t.RelatedTasks {

			if len(tasks) > 0 {
				log.Debugf("[creating structure] Creating %d related tasks for kind %v", len(tasks), kind)
			}

			for _, rt := range tasks {
				// First create the related tasks if they do not exist
				if _, exists := tasksByOldID[rt.ID]; !exists || rt.ID == 0 {
					oldid := rt.ID
					setBucketOrDefault(rt)
					rt.ProjectID = t.ProjectID
					err = rt.Create(s, user)
					if err != nil {
						return
					}
					tasksByOldID[oldid] = &models.TaskWithComments{Task: *rt}
					log.Debugf("[creating structure] Created related task %d", rt.ID)
				}

				// Then create the relation
				taskRel := &models.TaskRelation{
					TaskID:       t.ID,
					OtherTaskID:  rt.ID,
					RelationKind: kind,
				}
				if ttt, exists := tasksByOldID[rt.ID]; exists {
					taskRel.OtherTaskID = ttt.ID
				}
				err = taskRel.Create(s, user)
				if err != nil && !models.IsErrRelationAlreadyExists(err) {
					return
				}

				log.Debugf("[creating structure] Created task relation between task %d and %d", t.ID, rt.ID)

			}
		}

		// Create all attachments for each task
		if len(t.Attachments) > 0 {
			log.Debugf("[creating structure] Creating %d attachments", len(t.Attachments))
		}
		for _, a := range t.Attachments {
			// Check if we have a file to create
			if len(a.File.FileContent) > 0 {
				a.TaskID = t.ID
				fr := io.NopCloser(bytes.NewReader(a.File.FileContent))
				err = a.NewAttachment(s, fr, a.File.Name, a.File.Size, user)
				if err != nil {
					return
				}
				log.Debugf("[creating structure] Created new attachment %d", a.ID)
			}
		}

		// Create all labels
		for _, label := range t.Labels {
			// Check if we already have a label with that name + color combination and use it
			// If not, create one and save it for later
			var lb *models.Label
			var exists bool
			if label == nil {
				continue
			}
			lb, exists = labels[label.Title+label.HexColor]
			if !exists {
				err = label.Create(s, user)
				if err != nil {
					return err
				}
				log.Debugf("[creating structure] Created new label %d", label.ID)
				labels[label.Title+label.HexColor] = label
				lb = label
			}

			lt := &models.LabelTask{
				LabelID: lb.ID,
				TaskID:  t.ID,
			}
			err = lt.Create(s, user)
			if err != nil && !models.IsErrLabelIsAlreadyOnTask(err) {
				return err
			}
			log.Debugf("[creating structure] Associated task %d with label %d", t.ID, lb.ID)
		}

		for _, comment := range t.Comments {
			comment.TaskID = t.ID
			comment.ID = 0
			err = comment.Create(s, user)
			if err != nil {
				return
			}
			log.Debugf("[creating structure] Created new comment %d", comment.ID)
		}
	}

	// All tasks brought their own bucket with them, therefore the newly created default bucket is just extra space
	if !needsDefaultBucket {
		b := &models.Bucket{ProjectID: project.ID}
		bucketsIn, _, _, err := b.ReadAll(s, user, "", 1, 1)
		if err != nil {
			return err
		}
		buckets := bucketsIn.([]*models.Bucket)
		var newBacklogBucket *models.Bucket
		for _, b := range buckets {
			if b.Title == "Backlog" {
				newBacklogBucket = b
				break
			}
		}
		err = newBacklogBucket.Delete(s, user)
		if err != nil && !models.IsErrCannotRemoveLastBucket(err) {
			return err
		}
	}

	project.Tasks = tasks
	project.Buckets = originalBuckets

	return nil
}
