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

package migration

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/background/handler"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"
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
		if p.ID == models.FavoritesPseudoProjectID {
			continue
		}

		oldID := p.ID

		if p.ParentProjectID != 0 {
			childRelations[p.ParentProjectID] = append(childRelations[p.ParentProjectID], oldID)
			p.ParentProjectID = 0
		}

		p.ID = 0

		for _, view := range p.Views {
			view.ProjectID = 0
		}

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

func createProject(s *xorm.Session, project *models.ProjectWithTasksAndBuckets, archivedProjectIDs *[]int64, labels map[string]*models.Label, authUser *user.User) (err error) {
	err = createProjectWithEverything(s, project, archivedProjectIDs, labels, authUser)
	if err != nil {
		return err
	}

	log.Debugf("[creating structure] Created project %d", project.ID)

	return
}

func buildLabelCacheKey(label *models.Label) string {
	return label.Title + "|" + utils.NormalizeHex(label.HexColor)
}

func createLabelForMigration(s *xorm.Session, label *models.Label, creator *user.User) (*models.Label, error) {
	normalizedHex := utils.NormalizeHex(label.HexColor)
	createdAt := label.Created
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := label.Updated
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	insertData := map[string]interface{}{
		"title":         label.Title,
		"description":   label.Description,
		"hex_color":     normalizedHex,
		"created_by_id": creator.ID,
		"created":       createdAt.UTC(),
		"updated":       updatedAt.UTC(),
	}

	if _, err := s.Table("labels").Insert(insertData); err != nil {
		return nil, err
	}

	rows, err := s.Table("labels").Select("id").Where("created_by_id = ? AND title = ? AND hex_color = ?", creator.ID, label.Title, normalizedHex).Desc("id").Limit(1).QueryInterface()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("failed to load inserted label id")
	}

	rowID := rows[0]["id"]
	var labelID int64
	switch v := rowID.(type) {
	case int64:
		labelID = v
	case int:
		labelID = int64(v)
	case int32:
		labelID = int64(v)
	case []uint8:
		parsed, parseErr := strconv.ParseInt(string(v), 10, 64)
		if parseErr != nil {
			return nil, parseErr
		}
		labelID = parsed
	default:
		return nil, errors.New("unsupported label id type")
	}

	createdLabel := &models.Label{
		ID:          labelID,
		Title:       label.Title,
		Description: label.Description,
		HexColor:    normalizedHex,
		CreatedByID: creator.ID,
		Created:     createdAt.UTC(),
		Updated:     updatedAt.UTC(),
		CreatedBy:   creator,
	}

	return createdLabel, nil
}

func createCommentForMigration(s *xorm.Session, comment *models.TaskComment, task *models.Task, auth web.Auth) error {
	author, err := models.GetUserOrLinkShareUser(s, auth)
	if err != nil {
		return err
	}
	if author == nil {
		return errors.New("author is required for comment migration")
	}

	createdAt := comment.Created
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	} else {
		createdAt = createdAt.UTC()
	}

	updatedAt := comment.Updated
	if updatedAt.IsZero() {
		updatedAt = createdAt
	} else {
		updatedAt = updatedAt.UTC()
	}

	result, err := s.Exec(
		"INSERT INTO task_comments (comment, author_id, task_id, created, updated) VALUES (?, ?, ?, ?, ?)",
		comment.Comment,
		author.ID,
		task.ID,
		createdAt,
		updatedAt,
	)
	if err != nil {
		return err
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	comment.ID = commentID
	comment.TaskID = task.ID
	comment.AuthorID = author.ID
	comment.Author = author
	comment.Created = createdAt
	comment.Updated = updatedAt

	return events.Dispatch(&models.TaskCommentCreatedEvent{
		Task:    task,
		Comment: comment,
		Doer:    author,
	})
}

func createProjectWithEverything(s *xorm.Session, project *models.ProjectWithTasksAndBuckets, archivedProjects *[]int64, labels map[string]*models.Label, authUser *user.User) (err error) {
	// The tasks and bucket slices are going to be reset during the creation of the project, so we rescue it here
	// to be able to still loop over them aftere the project was created.
	tasks := project.Tasks
	originalBuckets := project.Buckets
	originalBackgroundInformation := project.BackgroundInformation
	originalAttachmentCreateFunc := models.AttachmentCreateFunc
	if originalAttachmentCreateFunc != nil {
		models.AttachmentCreateFunc = func(sess *xorm.Session, attachment *models.TaskAttachment, reader io.ReadCloser, filename string, size uint64, creator *user.User) (*models.TaskAttachment, error) {
			defer reader.Close()

			now := time.Now().UTC()
			result, err := sess.Exec("INSERT INTO files (name, mime, size, created, created_by_id) VALUES (?, ?, ?, ?, ?)", filename, "", size, now, creator.ID)
			if err != nil {
				return nil, err
			}

			fileID, err := result.LastInsertId()
			if err != nil {
				return nil, err
			}

			file := &files.File{
				ID:          fileID,
				Name:        filename,
				Mime:        "",
				Size:        size,
				Created:     now,
				CreatedByID: creator.ID,
			}

			if err := file.Save(reader); err != nil {
				return nil, err
			}

			attachment.File = file
			attachment.FileID = file.ID
			attachment.CreatedBy = creator
			attachment.CreatedByID = creator.ID

			_, err = sess.Exec("INSERT INTO task_attachments (task_id, file_id, created_by_id, created) VALUES (?, ?, ?, ?)", attachment.TaskID, attachment.FileID, attachment.CreatedByID, now)
			if err != nil {
				return nil, err
			}

			return attachment, nil
		}
		defer func() {
			models.AttachmentCreateFunc = originalAttachmentCreateFunc
		}()
	}
	needsDefaultBucket := false
	oldViews := project.Views

	// Saving the archived status to archive the project again after creating it
	var wasArchived bool
	if project.IsArchived {
		wasArchived = true
		project.IsArchived = false
	}

	project.ID = 0
	err = models.CreateProject(s, &project.Project, authUser, false, false)
	if err != nil && models.IsErrProjectIdentifierIsNotUnique(err) {
		project.Identifier = ""
		err = models.CreateProject(s, &project.Project, authUser, false, false)
	}
	if err != nil {
		return errors.New("error creating project: " + err.Error())
	}

	if wasArchived {
		*archivedProjects = append(*archivedProjects, project.ID)
	}

	log.Debugf("[creating structure] Created project %d", project.ID)

	log.Debugf("[creating structure] About to handle background file")
	bf, is := originalBackgroundInformation.(*bytes.Buffer)
	if is {

		backgroundFile := bytes.NewReader(bf.Bytes())

		log.Debugf("[creating structure] Creating a background file for project %d", project.ID)

		err = handler.SaveBackgroundFile(s, authUser, &project.Project, backgroundFile, "", uint64(backgroundFile.Len()))
		if err != nil {
			log.Errorf("[creating structure] Could not create background for project %d, error was %v", project.ID, err)
		}

		log.Debugf("[creating structure] Created a background file for project %d", project.ID)
	}

	log.Debugf("[creating structure] About to create buckets")
	// Create all buckets
	bucketsByOldID := make(map[int64]*models.Bucket) // old bucket id is the key
	if len(project.Buckets) > 0 {
		log.Debugf("[creating structure] Creating %d buckets", len(project.Buckets))
	}

	for _, bucket := range originalBuckets {
		if _, exists := bucketsByOldID[bucket.ID]; exists {
			continue
		}

		oldID := bucket.ID
		log.Debugf("[creating structure] Creating bucket with original ProjectViewID: %d", bucket.ProjectViewID)

		// Reset invalid ProjectViewIDs to 0 during migration
		// They will be assigned to the correct view later
		if bucket.ProjectViewID != 0 {
			bucket.ProjectViewID = 0
		}

		bucket.ID = 0 // We want a new id
		bucket.ProjectID = project.ID

		// Use direct database insert during migration to avoid service layer validation
		// The ProjectViewID will be set correctly later after views are created
		bucket.CreatedBy, err = models.GetUserOrLinkShareUser(s, authUser)
		if err != nil {
			return
		}
		bucket.CreatedByID = bucket.CreatedBy.ID

		_, err = s.Insert(bucket)
		if err != nil {
			return
		}
		fmt.Printf("inserted bucket old=%d new=%d title=%s\n", oldID, bucket.ID, bucket.Title)

		// Update position using the same logic as the model's Create method
		if bucket.Position == 0 {
			bucket.Position = float64(bucket.ID)
		}
		_, err = s.Where("id = ?", bucket.ID).Cols("position").Update(bucket)
		if err != nil {
			return
		}

		bucketsByOldID[oldID] = bucket
		log.Debugf("[creating structure] Created bucket %d, old ID was %d", bucket.ID, oldID)
	}

	log.Debugf("[creating structure] Finished creating buckets, about to create views")
	// Create all views, create default views if we don't have any
	viewsByOldIDs := make(map[int64]*models.ProjectView, len(oldViews))
	if len(oldViews) > 0 {
		log.Debugf("[creating structure] Creating %d existing views", len(oldViews))
		for _, view := range oldViews {
			oldID := view.ID
			view.ID = 0

			if view.DefaultBucketID != 0 {
				bucket, has := bucketsByOldID[view.DefaultBucketID]
				if has {
					view.DefaultBucketID = bucket.ID
				}
			}

			if view.DoneBucketID != 0 {
				bucket, has := bucketsByOldID[view.DoneBucketID]
				if has {
					view.DoneBucketID = bucket.ID
				}
			}

			view.ProjectID = project.ID

			log.Debugf("[creating structure] About to create view: %s", view.Title)
			err = view.Create(s, authUser)
			if err != nil {
				log.Errorf("[creating structure] Error creating view %s: %v", view.Title, err)
				return
			}
			// Remove default buckets created automatically during view creation; we'll reimport them from the export data
			_, err = s.Where("project_view_id = ?", view.ID).Delete(&models.Bucket{})
			if err != nil {
				return
			}
			view.DefaultBucketID = 0
			view.DoneBucketID = 0
			_, err = s.ID(view.ID).Cols("default_bucket_id", "done_bucket_id").Update(view)
			if err != nil {
				return
			}
			viewsByOldIDs[oldID] = view
		}

		for oldID, bucket := range bucketsByOldID {
			newView, has := viewsByOldIDs[bucket.ProjectViewID]
			if !has {
				err = bucket.Delete(s, authUser)
				if err != nil {
					return
				}
				delete(bucketsByOldID, oldID)
				continue
			}

			bucket.ProjectViewID = newView.ID
			// Use direct database update during migration to avoid service layer validation
			_, err = s.Where("id = ?", bucket.ID).Cols("project_view_id").Update(bucket)
			if err != nil {
				return
			}
		}
	} else {
		if len(project.Views) == 0 {
			err = models.CreateDefaultViewsForProject(s, &project.Project, authUser, false, true)
			if err != nil {
				return
			}
		}

		// Only using the default views
		// Add all buckets to the default kanban view
		for _, view := range project.Views {
			if view.ViewKind == models.ProjectViewKindKanban {
				for _, b := range bucketsByOldID {
					b.ProjectViewID = view.ID
					// Use direct database update during migration to avoid service layer validation
					_, err = s.Where("id = ?", b.ID).Cols("project_view_id").Update(b)
					if err != nil {
						return
					}
				}
				break
			}
		}

	}

	log.Debugf("[creating structure] Creating %d tasks", len(tasks))

	setBucketOrDefault := func(task *models.Task) (err error) {
		var bucketID = task.BucketID
		bucket, exists := bucketsByOldID[bucketID]
		if exists {
			bucketID = bucket.ID
			tb := &models.TaskBucket{
				TaskID:        task.ID,
				BucketID:      bucketID,
				ProjectID:     task.ProjectID,
				ProjectViewID: bucket.ProjectViewID,
			}
			// Use direct database operation during migration to avoid service layer validation
			_, err = s.Exec("DELETE FROM task_buckets WHERE task_id = ? AND project_view_id = ?", tb.TaskID, tb.ProjectViewID)
			if err != nil {
				log.Debugf("[creating structure] Error while deleting old task bucket for task %d: %s", task.ID, err.Error())
				return
			}
			_, err = s.Exec("INSERT INTO task_buckets (bucket_id, task_id, project_view_id) VALUES (?, ?, ?)", tb.BucketID, tb.TaskID, tb.ProjectViewID)
			if err != nil {
				log.Debugf("[creating structure] Error while inserting task bucket %d for task %d: %s", bucketID, task.ID, err.Error())
				return
			}
		} else if bucketID > 0 {
			log.Debugf("[creating structure] No bucket created for original bucket id %d", task.BucketID)
			bucketID = 0
		}
		if !exists || bucketID == 0 {
			fmt.Printf("needsDefaultBucket: task %d origBucket %d exists=%v\n", task.ID, task.BucketID, exists)
			needsDefaultBucket = true
		}

		return
	}

	tasksByOldID := make(map[int64]*models.TaskWithComments, len(tasks))
	newTaskIDs := []int64{}
	// Create all tasks
	for i, t := range tasks {
		oldid := t.ID
		t.ID = 0
		t.ProjectID = project.ID
		originalBucketID := t.BucketID
		t.BucketID = 0

		log.Debugf("[creating structure] Creating task (old id %d, title %s)", oldid, t.Title)
		err = t.Create(s, authUser)
		if err != nil && models.IsErrTaskCannotBeEmpty(err) {
			continue
		}
		if err != nil {
			return err
		}

		t.BucketID = originalBucketID

		err = setBucketOrDefault(&tasks[i].Task)
		if err != nil {
			return
		}

		newTaskIDs = append(newTaskIDs, t.ID)

		tasksByOldID[oldid] = t

		log.Debugf("[creating structure] Created task %d (old id %d)", t.ID, oldid)
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
					rt.ID = 0
					rt.ProjectID = t.ProjectID
					originalBucketID := rt.BucketID
					rt.BucketID = 0

					err = rt.Create(s, authUser)
					if err != nil {
						log.Debugf("[creating structure] Error while creating related task %d: %s", rt.ID, err.Error())
						return
					}

					rt.BucketID = originalBucketID

					err = setBucketOrDefault(rt)
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

				// Add this check to prevent self-relations
				if taskRel.TaskID == taskRel.OtherTaskID {
					log.Debugf("[creating structure] Skipping invalid self-relation for task %d", taskRel.TaskID)
					continue
				}

				err = taskRel.Create(s, authUser)
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
				oldID := a.ID
				a.ID = 0
				a.TaskID = t.ID
				fr := io.NopCloser(bytes.NewReader(a.File.FileContent))
				err = a.NewAttachment(s, fr, a.File.Name, a.File.Size, authUser)
				if err != nil {
					if models.IsErrTaskAttachmentIsTooLarge(err) {
						log.Warningf("[creating structure] Attachment %s is too large (%d bytes), skipping: %v", a.File.Name, a.File.Size, err)
						continue
					}
					return
				}
				log.Debugf("[creating structure] Created new attachment %d", a.ID)

				if t.CoverImageAttachmentID == oldID {
					t.CoverImageAttachmentID = a.ID
					_, err = s.Exec("UPDATE tasks SET cover_image_attachment_id = ? WHERE id = ?", t.CoverImageAttachmentID, t.ID)
					if err != nil {
						return
					}
				}
			}
		}

		// Create all labels
		for _, label := range t.Labels {
			if label == nil {
				continue
			}

			cacheKey := buildLabelCacheKey(label)
			lb, exists := labels[cacheKey]
			if !exists {
				lb, err = createLabelForMigration(s, label, authUser)
				if err != nil {
					return err
				}
				log.Debugf("[creating structure] Created new label %d", lb.ID)
				labels[cacheKey] = lb
			}

			label.ID = lb.ID
			label.CreatedBy = lb.CreatedBy
			label.CreatedByID = lb.CreatedByID
			label.Created = lb.Created
			label.Updated = lb.Updated
			label.HexColor = lb.HexColor

			exists, err := s.Table("label_tasks").Where("label_id = ? AND task_id = ?", lb.ID, t.ID).Exist()
			if err != nil {
				return err
			}
			if exists {
				log.Debugf("[creating structure] Label %d already associated with task %d", lb.ID, t.ID)
				continue
			}

			_, err = s.Table("label_tasks").Insert(map[string]interface{}{
				"label_id": lb.ID,
				"task_id":  t.ID,
				"created":  time.Now().UTC(),
			})
			if err != nil {
				return err
			}
			log.Debugf("[creating structure] Associated task %d with label %d", t.ID, lb.ID)
		}

		// Comments
		for _, comment := range t.Comments {
			comment.TaskID = t.ID
			comment.ID = 0
			err = createCommentForMigration(s, comment, &t.Task, authUser)
			if err != nil {
				return
			}
			log.Debugf("[creating structure] Created new comment %d", comment.ID)
		}
	}

	// All tasks brought their own bucket with them, therefore the newly created default bucket is just extra space
	if !needsDefaultBucket {
		b := &models.Bucket{ProjectID: project.ID}

		for _, view := range project.Views {
			if view.ViewKind == models.ProjectViewKindKanban {
				b.ProjectViewID = view.ID
				break
			}
		}

		bucketsIn, _, _, err := b.ReadAll(s, authUser, "", 1, 1)
		if err != nil {
			return err
		}
		buckets := bucketsIn.([]*models.Bucket)
		var newBacklogBucket *models.Bucket
		for _, b := range buckets {
			if b.Title == "To-Do" {
				newBacklogBucket = b
				newBacklogBucket.ProjectID = project.ID
				break
			}
		}
		// Only attempt to delete if we found a "To-Do" bucket
		// (may have been already deleted during view cleanup)
		if newBacklogBucket != nil {
			err = newBacklogBucket.Delete(s, authUser)
			if err != nil && !models.IsErrCannotRemoveLastBucket(err) {
				return err
			}
		}
	}

	if len(viewsByOldIDs) > 0 {
		newPositions := []*models.TaskPosition{}
		for _, pos := range project.Positions {
			_, hasTask := tasksByOldID[pos.TaskID]
			_, hasView := viewsByOldIDs[pos.ProjectViewID]
			if !hasTask || !hasView {
				continue
			}
			newPositions = append(newPositions, &models.TaskPosition{
				TaskID:        tasksByOldID[pos.TaskID].ID,
				ProjectViewID: viewsByOldIDs[pos.ProjectViewID].ID,
				Position:      pos.Position,
			})
		}

		if len(newPositions) > 0 {
			_, err = s.In("task_id", newTaskIDs).Delete(&models.TaskPosition{})
			if err != nil {
				return
			}
			_, err = s.Insert(newPositions)
			if err != nil {
				return
			}
		}

		newTaskBuckets := make([]*models.TaskBucket, 0, len(project.TaskBuckets))
		for _, tb := range project.TaskBuckets {
			_, hasTask := tasksByOldID[tb.TaskID]
			_, hasBucket := bucketsByOldID[tb.BucketID]
			if !hasTask || !hasBucket {
				continue
			}
			newTaskBuckets = append(newTaskBuckets, &models.TaskBucket{
				TaskID:        tasksByOldID[tb.TaskID].ID,
				BucketID:      bucketsByOldID[tb.BucketID].ID,
				ProjectViewID: bucketsByOldID[tb.BucketID].ProjectViewID,
			})
		}

		if len(newTaskBuckets) > 0 {
			_, err = s.In("task_id", newTaskIDs).Delete(&models.TaskBucket{})
			if err != nil {
				return
			}
			_, err = s.Insert(newTaskBuckets)
			if err != nil {
				return
			}
		}
	}

	project.Tasks = tasks
	project.Buckets = originalBuckets

	return nil
}
