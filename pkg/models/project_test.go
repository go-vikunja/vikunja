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
			defer s.Close()
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
			defer s.Close()
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
			defer s.Close()
			project := Project{
				Title:           "test",
				Description:     "Lorem Ipsum",
				ParentProjectID: 999999,
			}
			err := project.Create(s, usr)
			require.Error(t, err)
			assert.True(t, IsErrProjectDoesNotExist(err))
		})
		t.Run("nonexistent owner", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
			usr := &user.User{ID: 9482385}
			project := Project{
				Title:       "test",
				Description: "Lorem Ipsum",
			}
			err := project.Create(s, usr)
			require.Error(t, err)
			assert.True(t, user.IsErrUserDoesNotExist(err))
		})
		t.Run("existing identifier", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
			project := Project{
				Title:       "test",
				Description: "Lorem Ipsum",
				Identifier:  "test1",
			}
			err := project.Create(s, usr)
			require.Error(t, err)
			assert.True(t, IsErrProjectIdentifierIsNotUnique(err))
		})
		t.Run("non ascii characters", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
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
			defer s.Close()
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
			defer s.Close()
			project := Project{
				ID:    99999999,
				Title: "test",
			}
			err := project.Update(s, usr)
			require.Error(t, err)
			assert.True(t, IsErrProjectDoesNotExist(err))
		})
		t.Run("existing identifier", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
			project := Project{
				Title:       "test",
				Description: "Lorem Ipsum",
				Identifier:  "test1",
			}
			err := project.Create(s, usr)
			require.Error(t, err)
			assert.True(t, IsErrProjectIdentifierIsNotUnique(err))
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
				defer s.Close()
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
				defer s.Close()
				project := Project{
					ID:              1,
					Title:           "Test1",
					Description:     "Lorem Ipsum",
					ParentProjectID: 2, // from 1
				}
				can, _ := project.CanUpdate(s, usr)
				assert.False(t, can) // project is not writeable by us
			})
			t.Run("pseudo project", func(t *testing.T) {
				usr := &user.User{
					ID:       6,
					Username: "user6",
					Email:    "user6@example.com",
				}

				db.LoadAndAssertFixtures(t)
				s := db.NewSession()
				defer s.Close()
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
			defer s.Close()
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
			defer s.Close()
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

func assertSoftDeleted(t *testing.T, projectID int64) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()

	// Use Unscoped to bypass soft-delete filter
	p := &Project{}
	exists, err := s.Unscoped().Where("id = ?", projectID).Get(p)
	require.NoError(t, err)
	require.True(t, exists, "Project %d should still exist in db after soft-delete", projectID)
	assert.NotNil(t, p.DeletedAt, "Project %d should have deleted_at set", projectID)
}

func assertNotSoftDeleted(t *testing.T, projectID int64) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()

	p := &Project{}
	exists, err := s.Unscoped().Where("id = ?", projectID).Get(p)
	require.NoError(t, err)
	require.True(t, exists, "Project %d should exist in db", projectID)
	assert.Nil(t, p.DeletedAt, "Project %d should not have deleted_at set", projectID)
}

func TestProject_Delete(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		project := Project{
			ID: 1,
		}
		err := project.Delete(s, &user.User{ID: 1})
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		// With soft-delete, project row still exists but has deleted_at set
		assertSoftDeleted(t, 1)
		// Tasks should still exist (not permanently deleted)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id": 1,
		}, false)
	})
	t.Run("with background", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		s := db.NewSession()
		defer s.Close()
		project := Project{
			ID: 35,
		}
		err := project.Delete(s, &user.User{ID: 6})
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		// Project is soft-deleted, background file still exists
		assertSoftDeleted(t, 35)
		db.AssertExists(t, "files", map[string]interface{}{
			"id": 1,
		}, false)
	})
	t.Run("default project of the same user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
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
		defer s.Close()
		project := Project{
			ID: 4,
		}
		err := project.Delete(s, &user.User{ID: 2})
		require.Error(t, err)
		assert.True(t, IsErrCannotDeleteDefaultProject(err))
	})
	t.Run("soft-deletes archived parent and its child", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 22 is archived (is_archived=1), owned by user 1
		// Project 21 is a child of 22 (parent_project_id=22)
		project := Project{ID: 22}
		err := project.Delete(s, &user.User{ID: 1})
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		assertSoftDeleted(t, 22)
		assertSoftDeleted(t, 21)
	})
	t.Run("soft-deletes deeply nested child projects recursively", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project hierarchy: 27 -> 12 -> 25 -> 26 (all owned by user 6)
		project := Project{ID: 27}
		err := project.Delete(s, &user.User{ID: 6})
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		assertSoftDeleted(t, 27)
		assertSoftDeleted(t, 12)
		assertSoftDeleted(t, 25)
		assertSoftDeleted(t, 26)
	})
	t.Run("soft-deleted projects are excluded from ReadAll", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 41 is soft-deleted and owned by user 1
		p := &Project{}
		projects, _, _, err := p.ReadAll(s, &user.User{ID: 1}, "", 1, 50)
		require.NoError(t, err)

		projectList := projects.([]*Project)
		for _, proj := range projectList {
			assert.NotEqual(t, int64(41), proj.ID, "Soft-deleted project 41 should not appear in ReadAll")
			assert.NotEqual(t, int64(42), proj.ID, "Soft-deleted project 42 should not appear in ReadAll")
			assert.NotEqual(t, int64(43), proj.ID, "Soft-deleted project 43 should not appear in ReadAll")
		}
	})
	t.Run("soft-deleted projects return no permission", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 41 is soft-deleted and owned by user 1
		// XORM auto-filters soft-deleted projects, so CanRead will error with "Project does not exist"
		p := &Project{ID: 41}
		canRead, _, err := p.CanRead(s, &user.User{ID: 1})
		if err != nil {
			assert.True(t, IsErrProjectDoesNotExist(err), "Expected project not found error for soft-deleted project")
		} else {
			assert.False(t, canRead, "Should not be able to read soft-deleted project")
		}
	})
}

func TestProject_Restore(t *testing.T) {
	t.Run("restore soft-deleted project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 41 is soft-deleted and owned by user 1
		project, err := RestoreProject(s, 41, &user.User{ID: 1})
		require.NoError(t, err)
		require.NotNil(t, project)
		assert.Nil(t, project.DeletedAt)
		err = s.Commit()
		require.NoError(t, err)

		assertNotSoftDeleted(t, 41)
	})
	t.Run("restore soft-deleted parent restores children", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 42 is soft-deleted parent, 43 is soft-deleted child
		_, err := RestoreProject(s, 42, &user.User{ID: 1})
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		assertNotSoftDeleted(t, 42)
		assertNotSoftDeleted(t, 43)
	})
	t.Run("restore non-existent project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := RestoreProject(s, 999, &user.User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrProjectDoesNotExist(err))
	})
	t.Run("restore project without admin access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project 41 is owned by user 1, user 2 has no access
		_, err := RestoreProject(s, 41, &user.User{ID: 2})
		require.Error(t, err)
	})
}

func TestProject_GetDeletedProjects(t *testing.T) {
	t.Run("returns deleted projects for user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 1 owns soft-deleted projects 41, 42, 43
		projects, err := GetDeletedProjects(s, &user.User{ID: 1})
		require.NoError(t, err)
		assert.Len(t, projects, 3)

		ids := make(map[int64]bool)
		for _, p := range projects {
			ids[p.ID] = true
		}
		assert.True(t, ids[41])
		assert.True(t, ids[42])
		assert.True(t, ids[43])
	})
	t.Run("returns empty for user with no deleted projects", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// User 2 owns no soft-deleted projects
		projects, err := GetDeletedProjects(s, &user.User{ID: 2})
		require.NoError(t, err)
		assert.Empty(t, projects)
	})
}

func TestProject_PermanentDelete(t *testing.T) {
	t.Run("permanently deletes project and all related entities", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// First soft-delete the project
		project := Project{ID: 1}
		err := project.Delete(s, &user.User{ID: 1})
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// Now permanently delete it
		s2 := db.NewSession()
		defer s2.Close()

		p := &Project{ID: 1}
		err = p.PermanentDelete(s2, &user.User{ID: 1})
		require.NoError(t, err)
		err = s2.Commit()
		require.NoError(t, err)

		db.AssertMissing(t, "projects", map[string]interface{}{"id": 1})
		db.AssertMissing(t, "tasks", map[string]interface{}{"id": 1})
	})
}

func TestProject_DeleteBackgroundFileIfExists(t *testing.T) {
	t.Run("project with background", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		s := db.NewSession()
		defer s.Close()
		file := &files.File{ID: 1}
		project := Project{
			ID:               1,
			BackgroundFileID: file.ID,
		}
		err := SetProjectBackground(s, project.ID, file, "")
		require.NoError(t, err)
		err = project.DeleteBackgroundFileIfExists(s)
		require.NoError(t, err)
	})
	t.Run("project with invalid background", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		s := db.NewSession()
		defer s.Close()
		file := &files.File{ID: 9999}
		project := Project{
			ID:               1,
			BackgroundFileID: file.ID,
		}
		err := SetProjectBackground(s, project.ID, file, "")
		require.NoError(t, err)
		err = project.DeleteBackgroundFileIfExists(s)
		require.NoError(t, err)
	})
	t.Run("project without background", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		s := db.NewSession()
		defer s.Close()
		project := Project{ID: 1}
		err := project.DeleteBackgroundFileIfExists(s)
		require.NoError(t, err)
	})
}

func TestProject_ReadAll(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		projects, _, err := getAllProjectsForUser(s, 6, &projectOptions{})
		require.NoError(t, err)
		assert.Len(t, projects, 25)
	})
	t.Run("all projects for user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		u := &user.User{ID: 1}
		project := Project{}
		projects3, _, _, err := project.ReadAll(s, u, "", 1, 50)

		require.NoError(t, err)
		assert.Equal(t, reflect.Slice, reflect.TypeOf(projects3).Kind())
		ls := projects3.([]*Project)
		assert.Len(t, ls, 27)
		assert.Equal(t, int64(3), ls[0].ID) // Project 3 has a position of 1 and should be sorted first
		assert.Equal(t, int64(1), ls[1].ID)
		assert.Equal(t, int64(6), ls[2].ID)
		assert.Equal(t, int64(-1), ls[25].ID)
		assert.Equal(t, int64(-2), ls[26].ID)
	})
	t.Run("projects for nonexistent user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		usr := &user.User{ID: 999999}
		project := Project{}
		_, _, _, err := project.ReadAll(s, usr, "", 1, 50)
		require.Error(t, err)
		assert.True(t, user.IsErrUserDoesNotExist(err))
	})
	t.Run("search", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		u := &user.User{ID: 1}
		project := Project{}
		projects3, _, _, err := project.ReadAll(s, u, "TEST10", 1, 50)

		require.NoError(t, err)
		ls := projects3.([]*Project)

		if db.ParadeDBAvailable() {
			// ParadeDB fuzzy(1, prefix=true) on "TEST10" also matches
			// "test1", "test11", "test19", "test30" (edit distance 1), etc.
			require.Len(t, ls, 6)
			projectIDs := make([]int64, len(ls))
			for i, p := range ls {
				projectIDs[i] = p.ID
			}
			assert.Contains(t, projectIDs, int64(10))
			assert.Contains(t, projectIDs, int64(-1))
		} else {
			require.Len(t, ls, 2)
			assert.Equal(t, int64(10), ls[0].ID)
			assert.Equal(t, int64(-1), ls[1].ID)
		}
	})
	t.Run("search returns filters as well", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		u := &user.User{ID: 1}
		project := Project{}
		projects3, _, _, err := project.ReadAll(s, u, "testfilter", 1, 50)

		require.NoError(t, err)
		ls := projects3.([]*Project)
		require.Len(t, ls, 2)
		assert.Equal(t, int64(-1), ls[0].ID)
		assert.Equal(t, int64(-2), ls[1].ID)
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

func TestCheckIsArchived(t *testing.T) {
	t.Run("child project archived individually with non-archived parent", func(t *testing.T) {
		// Project 40 is archived individually (is_archived=true) but its parent
		// (project 1) is not archived. CheckIsArchived must still return an error.
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		p := &Project{ID: 40, ParentProjectID: 3}
		err := p.CheckIsArchived(s)
		require.Error(t, err)
		assert.True(t, IsErrProjectIsArchived(err))
	})
	t.Run("root project archived", func(t *testing.T) {
		// Project 22 is archived individually with no parent.
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		p := &Project{ID: 22}
		err := p.CheckIsArchived(s)
		require.Error(t, err)
		assert.True(t, IsErrProjectIsArchived(err))
	})
	t.Run("child project inherits archived from parent", func(t *testing.T) {
		// Project 21's parent (project 22) is archived.
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		p := &Project{ID: 21, ParentProjectID: 22}
		err := p.CheckIsArchived(s)
		require.Error(t, err)
		assert.True(t, IsErrProjectIsArchived(err))
	})
	t.Run("non-archived project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		p := &Project{ID: 1}
		err := p.CheckIsArchived(s)
		require.NoError(t, err)
	})
}
