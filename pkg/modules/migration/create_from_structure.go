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

package migration

import (
	"bytes"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"io/ioutil"
)

// InsertFromStructure takes a fully nested Vikunja data structure and a user and then creates everything for this user
// (Namespaces, tasks, etc. Even attachments and relations.)
func InsertFromStructure(str []*models.NamespaceWithLists, user *user.User) (err error) {

	log.Debugf("[creating structure] Creating %d namespaces", len(str))

	// Create all namespaces
	for _, n := range str {
		err = n.Create(user)
		if err != nil {
			return
		}

		log.Debugf("[creating structure] Created namespace %d", n.ID)
		log.Debugf("[creating structure] Creating %d lists", len(n.Lists))

		// Create all lists
		for _, l := range n.Lists {
			// The tasks slice is going to be reset during the creation of the list so we rescue it here to be able
			// to still loop over the tasks aftere the list was created.
			tasks := l.Tasks

			l.NamespaceID = n.ID
			err = l.Create(user)
			if err != nil {
				return
			}

			log.Debugf("[creating structure] Created list %d", l.ID)
			log.Debugf("[creating structure] Creating %d tasks", len(tasks))

			// Create all tasks
			for _, t := range tasks {
				t.ListID = l.ID
				err = t.Create(user)
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
							rt.ListID = t.ListID
							err = rt.Create(user)
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
						err = taskRel.Create(user)
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
						err = a.NewAttachment(fr, a.File.Name, a.File.Size, user)
						if err != nil {
							return
						}
						log.Debugf("[creating structure] Created new attachment %d", a.ID)
					}
				}
			}
		}
	}

	log.Debugf("[creating structure] Done inserting new task structure")

	return nil
}
