package services

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestTaskService_Create(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	ts := &TaskService{}
	u := &user.User{ID: 1}
	projectWithIdentifier := &models.Project{ID: 1} // From fixtures

	t.Run("Create a simple task", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		var maxIndex int64
		_, err := s.SQL("SELECT MAX(`index`) FROM `tasks` WHERE `project_id` = ?", projectWithIdentifier.ID).Get(&maxIndex)
		assert.NoError(t, err)

		task := &models.Task{
			Title:     "Simple Test Task",
			ProjectID: projectWithIdentifier.ID,
		}

		err = ts.Create(s, task, u)
		assert.NoError(t, err)
		assert.NotZero(t, task.ID, "Task ID should be set")
		assert.Equal(t, maxIndex+1, task.Index, "Task Index should be the next available one")

		// Verify in DB
		var savedTask models.Task
		has, err := s.ID(task.ID).Get(&savedTask)
		assert.NoError(t, err)
		assert.True(t, has, "Task should be saved in the database")
		assert.Equal(t, task.Title, savedTask.Title)
		assert.Equal(t, u.ID, savedTask.CreatedByID)

		// Verify event
		events.AssertDispatched(t, &models.TaskCreatedEvent{})
	})

	t.Run("Create a task with labels", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		task := &models.Task{
			Title:     "Task with Labels",
			ProjectID: projectWithIdentifier.ID,
			Labels: []*models.Label{
				{ID: 1}, // from fixtures
				{ID: 2}, // from fixtures
			},
		}
		err := ts.Create(s, task, u)
		assert.NoError(t, err)

		// Verify label associations
		var labelTasks []models.LabelTask
		err = s.Where("task_id = ?", task.ID).Find(&labelTasks)
		assert.NoError(t, err)
		assert.Len(t, labelTasks, 2)
	})

	t.Run("Create a task with assignees", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		assignee := &user.User{ID: 2} // from fixtures
		_, err := s.Insert(&models.ProjectUser{UserID: assignee.ID, ProjectID: projectWithIdentifier.ID, Permission: models.PermissionWrite})
		assert.NoError(t, err)

		task := &models.Task{
			Title:     "Task with Assignees",
			ProjectID: projectWithIdentifier.ID,
			Assignees: []*user.User{assignee},
		}

		err = ts.Create(s, task, u)
		assert.NoError(t, err)

		// Verify assignee association
		var taskAssignee models.TaskAssginee
		has, err := s.Where("task_id = ? AND user_id = ?", task.ID, assignee.ID).Get(&taskAssignee)
		assert.NoError(t, err)
		assert.True(t, has)

		// Verify subscription
		var subscription models.Subscription
		has, err = s.Where("entity_type = ? AND entity_id = ? AND user_id = ?", models.SubscriptionEntityTask, task.ID, assignee.ID).Get(&subscription)
		assert.NoError(t, err)
		assert.True(t, has)
	})

	t.Run("Create a task with an attachment", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		fs := afero.NewMemMapFs()
		aferoFile, _ := fs.Create("test.txt")
		fileContent := "hello world"
		_, _ = aferoFile.WriteString(fileContent)
		_ = aferoFile.Sync()
		_, _ = aferoFile.Seek(0, 0)

		task := &models.Task{
			Title:     "Task with Attachment",
			ProjectID: projectWithIdentifier.ID,
			Attachments: []*models.TaskAttachment{
				{
					File: &files.File{
						File: aferoFile,
						Name: "test.txt",
						Size: uint64(len(fileContent)),
					},
				},
			},
		}

		err := ts.Create(s, task, u)
		assert.NoError(t, err)

		// Verify attachment
		var attachment models.TaskAttachment
		has, err := s.Where("task_id = ?", task.ID).Get(&attachment)
		assert.NoError(t, err)
		assert.True(t, has)
		assert.NotZero(t, attachment.FileID)

		// Verify file
		var file files.File
		has, err = s.ID(attachment.FileID).Get(&file)
		assert.NoError(t, err)
		assert.True(t, has)
		assert.Equal(t, "test.txt", file.Name)
	})

	t.Run("Creating a task should update project timestamp", func(t *testing.T) {
		s := db.NewSession()
		defer s.Close()

		var projectBefore models.Project
		_, err := s.ID(projectWithIdentifier.ID).Get(&projectBefore)
		assert.NoError(t, err)

		time.Sleep(1 * time.Second) // Ensure timestamp is different

		task := &models.Task{
			Title:     "Timestamp Test Task",
			ProjectID: projectWithIdentifier.ID,
		}
		err = ts.Create(s, task, u)
		assert.NoError(t, err)

		var projectAfter models.Project
		_, err = s.ID(projectWithIdentifier.ID).Get(&projectAfter)
		assert.NoError(t, err)

		assert.True(t, projectAfter.Updated.After(projectBefore.Updated), "Project Updated timestamp should be newer")
	})
}
