package services

import (
	"errors"
	"time"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"
)

// TaskService is a service for tasks.
type TaskService struct {
}

// Create creates a new task.
func (ts *TaskService) Create(s *xorm.Session, t *models.Task, u *user.User) error {
	if t.ProjectID == 0 {
		return errors.New("a task needs to be in a project")
	}

	projectService := &Project{}
	can, err := projectService.HasPermission(s, t.ProjectID, u, models.PermissionWrite)
	if err != nil {
		return err
	}
	if !can {
		return &models.ErrGenericForbidden{}
	}

	t.CreatedByID = u.ID
	t.Created = time.Now()

	// Handle Identifier Index
	var maxIndex int64
	_, err = s.SQL("SELECT MAX(`index`) FROM `tasks` WHERE `project_id` = ?", t.ProjectID).Get(&maxIndex)
	if err != nil {
		return err
	}
	t.Index = maxIndex + 1

	// Perform Database Insert
	if _, err := s.Insert(t); err != nil {
		return err
	}

	// Handle Labels
	if len(t.Labels) > 0 {
		for _, label := range t.Labels {
			labelTask := &models.LabelTask{
				TaskID:  t.ID,
				LabelID: label.ID,
			}
			if _, err := s.Insert(labelTask); err != nil {
				return err
			}
		}
	}

	// Handle Assignees
	if len(t.Assignees) > 0 {
		for _, assignee := range t.Assignees {
			// Check if the user has access to the project.
			canRead, err := projectService.HasPermission(s, t.ProjectID, assignee, models.PermissionRead)
			if err != nil {
				return err
			}
			if !canRead {
				return &models.ErrUserDoesNotHaveAccessToProject{ProjectID: t.ProjectID, UserID: assignee.ID}
			}

			taskAssignee := &models.TaskAssginee{
				TaskID: t.ID,
				UserID: assignee.ID,
			}
			if _, err := s.Insert(taskAssignee); err != nil {
				return err
			}

			// Create subscription
			sub := &models.Subscription{
				UserID:     assignee.ID,
				EntityType: models.SubscriptionEntityTask,
				EntityID:   t.ID,
			}
			if err := sub.Create(s, assignee); err != nil && !models.IsErrSubscriptionAlreadyExists(err) {
				return err
			}
		}
	}

	// Handle Attachments
	if len(t.Attachments) > 0 {
		for _, attachment := range t.Attachments {
			if attachment.File == nil || attachment.File.File == nil {
				continue
			}

			file, err := files.Create(attachment.File.File, attachment.File.Name, attachment.File.Size, u)
			if err != nil {
				return err
			}

			attachment.FileID = file.ID
			attachment.TaskID = t.ID
			attachment.CreatedByID = u.ID
			if _, err := s.Insert(attachment); err != nil {
				return err
			}
		}
	}

	// Update Project Timestamp
	if _, err := s.ID(t.ProjectID).Cols("updated").Update(&models.Project{}); err != nil {
		return err
	}

	// Dispatch Event
	err = events.Dispatch(&models.TaskCreatedEvent{
		Task: t,
		Doer: u,
	})
	if err != nil {
		return err
	}

	return nil
}
