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

		labels, err := ls.GetAll(s, u)
		assert.NoError(t, err)
		assert.NotNil(t, labels)
		assert.Len(t, labels, 3) // 2 from fixtures, 1 created
	})

	t.Run("should return an empty slice for a user with no labels", func(t *testing.T) {
		otherUser := &user.User{ID: 999} // Use a user ID that doesn't exist in fixtures
		labels, err := ls.GetAll(s, otherUser)
		assert.NoError(t, err)
		assert.NotNil(t, labels)
		assert.Len(t, labels, 0)
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
