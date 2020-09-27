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

package models

import (
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"testing"
	"xorm.io/xorm/schemas"
)

func TestSavedFilter_getListIDFromFilter(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		assert.Equal(t, int64(-2), getListIDFromSavedFilterID(1))
	})
	t.Run("invalid", func(t *testing.T) {
		assert.Equal(t, int64(0), getListIDFromSavedFilterID(-1))
	})
}

func TestSavedFilter_getFilterIDFromListID(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		assert.Equal(t, int64(1), getSavedFilterIDFromListID(-2))
	})
	t.Run("invalid", func(t *testing.T) {
		assert.Equal(t, int64(0), getSavedFilterIDFromListID(2))
	})
}

func TestSavedFilter_Create(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	sf := &SavedFilter{
		Title:       "test",
		Description: "Lorem Ipsum dolor sit amet",
		Filters:     &TaskCollection{}, // Empty filter
	}

	u := &user.User{ID: 1}
	err := sf.Create(u)
	assert.NoError(t, err)
	assert.Equal(t, u.ID, sf.OwnerID)
	vals := map[string]interface{}{
		"title":       "'test'",
		"description": "'Lorem Ipsum dolor sit amet'",
		"filters":     "'{\"sort_by\":null,\"order_by\":null,\"filter_by\":null,\"filter_value\":null,\"filter_comparator\":null,\"filter_concat\":\"\",\"filter_include_nulls\":false}'",
		"owner_id":    1,
	}
	// Postgres can't compare json values directly, see https://dba.stackexchange.com/a/106290/210721
	if x.Dialect().URI().DBType == schemas.POSTGRES {
		vals["filters::jsonb"] = vals["filters"].(string) + "::jsonb"
		delete(vals, "filters")
	}
	db.AssertExists(t, "saved_filters", vals, true)
}

func TestSavedFilter_ReadOne(t *testing.T) {
	user1 := &user.User{ID: 1}
	db.LoadAndAssertFixtures(t)
	sf := &SavedFilter{
		ID: 1,
	}
	// canRead pre-populates the struct
	_, _, err := sf.CanRead(user1)
	assert.NoError(t, err)
	err = sf.ReadOne()
	assert.NoError(t, err)
	assert.NotNil(t, sf.Owner)
}

func TestSavedFilter_Update(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	sf := &SavedFilter{
		ID:          1,
		Title:       "NewTitle",
		Description: "", // Explicitly reset the description
		Filters:     &TaskCollection{},
	}
	err := sf.Update()
	assert.NoError(t, err)
	db.AssertExists(t, "saved_filters", map[string]interface{}{
		"id":          1,
		"title":       "NewTitle",
		"description": "",
	}, false)
}

func TestSavedFilter_Delete(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	sf := &SavedFilter{
		ID: 1,
	}
	err := sf.Delete()
	assert.NoError(t, err)
	db.AssertMissing(t, "saved_filters", map[string]interface{}{
		"id": 1,
	})
}

func TestSavedFilter_Rights(t *testing.T) {
	user1 := &user.User{ID: 1}
	user2 := &user.User{ID: 2}
	ls := &LinkSharing{ID: 1}

	t.Run("create", func(t *testing.T) {
		// Should always be true
		db.LoadAndAssertFixtures(t)
		can, err := (&SavedFilter{}).CanCreate(user1)
		assert.NoError(t, err)
		assert.True(t, can)
	})
	t.Run("read", func(t *testing.T) {
		t.Run("owner", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			sf := &SavedFilter{
				ID:    1,
				Title: "Lorem",
			}
			can, max, err := sf.CanRead(user1)
			assert.NoError(t, err)
			assert.Equal(t, int(RightAdmin), max)
			assert.True(t, can)
		})
		t.Run("not owner", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			sf := &SavedFilter{
				ID:    1,
				Title: "Lorem",
			}
			can, _, err := sf.CanRead(user2)
			assert.NoError(t, err)
			assert.False(t, can)
		})
		t.Run("nonexisting", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			sf := &SavedFilter{
				ID:    9999,
				Title: "Lorem",
			}
			can, _, err := sf.CanRead(user1)
			assert.Error(t, err)
			assert.True(t, IsErrSavedFilterDoesNotExist(err))
			assert.False(t, can)
		})
		t.Run("link share", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			sf := &SavedFilter{
				ID:    1,
				Title: "Lorem",
			}
			can, _, err := sf.CanRead(ls)
			assert.Error(t, err)
			assert.True(t, IsErrSavedFilterNotAvailableForLinkShare(err))
			assert.False(t, can)
		})
	})
	t.Run("update", func(t *testing.T) {
		t.Run("owner", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			sf := &SavedFilter{
				ID:    1,
				Title: "Lorem",
			}
			can, err := sf.CanUpdate(user1)
			assert.NoError(t, err)
			assert.True(t, can)
		})
		t.Run("not owner", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			sf := &SavedFilter{
				ID:    1,
				Title: "Lorem",
			}
			can, err := sf.CanUpdate(user2)
			assert.NoError(t, err)
			assert.False(t, can)
		})
		t.Run("nonexisting", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			sf := &SavedFilter{
				ID:    9999,
				Title: "Lorem",
			}
			can, err := sf.CanUpdate(user1)
			assert.Error(t, err)
			assert.True(t, IsErrSavedFilterDoesNotExist(err))
			assert.False(t, can)
		})
		t.Run("link share", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			sf := &SavedFilter{
				ID:    1,
				Title: "Lorem",
			}
			can, err := sf.CanUpdate(ls)
			assert.Error(t, err)
			assert.True(t, IsErrSavedFilterNotAvailableForLinkShare(err))
			assert.False(t, can)
		})
	})
	t.Run("delete", func(t *testing.T) {
		t.Run("owner", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			sf := &SavedFilter{
				ID: 1,
			}
			can, err := sf.CanDelete(user1)
			assert.NoError(t, err)
			assert.True(t, can)
		})
		t.Run("not owner", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			sf := &SavedFilter{
				ID: 1,
			}
			can, err := sf.CanDelete(user2)
			assert.NoError(t, err)
			assert.False(t, can)
		})
		t.Run("nonexisting", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			sf := &SavedFilter{
				ID:    9999,
				Title: "Lorem",
			}
			can, err := sf.CanDelete(user1)
			assert.Error(t, err)
			assert.True(t, IsErrSavedFilterDoesNotExist(err))
			assert.False(t, can)
		})
		t.Run("link share", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			sf := &SavedFilter{
				ID:    1,
				Title: "Lorem",
			}
			can, err := sf.CanDelete(ls)
			assert.Error(t, err)
			assert.True(t, IsErrSavedFilterNotAvailableForLinkShare(err))
			assert.False(t, can)
		})
	})
}
