// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
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
	"io/ioutil"

	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
)

// InsertFromStructure takes a fully nested Vikunja data structure and a user and then creates everything for this user
// (Namespaces, tasks, etc. Even attachments and relations.)
func InsertFromStructure(str []*models.NamespaceWithListsAndTasks, user *user.User) (err error) {
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

func insertFromStructure(s *xorm.Session, str []*models.NamespaceWithListsAndTasks, user *user.User) (err error) {

	log.Debugf("[creating structure] Creating %d namespaces", len(str))

	labels := make(map[string]*models.Label)

	archivedLists := []int64{}
	archivedNamespaces := []int64{}

	// Create all namespaces
	for _, n := range str {
		n.ID = 0

		// Saving the archived status to archive the namespace again after creating it
		var wasArchived bool
		if n.IsArchived {
			n.IsArchived = false
			wasArchived = true
		}

		err = n.Create(s, user)
		if err != nil {
			return
		}

		if wasArchived {
			archivedNamespaces = append(archivedNamespaces, n.ID)
		}

		log.Debugf("[creating structure] Created namespace %d", n.ID)
		log.Debugf("[creating structure] Creating %d lists", len(n.Lists))

		// Create all lists
		for _, l := range n.Lists {
			// The tasks and bucket slices are going to be reset during the creation of the list so we rescue it here
			// to be able to still loop over them aftere the list was created.
			tasks := l.Tasks
			originalBuckets := l.Buckets
			originalBackgroundInformation := l.BackgroundInformation
			needsDefaultBucket := false

			// Saving the archived status to archive the list again after creating it
			var wasArchived bool
			if l.IsArchived {
				wasArchived = true
				l.IsArchived = false
			}

			l.NamespaceID = n.ID
			l.ID = 0
			err = l.Create(s, user)
			if err != nil {
				return
			}

			if wasArchived {
				archivedLists = append(archivedLists, l.ID)
			}

			log.Debugf("[creating structure] Created list %d", l.ID)

			backgroundFile, is := originalBackgroundInformation.(*bytes.Buffer)
			if is {
				log.Debugf("[creating structure] Creating a background file for list %d", l.ID)

				file, err := files.Create(backgroundFile, "", uint64(backgroundFile.Len()), user)
				if err != nil {
					return err
				}

				err = models.SetListBackground(s, l.ID, file)
				if err != nil {
					return err
				}

				log.Debugf("[creating structure] Created a background file as new file %d for list %d", file.ID, l.ID)
			}

			// Create all buckets
			buckets := make(map[int64]*models.Bucket) // old bucket id is the key
			if len(l.Buckets) > 0 {
				log.Debugf("[creating structure] Creating %d buckets", len(l.Buckets))
			}
			for _, bucket := range originalBuckets {
				oldID := bucket.ID
				bucket.ID = 0 // We want a new id
				bucket.ListID = l.ID
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

			// Create all tasks
			for _, t := range tasks {
				setBucketOrDefault(&t.Task)

				t.ListID = l.ID
				err = t.Create(s, user)
				if err != nil {
					return
				}

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
						if rt.ID == 0 {
							setBucketOrDefault(rt)
							rt.ListID = t.ListID
							err = rt.Create(s, user)
							if err != nil {
								return
							}
							log.Debugf("[creating structure] Created related task %d", rt.ID)
						}

						// Then create the relation
						taskRel := &models.TaskRelation{
							TaskID:       t.ID,
							OtherTaskID:  rt.ID,
							RelationKind: kind,
						}
						err = taskRel.Create(s, user)
						if err != nil {
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
						fr := ioutil.NopCloser(bytes.NewReader(a.File.FileContent))
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
					err = comment.Create(s, user)
					if err != nil {
						return
					}
					log.Debugf("[creating structure] Created new comment %d", comment.ID)
				}
			}

			// All tasks brought their own bucket with them, therefore the newly created default bucket is just extra space
			if !needsDefaultBucket {
				b := &models.Bucket{ListID: l.ID}
				bucketsIn, _, _, err := b.ReadAll(s, user, "", 1, 1)
				if err != nil {
					return err
				}
				buckets := bucketsIn.([]*models.Bucket)
				err = buckets[0].Delete(s, user)
				if err != nil && !models.IsErrCannotRemoveLastBucket(err) {
					return err
				}
			}

			l.Tasks = tasks
			l.Buckets = originalBuckets
		}
	}

	if len(archivedLists) > 0 {
		_, err = s.
			Cols("is_archived").
			In("id", archivedLists).
			Update(&models.List{IsArchived: true})
		if err != nil {
			return err
		}
	}

	if len(archivedNamespaces) > 0 {
		_, err = s.
			Cols("is_archived").
			In("id", archivedNamespaces).
			Update(&models.Namespace{IsArchived: true})
		if err != nil {
			return err
		}
	}

	log.Debugf("[creating structure] Done inserting new task structure")

	return nil
}
