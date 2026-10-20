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
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
)

func MaxItems() int64 {
	if maxItems := config.MigrationMaxItems.GetInt64(); maxItems > 0 {
		return maxItems
	}
	return 100000
}

func CheckItemCount(count int64) error {
	if maxItems := MaxItems(); count > maxItems {
		return &ErrImportItemLimitExceeded{MaxItems: maxItems}
	}
	return nil
}

func CheckItemLimit(str []*models.ProjectWithTasksAndBuckets) error {
	var count int64
	var walk func(projects []*models.ProjectWithTasksAndBuckets) error
	walk = func(projects []*models.ProjectWithTasksAndBuckets) error {
		for _, p := range projects {
			if p == nil {
				continue
			}
			count += 1 + int64(len(p.Buckets)) + int64(len(p.Tasks))
			if err := CheckItemCount(count); err != nil {
				return err
			}
			if err := countTasks(p.Tasks, &count); err != nil {
				return err
			}
			if err := walk(p.ChildProjects); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(str)
}

// countTasks mirrors the queueing in createProjectWithEverything, so nested related tasks count too.
func countTasks(tasks []*models.TaskWithComments, count *int64) error {
	queued := make([]*models.Task, 0, len(tasks))
	queuedIDs := make(map[int64]bool, len(tasks))
	for _, t := range tasks {
		if t == nil {
			continue
		}
		*count += int64(len(t.Comments))
		if t.Title == "" {
			continue
		}
		queued = append(queued, &t.Task)
		if t.ID != 0 {
			queuedIDs[t.ID] = true
		}
	}
	for i := 0; i < len(queued); i++ {
		task := queued[i]
		*count += int64(len(task.Attachments)) + int64(len(task.Labels)) + int64(len(task.Assignees)) + int64(len(task.Reminders))
		for _, relatedTasks := range task.RelatedTasks {
			for _, related := range relatedTasks {
				if related == nil {
					continue
				}
				*count++
				if queuedIDs[related.ID] || related.Title == "" {
					continue
				}
				*count++
				queued = append(queued, related)
				if related.ID != 0 {
					queuedIDs[related.ID] = true
				}
			}
		}
		// Also stops a pointer cycle among id-less related tasks.
		if err := CheckItemCount(*count); err != nil {
			return err
		}
	}
	return nil
}
