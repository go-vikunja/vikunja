package services

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
)

func TestTaskService_GetAllByProject_Sort(t *testing.T) {
	s := db.NewSession()
	defer s.Close()

	ts := NewTaskService()
	testUser := &user.User{ID: 1}

	// Create a new project for this test to avoid conflicts with fixture data
	testProject := &models.Project{
		Title:   "Sorting Test Project",
		OwnerID: testUser.ID,
	}
	err := testProject.Create(s, testUser)
	assert.NoError(t, err)

	// Create tasks with different due dates
	now := time.Now()
	task1 := &models.Task{Title: "Task 1", ProjectID: testProject.ID, DueDate: now.Add(24 * time.Hour)}
	task2 := &models.Task{Title: "Task 2", ProjectID: testProject.ID, DueDate: now}
	task3 := &models.Task{Title: "Task 3", ProjectID: testProject.ID, DueDate: now.Add(-24 * time.Hour)}

	err = task1.Create(s, testUser)
	assert.NoError(t, err)
	err = task2.Create(s, testUser)
	assert.NoError(t, err)
	err = task3.Create(s, testUser)
	assert.NoError(t, err)

	// Test sorting by due date ascending
	pagedResult, _, _, err := ts.GetByProject(s, testProject.ID, testUser, "", 1, 10, TaskOptions{TaskSortBy: models.TaskSortBy{SortBy: []string{"due_date"}}})
	assert.NoError(t, err)
	tasks, ok := pagedResult.([]*models.Task)
	assert.True(t, ok)
	assert.Len(t, tasks, 3)
	assert.Equal(t, task3.ID, tasks[0].ID)
	assert.Equal(t, task2.ID, tasks[1].ID)
	assert.Equal(t, task1.ID, tasks[2].ID)

	// Test sorting by due date descending
	pagedResult, _, _, err = ts.GetByProject(s, testProject.ID, testUser, "", 1, 10, TaskOptions{TaskSortBy: models.TaskSortBy{SortBy: []string{"-due_date"}}})
	assert.NoError(t, err)
	tasks, ok = pagedResult.([]*models.Task)
	assert.True(t, ok)
	assert.Len(t, tasks, 3)
	assert.Equal(t, task1.ID, tasks[0].ID)
	assert.Equal(t, task2.ID, tasks[1].ID)
	assert.Equal(t, task3.ID, tasks[2].ID)
}
