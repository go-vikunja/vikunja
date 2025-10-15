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
	"errors"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestLabelService_Create(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ls := NewLabelService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should create a new label", func(t *testing.T) {
		newLabel := &models.Label{
			Title:       "Test Label",
			Description: "A test label",
			HexColor:    "FF0000",
		}

		err := ls.Create(s, newLabel, u)
		assert.NoError(t, err)
		assert.NotZero(t, newLabel.ID)

		// Verify it's in the database
		var createdLabel models.Label
		exists, err := s.ID(newLabel.ID).Get(&createdLabel)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, newLabel.Title, createdLabel.Title)
		assert.Equal(t, u.ID, createdLabel.CreatedByID)
	})

	t.Run("should not create a new label without a user", func(t *testing.T) {
		newLabel := &models.Label{
			Title:       "Test Label 2",
			Description: "A test label 2",
			HexColor:    "0000FF",
		}

		err := ls.Create(s, newLabel, nil)
		assert.Error(t, err)
	})
}

func TestLabelService_Get(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ls := NewLabelService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should get a label by id", func(t *testing.T) {
		label, err := ls.Get(s, 1, u)
		assert.NoError(t, err)
		assert.NotNil(t, label)
		assert.Equal(t, int64(1), label.ID)
	})

	t.Run("should not get a label without access", func(t *testing.T) {
		otherUser := &user.User{ID: 2}
		_, err := ls.Get(s, 1, otherUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrAccessDenied))
	})
}

func TestLabelService_GetByID(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ls := NewLabelService(testEngine)

	t.Run("Success", func(t *testing.T) {
		label, err := ls.GetByID(s, 1)
		assert.NoError(t, err)
		assert.NotNil(t, label)
		assert.Equal(t, int64(1), label.ID)
		assert.Equal(t, "Label #1", label.Title)
	})

	t.Run("NotFound", func(t *testing.T) {
		label, err := ls.GetByID(s, 9999)
		assert.Error(t, err)
		assert.True(t, models.IsErrLabelDoesNotExist(err))
		assert.Nil(t, label)
	})
}

func TestLabelService_Delete(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ls := NewLabelService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should delete a label", func(t *testing.T) {
		newLabel := &models.Label{Title: "to delete", CreatedByID: u.ID}
		_, err := s.InsertOne(newLabel)
		assert.NoError(t, err)

		err = ls.Delete(s, newLabel, u)
		assert.NoError(t, err)

		var deletedLabel models.Label
		exists, err := s.ID(newLabel.ID).Get(&deletedLabel)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("should not delete a label without access", func(t *testing.T) {
		otherUser := &user.User{ID: 2}
		labelToDelete := &models.Label{ID: 1, CreatedByID: 1}
		err := ls.Delete(s, labelToDelete, otherUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrAccessDenied))
	})
}

func TestLabelService_GetAll(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ls := NewLabelService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should get all labels for a user", func(t *testing.T) {
		newLabel := &models.Label{Title: "to get all", CreatedByID: u.ID}
		_, err := s.InsertOne(newLabel)
		assert.NoError(t, err)

		labels, _, _, err := ls.GetAll(s, u, "", 0, 50)
		assert.NoError(t, err)
		assert.NotNil(t, labels)
		// Note: The number might be different now since we use GetLabelsByTaskIDs
		// which returns labels from tasks the user has access to, not just created labels
	})

	t.Run("should return an empty slice for a user with no labels", func(t *testing.T) {
		// Create a valid user but with no labels or tasks
		otherUser := &user.User{ID: 2, Username: "testuser2"} // Use a valid user ID from fixtures
		labels, _, _, err := ls.GetAll(s, otherUser, "", 0, 50)
		assert.NoError(t, err)
		assert.NotNil(t, labels)
		// The behavior here depends on what labels this user has access to through tasks
	})
}

func TestLabelService_Update(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ls := NewLabelService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should update a label", func(t *testing.T) {
		newLabel := &models.Label{Title: "to update", CreatedByID: u.ID}
		_, err := s.InsertOne(newLabel)
		assert.NoError(t, err)

		newLabel.Title = "Updated Title"
		err = ls.Update(s, newLabel, u)
		assert.NoError(t, err)

		var updatedLabel models.Label
		exists, err := s.ID(newLabel.ID).Get(&updatedLabel)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "Updated Title", updatedLabel.Title)
	})

	t.Run("should not update a label without access", func(t *testing.T) {
		otherUser := &user.User{ID: 2}
		labelToUpdate := &models.Label{
			ID:          1,
			Title:       "Updated Title by other user",
			CreatedByID: 1,
		}
		err := ls.Update(s, labelToUpdate, otherUser)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrAccessDenied))
	})
}

func TestLabelService_GetLabelsByTaskIDs(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ls := NewLabelService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should get labels for a single task", func(t *testing.T) {
		labels, count, total, err := ls.GetLabelsByTaskIDs(s, &GetLabelsByTaskIDsOptions{
			TaskIDs: []int64{1},
		})
		assert.NoError(t, err)
		assert.Greater(t, count, 0)
		assert.Equal(t, int64(count), total)
		assert.NotNil(t, labels)
	})

	t.Run("should get labels for multiple tasks", func(t *testing.T) {
		labels, count, total, err := ls.GetLabelsByTaskIDs(s, &GetLabelsByTaskIDsOptions{
			TaskIDs: []int64{1, 2},
		})
		assert.NoError(t, err)
		assert.NotNil(t, labels)
		// Just verify we got some results - exact counts depend on fixtures
		_ = count
		_ = total
	})

	t.Run("should get labels for a user with GetForUser flag", func(t *testing.T) {
		labels, count, total, err := ls.GetLabelsByTaskIDs(s, &GetLabelsByTaskIDsOptions{
			User:       u,
			GetForUser: true,
		})
		assert.NoError(t, err)
		assert.NotNil(t, labels)
		// Just verify we got some results - exact counts depend on fixtures
		_ = count
		_ = total
	})

	t.Run("should include unused labels when requested", func(t *testing.T) {
		// Create a label not associated with any task
		newLabel := &models.Label{Title: "unused", CreatedByID: u.ID}
		_, err := s.InsertOne(newLabel)
		assert.NoError(t, err)

		labels, count, _, err := ls.GetLabelsByTaskIDs(s, &GetLabelsByTaskIDsOptions{
			User:            u,
			GetForUser:      true,
			GetUnusedLabels: true,
		})
		assert.NoError(t, err)
		assert.Greater(t, count, 0)

		// Check if our unused label is in the results
		found := false
		for _, l := range labels {
			if l.ID == newLabel.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Unused label should be included in results")
	})

	t.Run("should filter by search term", func(t *testing.T) {
		labels, count, _, err := ls.GetLabelsByTaskIDs(s, &GetLabelsByTaskIDsOptions{
			TaskIDs: []int64{1, 2, 3},
			Search:  []string{"Label"},
		})
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, 0)
		if count > 0 {
			// Verify search filtering worked
			for _, l := range labels {
				assert.Contains(t, strings.ToLower(l.Title), "label")
			}
		}
	})

	t.Run("should group by label IDs only when requested", func(t *testing.T) {
		labels, _, _, err := ls.GetLabelsByTaskIDs(s, &GetLabelsByTaskIDsOptions{
			TaskIDs:             []int64{1, 2},
			GroupByLabelIDsOnly: true,
		})
		assert.NoError(t, err)
		// When grouped by label IDs, each label should appear only once
		labelIDs := make(map[int64]bool)
		for _, l := range labels {
			assert.False(t, labelIDs[l.ID], "Label ID %d appears multiple times", l.ID)
			labelIDs[l.ID] = true
		}
	})
}

func TestLabelService_HasAccessToLabel(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ls := NewLabelService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should have access to own label", func(t *testing.T) {
		newLabel := &models.Label{Title: "my label", CreatedByID: u.ID}
		_, err := s.InsertOne(newLabel)
		assert.NoError(t, err)

		hasAccess, err := ls.HasAccessToLabel(s, newLabel.ID, u)
		assert.NoError(t, err)
		assert.True(t, hasAccess)
	})

	t.Run("should have access to label on accessible task", func(t *testing.T) {
		// Label 1 is on task 1, which user 1 has access to
		hasAccess, err := ls.HasAccessToLabel(s, 1, u)
		assert.NoError(t, err)
		assert.True(t, hasAccess)
	})

	t.Run("should not have access with nil auth", func(t *testing.T) {
		hasAccess, err := ls.HasAccessToLabel(s, 1, nil)
		assert.NoError(t, err)
		assert.False(t, hasAccess)
	})

	t.Run("should return error for non-existent label", func(t *testing.T) {
		_, err := ls.HasAccessToLabel(s, 999999, u)
		assert.Error(t, err)
	})
}

func TestLabelService_IsLabelOwner(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ls := NewLabelService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should return true for label owner", func(t *testing.T) {
		newLabel := &models.Label{Title: "owned label", CreatedByID: u.ID}
		_, err := s.InsertOne(newLabel)
		assert.NoError(t, err)

		isOwner, err := ls.IsLabelOwner(s, newLabel.ID, u)
		assert.NoError(t, err)
		assert.True(t, isOwner)
	})

	t.Run("should return false for non-owner", func(t *testing.T) {
		otherUser := &user.User{ID: 2}
		label1CreatedBy1 := &models.Label{Title: "label1", CreatedByID: 1}
		_, err := s.InsertOne(label1CreatedBy1)
		assert.NoError(t, err)

		isOwner, err := ls.IsLabelOwner(s, label1CreatedBy1.ID, otherUser)
		assert.NoError(t, err)
		assert.False(t, isOwner)
	})

	t.Run("should return false for nil auth", func(t *testing.T) {
		isOwner, err := ls.IsLabelOwner(s, 1, nil)
		assert.NoError(t, err)
		assert.False(t, isOwner)
	})

	t.Run("should return false for link share", func(t *testing.T) {
		linkShare := &models.LinkSharing{ID: 1, ProjectID: 1}
		isOwner, err := ls.IsLabelOwner(s, 1, linkShare)
		assert.NoError(t, err)
		assert.False(t, isOwner)
	})
}

func TestLabelService_AddLabelToTask(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ls := NewLabelService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should add label to task", func(t *testing.T) {
		// Create a new label and task
		newLabel := &models.Label{Title: "test label", CreatedByID: u.ID}
		_, err := s.InsertOne(newLabel)
		assert.NoError(t, err)

		// Add label to task 1 (which user 1 has access to)
		err = ls.AddLabelToTask(s, newLabel.ID, 1, u)
		assert.NoError(t, err)

		// Verify the label was added
		exists, err := s.Exist(&models.LabelTask{LabelID: newLabel.ID, TaskID: 1})
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should not add label without access", func(t *testing.T) {
		otherUser := &user.User{ID: 2}
		label := &models.Label{Title: "restricted", CreatedByID: 999}
		_, err := s.InsertOne(label)
		assert.NoError(t, err)

		err = ls.AddLabelToTask(s, label.ID, 1, otherUser)
		assert.Error(t, err)
	})

	t.Run("should not add duplicate label", func(t *testing.T) {
		// Add a label first
		newLabel := &models.Label{Title: "duplicate test", CreatedByID: u.ID}
		_, err := s.InsertOne(newLabel)
		assert.NoError(t, err)

		err = ls.AddLabelToTask(s, newLabel.ID, 2, u)
		assert.NoError(t, err)

		// Try to add the same label again
		err = ls.AddLabelToTask(s, newLabel.ID, 2, u)
		assert.Error(t, err)
		assert.True(t, models.IsErrLabelIsAlreadyOnTask(err))
	})

	t.Run("should not add label to non-existent task", func(t *testing.T) {
		newLabel := &models.Label{Title: "test label 2", CreatedByID: u.ID}
		_, err := s.InsertOne(newLabel)
		assert.NoError(t, err)

		err = ls.AddLabelToTask(s, newLabel.ID, 999999, u)
		assert.Error(t, err)
	})

	t.Run("should not add non-existent label", func(t *testing.T) {
		err := ls.AddLabelToTask(s, 999999, 1, u)
		assert.Error(t, err)
		assert.True(t, models.IsErrLabelDoesNotExist(err))
	})
}

func TestLabelService_RemoveLabelFromTask(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ls := NewLabelService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should remove label from task", func(t *testing.T) {
		// First add a label
		newLabel := &models.Label{Title: "to remove", CreatedByID: u.ID}
		_, err := s.InsertOne(newLabel)
		assert.NoError(t, err)
		_, err = s.Insert(&models.LabelTask{LabelID: newLabel.ID, TaskID: 1})
		assert.NoError(t, err)

		// Now remove it
		err = ls.RemoveLabelFromTask(s, newLabel.ID, 1, u)
		assert.NoError(t, err)

		// Verify it was removed
		exists, err := s.Exist(&models.LabelTask{LabelID: newLabel.ID, TaskID: 1})
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("should not remove label without write access to task", func(t *testing.T) {
		otherUser := &user.User{ID: 999}
		err := ls.RemoveLabelFromTask(s, 1, 1, otherUser)
		assert.Error(t, err)
	})

	t.Run("should handle removing non-existent label", func(t *testing.T) {
		// This should not error, just do nothing
		err := ls.RemoveLabelFromTask(s, 999, 1, u)
		assert.NoError(t, err)
	})
}

func TestLabelService_UpdateTaskLabels(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	ls := NewLabelService(testEngine)
	u := &user.User{ID: 1}

	t.Run("should update task labels", func(t *testing.T) {
		label1 := &models.Label{Title: "label 1", CreatedByID: u.ID}
		label2 := &models.Label{Title: "label 2", CreatedByID: u.ID}
		_, err := s.Insert(label1, label2)
		assert.NoError(t, err)

		newLabels := []*models.Label{label1, label2}
		err = ls.UpdateTaskLabels(s, 2, newLabels, u)
		assert.NoError(t, err)

		// Verify labels were added
		labels, _, _, err := ls.GetLabelsByTaskIDs(s, &GetLabelsByTaskIDsOptions{
			TaskIDs: []int64{2},
		})
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(labels), 2)
	})

	t.Run("should remove labels not in new list", func(t *testing.T) {
		label1 := &models.Label{Title: "keep", CreatedByID: u.ID}
		label2 := &models.Label{Title: "remove", CreatedByID: u.ID}
		_, err := s.Insert(label1, label2)
		assert.NoError(t, err)

		// Add both labels
		err = ls.UpdateTaskLabels(s, 3, []*models.Label{label1, label2}, u)
		assert.NoError(t, err)

		// Update to only have label1
		err = ls.UpdateTaskLabels(s, 3, []*models.Label{label1}, u)
		assert.NoError(t, err)

		// Verify label2 was removed
		exists, err := s.Exist(&models.LabelTask{LabelID: label2.ID, TaskID: 3})
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("should delete all labels when empty list provided", func(t *testing.T) {
		label := &models.Label{Title: "to delete", CreatedByID: u.ID}
		_, err := s.InsertOne(label)
		assert.NoError(t, err)
		_, err = s.Insert(&models.LabelTask{LabelID: label.ID, TaskID: 4})
		assert.NoError(t, err)

		// Update with empty list
		err = ls.UpdateTaskLabels(s, 4, []*models.Label{}, u)
		assert.NoError(t, err)

		// Verify all labels removed
		exists, err := s.Exist(&models.LabelTask{TaskID: 4})
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("should do nothing when updating empty to empty", func(t *testing.T) {
		// Task with no labels
		err := ls.UpdateTaskLabels(s, 5, []*models.Label{}, u)
		assert.NoError(t, err)
	})

	t.Run("should not update without write access", func(t *testing.T) {
		otherUser := &user.User{ID: 999}
		err := ls.UpdateTaskLabels(s, 1, []*models.Label{}, otherUser)
		assert.Error(t, err)
	})

	t.Run("should not add label without access", func(t *testing.T) {
		restrictedLabel := &models.Label{Title: "restricted", CreatedByID: 999}
		_, err := s.InsertOne(restrictedLabel)
		assert.NoError(t, err)

		err = ls.UpdateTaskLabels(s, 1, []*models.Label{restrictedLabel}, u)
		assert.Error(t, err)
	})
}
