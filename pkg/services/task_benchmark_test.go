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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
)

// BenchmarkTaskService_AddLabelsToTasks benchmarks the label loading performance
// with and without caching
func BenchmarkTaskService_AddLabelsToTasks(b *testing.B) {
	// Setup test database
	db.LoadFixtures()
	s := db.NewSession()
	defer s.Close()

	ts := NewTaskService(db.GetEngine())

	// Get a list of task IDs from the database (using fixture data)
	var taskIDs []int64
	err := s.Table("tasks").
		Where("project_id = ?", 1).
		Limit(50).
		Cols("id").
		Find(&taskIDs)
	if err != nil {
		b.Fatalf("Failed to get task IDs: %v", err)
	}

	if len(taskIDs) == 0 {
		b.Skip("No tasks found in fixtures, skipping benchmark")
	}

	// Create task map
	taskMap := make(map[int64]*models.Task)
	for _, id := range taskIDs {
		taskMap[id] = &models.Task{ID: id}
	}

	b.ResetTimer()

	// Benchmark the current implementation
	for i := 0; i < b.N; i++ {
		// Reset labels to simulate fresh load
		for _, task := range taskMap {
			task.Labels = nil
		}

		err := ts.addLabelsToTasks(s, taskIDs, taskMap)
		if err != nil {
			b.Fatalf("addLabelsToTasks failed: %v", err)
		}
	}
}
