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

package models

import (
	"reflect"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProject_CreateOrUpdate(t *testing.T) {
	usr := &user.User{
		ID:       1,
		Username: "user1",
		Email:    "user1@example.com",
	}

	t.Run("create", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			project := Project{
				Title:       "test",
				Description: "Lorem Ipsum",
			}
			err := project.Create(s, usr)
			require.NoError(t, err)
			err = s.Commit()
			require.NoError(t, err)
			db.AssertExists(t, "projects", map[string]interface{}{
				"id":                project.ID,
				"title":             project.Title,
				"description":       project.Description,
				"parent_project_id": 0,
			}, false)
			db.AssertExists(t, "project_views", map[string]interface{}{
				"project_id": project.ID,
				"view_kind":  ProjectViewKindList,
			}, false)
			db.AssertExists(t, "project_views", map[string]interface{}{
				"project_id": project.ID,
				"view_kind":  ProjectViewKindGantt,
			}, false)
			db.AssertExists(t, "project_views", map[string]interface{}{
				"project_id": project.ID,
				"view_kind":  ProjectViewKindTable,
			}, false)
			db.AssertExists(t, "project_views", map[string]interface{}{
				"project_id":                project.ID,
				"view_kind":                 ProjectViewKindKanban,
				"bucket_configuration_mode": BucketConfigurationModeManual,
			}, false)

			kanbanView := &ProjectView{}
			_, err = s.Where("project_id = ? AND view_kind = ?", project.ID, ProjectViewKindKanban).Get(kanbanView)
			require.NoError(t, err)
			db.AssertExists(t, "buckets", map[string]interface{}{
				"project_view_id": kanbanView.ID,
			}, false)
		})
		t.Run("kanban view creates To-Do, doing, done buckets", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			project := Project{
				Title:       "test kanban buckets",
				Description: "Lorem Ipsum",
			}
			err := project.Create(s, usr)
			require.NoError(t, err)
			err = s.Commit()
			require.NoError(t, err)

			// Get the kanban view
			kanbanView := &ProjectView{}
			_, err = s.Where("project_id = ? AND view_kind = ?", project.ID, ProjectViewKindKanban).Get(kanbanView)
			require.NoError(t, err)

			// Check that three buckets were created
			var bucketCount int64
			bucketCount, err = s.Where("project_view_id = ?", kanbanView.ID).Count(&Bucket{})
			require.NoError(t, err)
			assert.Equal(t, int64(3), bucketCount, "Should have created three buckets")

			// Check that the buckets are named correctly
			var buckets []*Bucket
			err = s.Where("project_view_id = ?", kanbanView.ID).OrderBy("position ASC").Find(&buckets)
			require.NoError(t, err)
			require.Len(t, buckets, 3, "Should have three buckets")
			assert.Equal(t, "To-Do", buckets[0].Title)
			assert.Equal(t, "Doing", buckets[1].Title)
			assert.Equal(t, "Done", buckets[2].Title)

			// Check that Backlog is the default bucket
			assert.Equal(t, buckets[0].ID, kanbanView.DefaultBucketID, "To-Do should be the default bucket")

			// Check that Done is the done bucket
			assert.Equal(t, buckets[2].ID, kanbanView.DoneBucketID, "Done should be the done bucket")
		})
		t.Run("nonexistent parent", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			project := Project{
				Title:           "test",
				Description:     "Lorem Ipsum",
				ParentProjectID: 999999,
			}
			err := project.Create(s, usr)
			require.Error(t, err)
			assert.True(t, IsErrProjectDoesNotExist(err))
			_ = s.Close()
		})
		t.Run("nonexistent owner", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			usr := &user.User{ID: 9482385}
			project := Project{
				Title:       "test",
				Description: "Lorem Ipsum",
			}
			err := project.Create(s, usr)
			require.Error(t, err)
			assert.True(t, user.IsErrUserDoesNotExist(err))
			_ = s.Close()
		})
		t.Run("existing identifier", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			project := Project{
				Title:       "test",
				Description: "Lorem Ipsum",
				Identifier:  "test1",
			}
			err := project.Create(s, usr)
			require.Error(t, err)
			assert.True(t, IsErrProjectIdentifierIsNotUnique(err))
			_ = s.Close()
		})
		t.Run("non ascii characters", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			project := Project{
				Title:       "приффки фсем",
				Description: "Lorem Ipsum",
			}
			err := project.Create(s, usr)
			require.NoError(t, err)
			err = s.Commit()
			require.NoError(t, err)
			db.AssertExists(t, "projects", map[string]interface{}{
				"id":          project.ID,
				"title":       project.Title,
				"description": project.Description,
			}, false)
		})
	})

	t.Run("update", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			project := Project{
				ID:          1,
				Title:       "test",
				Description: "Lorem Ipsum",
			}
			project.Description = "Lorem Ipsum dolor sit amet."
			err := project.Update(s, usr)
			require.NoError(t, err)
			err = s.Commit()
			require.NoError(t, err)
			db.AssertExists(t, "projects", map[string]interface{}{
				"id":          project.ID,
				"title":       project.Title,
				"description": project.Description,
			}, false)
		})
		t.Run("nonexistent", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			project := Project{
				ID:    99999999,
				Title: "test",
			}
			err := project.Update(s, usr)
			require.Error(t, err)
			assert.True(t, IsErrProjectDoesNotExist(err))
			_ = s.Close()

		})
		t.Run("existing identifier", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			project := Project{
				Title:       "test",
				Description: "Lorem Ipsum",
				Identifier:  "test1",
			}
			err := project.Create(s, usr)
			require.Error(t, err)
			assert.True(t, IsErrProjectIdentifierIsNotUnique(err))
			_ = s.Close()
		})
		t.Run("change parent project", func(t *testing.T) {
			t.Run("own", func(t *testing.T) {
				usr := &user.User{
					ID:       6,
					Username: "user6",
					Email:    "user6@example.com",
				}

				db.LoadAndAssertFixtures(t)
				s := db.NewSession()
				project := Project{
					ID:              6,
					Title:           "Test6",
					Description:     "Lorem Ipsum",
					ParentProjectID: 7, // from 6
				}
				can, err := project.CanUpdate(s, usr)
				require.NoError(t, err)
				assert.True(t, can)
				err = project.Update(s, usr)
				require.NoError(t, err)
				err = s.Commit()
				require.NoError(t, err)
				db.AssertExists(t, "projects", map[string]interface{}{
					"id":                project.ID,
					"title":             project.Title,
					"description":       project.Description,
					"parent_project_id": project.ParentProjectID,
				}, false)
			})
			t.Run("others", func(t *testing.T) {
				db.LoadAndAssertFixtures(t)
				s := db.NewSession()
				project := Project{
					ID:              1,
					Title:           "Test1",
					Description:     "Lorem Ipsum",
					ParentProjectID: 2, // from 1
				}
				can, _ := project.CanUpdate(s, usr)
				assert.False(t, can) // project is not writeable by us
				_ = s.Close()
			})
			t.Run("pseudo project", func(t *testing.T) {
				usr := &user.User{
					ID:       6,
					Username: "user6",
					Email:    "user6@example.com",
				}

				db.LoadAndAssertFixtures(t)
				s := db.NewSession()
				project := Project{
					ID:              6,
					Title:           "Test6",
					Description:     "Lorem Ipsum",
					ParentProjectID: -1,
				}
				err := project.Update(s, usr)
				require.Error(t, err)
				assert.True(t, IsErrProjectCannotBelongToAPseudoParentProject(err))
			})
		})
		t.Run("archive default project of the same user", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			project := Project{
				ID:         4,
				IsArchived: true,
			}
			err := project.Update(s, &user.User{ID: 3})
			require.Error(t, err)
			assert.True(t, IsErrCannotArchiveDefaultProject(err))
		})
		t.Run("archive default project of another user", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			project := Project{
				ID:         4,
				IsArchived: true,
			}
			err := project.Update(s, &user.User{ID: 2})
			require.Error(t, err)
			assert.True(t, IsErrCannotArchiveDefaultProject(err))
		})
		t.Run("archive parent archives child", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			actingUser := &user.User{ID: 6}

			projectToArchive := Project{
				ID: 27,
			}

			// We need to load the project first to have its fields populated for the update
			can, err := projectToArchive.CanUpdate(s, actingUser)
			require.NoError(t, err, "Failed to read project 27 before archiving")
			assert.True(t, can)
			projectToArchive.IsArchived = true // Ensure IsArchived is set after reading

			err = projectToArchive.Update(s, actingUser)
			require.NoError(t, err, "Failed to archive project")
			err = s.Commit()
			require.NoError(t, err, "Failed to commit session after archiving project")

			db.AssertExists(t, "projects", map[string]interface{}{
				"id":          27,
				"is_archived": true,
			}, false)
			// Assert child project (ID 12) is also archived
			db.AssertExists(t, "projects", map[string]interface{}{
				"id":          12,
				"is_archived": true,
			}, false)
		})
	})
}

func TestProject_Delete(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		project := Project{
			ID: 1,
		}
		err := project.Delete(s, &user.User{ID: 1})
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		db.AssertMissing(t, "projects", map[string]interface{}{
			"id": 1,
		})
		db.AssertMissing(t, "tasks", map[string]interface{}{
			"id": 1,
		})
	})
	t.Run("with background", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		s := db.NewSession()
		project := Project{
			ID: 35,
		}
		err := project.Delete(s, &user.User{ID: 6})
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		db.AssertMissing(t, "projects", map[string]interface{}{
			"id": 35,
		})
		db.AssertMissing(t, "files", map[string]interface{}{
			"id": 1,
		})
	})
	t.Run("default project of the same user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		project := Project{
			ID: 4,
		}
		err := project.Delete(s, &user.User{ID: 3})
		require.Error(t, err)
		assert.True(t, IsErrCannotDeleteDefaultProject(err))
	})
	t.Run("default project of a different user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		project := Project{
			ID: 4,
		}
		err := project.Delete(s, &user.User{ID: 2})
		require.Error(t, err)
		assert.True(t, IsErrCannotDeleteDefaultProject(err))
	})
}

func TestProject_DeleteBackgroundFileIfExists(t *testing.T) {
	t.Run("project with background", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		s := db.NewSession()
		file := &files.File{ID: 1}
		project := Project{
			ID:               1,
			BackgroundFileID: file.ID,
		}
		err := SetProjectBackground(s, project.ID, file, "")
		require.NoError(t, err)
		err = project.DeleteBackgroundFileIfExists()
		require.NoError(t, err)
	})
	t.Run("project with invalid background", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		s := db.NewSession()
		file := &files.File{ID: 9999}
		project := Project{
			ID:               1,
			BackgroundFileID: file.ID,
		}
		err := SetProjectBackground(s, project.ID, file, "")
		require.NoError(t, err)
		err = project.DeleteBackgroundFileIfExists()
		require.NoError(t, err)
	})
	t.Run("project without background", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		project := Project{ID: 1}
		err := project.DeleteBackgroundFileIfExists()
		require.NoError(t, err)
	})
}

func TestProject_ReadAll(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		projects, _, err := getAllProjectsForUser(s, 6, &projectOptions{})
		require.NoError(t, err)
		assert.Len(t, projects, 25)
		_ = s.Close()
	})
	t.Run("all projects for user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		u := &user.User{ID: 1}
		project := Project{}
		projects3, _, _, err := project.ReadAll(s, u, "", 1, 50)

		require.NoError(t, err)
		assert.Equal(t, reflect.Slice, reflect.TypeOf(projects3).Kind())
		ls := projects3.([]*Project)
		assert.Len(t, ls, 28)
		assert.Equal(t, int64(3), ls[0].ID) // Project 3 has a position of 1 and should be sorted first
		assert.Equal(t, int64(1), ls[1].ID)
		assert.Equal(t, int64(6), ls[2].ID)
		assert.Equal(t, int64(-1), ls[26].ID)
		assert.Equal(t, int64(-2), ls[27].ID)
		_ = s.Close()
	})
	t.Run("projects for nonexistent user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		usr := &user.User{ID: 999999}
		project := Project{}
		_, _, _, err := project.ReadAll(s, usr, "", 1, 50)
		require.Error(t, err)
		assert.True(t, user.IsErrUserDoesNotExist(err))
		_ = s.Close()
	})
	t.Run("search", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		u := &user.User{ID: 1}
		project := Project{}
		projects3, _, _, err := project.ReadAll(s, u, "TEST10", 1, 50)

		require.NoError(t, err)
		ls := projects3.([]*Project)
		require.Len(t, ls, 2)
		assert.Equal(t, int64(10), ls[0].ID)
		assert.Equal(t, int64(-1), ls[1].ID)
		_ = s.Close()
	})
	t.Run("search returns filters as well", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		u := &user.User{ID: 1}
		project := Project{}
		projects3, _, _, err := project.ReadAll(s, u, "testfilter", 1, 50)

		require.NoError(t, err)
		ls := projects3.([]*Project)
		require.Len(t, ls, 2)
		assert.Equal(t, int64(-1), ls[0].ID)
		assert.Equal(t, int64(-2), ls[1].ID)
		_ = s.Close()
	})
}

func TestProject_ReadOne(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		l := &Project{ID: 1}
		can, _, err := l.CanRead(s, u)
		require.NoError(t, err)
		assert.True(t, can)
		err = l.ReadOne(s, u)
		require.NoError(t, err)
		assert.Equal(t, "Test1", l.Title)
	})
	t.Run("with subscription", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 6}
		l := &Project{ID: 12}
		can, _, err := l.CanRead(s, u)
		require.NoError(t, err)
		assert.True(t, can)
		err = l.ReadOne(s, u)
		require.NoError(t, err)
		assert.NotNil(t, l.Subscription)
	})
}
