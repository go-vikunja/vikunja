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
	"xorm.io/builder"
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
		t.Run("pseudo parent project id is rejected", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
			project := Project{
				Title:           "pseudo parent create",
				ParentProjectID: Ptr(int64(-1)),
			}
			err := project.Create(s, usr)
			require.Error(t, err)
			assert.True(t, IsErrProjectCannotBelongToAPseudoParentProject(err))
			db.AssertMissing(t, "projects", map[string]interface{}{"title": "pseudo parent create"})
		})
		t.Run("bot project owned by bot owner", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			bot := &user.User{ID: 23, Username: "bot-owner-a-assistant", BotOwnerID: 21}
			project := Project{Title: "bot-created"}
			err := project.Create(s, bot)
			require.NoError(t, err)
			require.NoError(t, s.Commit())

			// The bot's owner should own the project.
			db.AssertExists(t, "projects", map[string]interface{}{
				"id":       project.ID,
				"owner_id": 21,
			}, false)
			// The bot should retain access via a share.
			db.AssertExists(t, "users_projects", map[string]interface{}{
				"user_id":    23,
				"project_id": project.ID,
				"permission": PermissionAdmin,
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
				ParentProjectID: Ptr(int64(999999)),
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
		t.Run("tiny position triggers recalculation which includes the project itself", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			project := Project{
				ID:       1,
				Title:    "Test1",
				Position: 0.00001,
			}
			err := project.Update(s, usr)
			require.NoError(t, err)
			require.NoError(t, s.Commit())

			// Had the lowest position, so it lands in the first slot of the new range.
			assert.Greater(t, project.Position, 0.1)

			verifySession := db.NewSession()
			defer verifySession.Close()
			rootProjects := []*Project{}
			require.NoError(t, verifySession.
				Where(builder.Or(builder.Eq{"parent_project_id": 0}, builder.IsNull{"parent_project_id"})).
				OrderBy("position asc, id asc").
				Find(&rootProjects))

			require.NotEmpty(t, rootProjects)
			assert.Equal(t, int64(1), rootProjects[0].ID)
			for _, p := range rootProjects {
				assert.Greater(t, p.Position, 0.1)
			}
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
					ParentProjectID: Ptr(int64(7)), // from 6
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
					"parent_project_id": project.parentID(),
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
					ParentProjectID: Ptr(int64(2)), // from 1
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
					ParentProjectID: Ptr(int64(-1)),
				}
				err := project.Update(s, usr)
				require.Error(t, err)
				assert.True(t, IsErrProjectCannotBelongToAPseudoParentProject(err))
			})
			t.Run("attacker with direct Write on victim project cannot reparent it (GHSA-2vq4-854f-5c72)", func(t *testing.T) {
				// User 1 has direct Write on project 10 (owner=6) via
				// users_projects id=4 and owns root project 1. Pre-fix, a
				// reparent of 10 under 1 passed the CanWrite check and the
				// CTE then cascaded Admin on 10 via ownership of the new
				// parent.
				db.LoadAndAssertFixtures(t)
				s := db.NewSession()
				defer s.Close()
				project := Project{
					ID:              10,
					Title:           "Test10",
					ParentProjectID: Ptr(int64(1)), // attacker-owned root
				}
				err := project.Update(s, usr)
				require.Error(t, err)
				assert.True(t, IsErrGenericForbidden(err))
			})
			t.Run("attacker with inherited Write cannot reparent child to attacker root (GHSA-2vq4-854f-5c72)", func(t *testing.T) {
				// User 1 has Write on project 10 and therefore inherits Write on its
				// child (project 43, added in the fixture above) via the CTE. User 1
				// owns project 1. Reparenting 43 under 1 must be rejected.
				db.LoadAndAssertFixtures(t)
				s := db.NewSession()
				defer s.Close()
				project := Project{
					ID:              43,
					Title:           "Reparent Escalation Test Child",
					ParentProjectID: Ptr(int64(1)),
				}
				err := project.Update(s, usr)
				require.Error(t, err)
				assert.True(t, IsErrGenericForbidden(err))
			})
			t.Run("non-reparent update with Write still permitted (regression)", func(t *testing.T) {
				// User 1 has Write (not Admin) on project 43 via project 10;
				// a rename with parent unchanged must not trip the Admin gate.
				db.LoadAndAssertFixtures(t)
				s := db.NewSession()
				defer s.Close()
				project := Project{
					ID:              43,
					Title:           "Reparent Escalation Test Child renamed",
					ParentProjectID: Ptr(int64(10)), // unchanged — no reparent intent
				}
				err := project.Update(s, usr)
				require.NoError(t, err)
			})
			t.Run("partial update with omitted parent keeps the parent (regression)", func(t *testing.T) {
				// A partial update (ParentProjectID nil) must not silently
				// detach project 43 from its parent 10.
				db.LoadAndAssertFixtures(t)
				s := db.NewSession()
				defer s.Close()
				project := Project{
					ID:    43,
					Title: "Reparent Escalation Test Child renamed again",
				}
				err := project.Update(s, usr)
				require.NoError(t, err)
				stored, err := GetProjectSimpleByID(s, 43)
				require.NoError(t, err)
				assert.Equal(t, int64(10), stored.parentID())
			})
			t.Run("attacker with Write cannot detach a child to root via parent_project_id=0 (GHSA-44v6-7fxq-vgf4)", func(t *testing.T) {
				// User 1 has Write (not Admin) on project 43 via project 10.
				// Sending an explicit parent_project_id=0 detaches the child to
				// the top level, which severs the inherited-permission chain and
				// must require Admin on the moved project.
				db.LoadAndAssertFixtures(t)
				s := db.NewSession()
				defer s.Close()
				project := Project{
					ID:              43,
					Title:           "Reparent Escalation Test Child",
					ParentProjectID: Ptr(int64(0)), // explicit detach-to-root
				}
				err := project.Update(s, usr)
				require.Error(t, err)
				assert.True(t, IsErrGenericForbidden(err))
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
		t.Run("plain edit of parent keeps individually archived child archived", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// 40 is archived individually under unarchived 3
			project := Project{ID: 3, Title: "Test3 renamed", IsArchived: false}
			err := project.Update(s, &user.User{ID: 1})
			require.NoError(t, err)
			require.NoError(t, s.Commit())

			db.AssertExists(t, "projects", map[string]interface{}{"id": 40, "is_archived": true}, false)
		})
		t.Run("move under archived parent is rejected", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			project := Project{ID: 1, ParentProjectID: Ptr(int64(22))}
			_, err := project.CanUpdate(s, &user.User{ID: 1})
			require.Error(t, err)
			assert.True(t, IsErrProjectIsArchived(err))

			err = UpdateProject(s, &Project{ID: 1, Title: "Test1", ParentProjectID: Ptr(int64(22))}, &user.User{ID: 1}, false)
			var e ErrParentProjectIsArchived
			require.ErrorAs(t, err, &e)
			assert.Equal(t, int64(22), e.ParentProjectID)
		})
		t.Run("move archived project to root keeps it archived", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			project := Project{ID: 40, Title: "Test40", IsArchived: true, ParentProjectID: Ptr(int64(0))}
			err := project.Update(s, &user.User{ID: 1})
			require.NoError(t, err)
			require.NoError(t, s.Commit())

			db.AssertExists(t, "projects", map[string]interface{}{"id": 40, "parent_project_id": 0, "is_archived": true}, false)
		})
		t.Run("unarchive child under archived parent is rejected", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			project := Project{ID: 21, IsArchived: false}
			err := project.Update(s, &user.User{ID: 1})
			var e ErrParentProjectIsArchived
			require.ErrorAs(t, err, &e)
			assert.Equal(t, int64(22), e.ParentProjectID)
		})
		t.Run("unarchive orphaned project", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			project := Project{ID: 39, Title: "Orphaned project with deleted parent", IsArchived: false}
			err := project.Update(s, &user.User{ID: 1})
			require.NoError(t, err)
			require.NoError(t, s.Commit())

			db.AssertExists(t, "projects", map[string]interface{}{"id": 39, "is_archived": false}, false)
		})
		t.Run("unarchive orphaned project with echoed parent", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			// The frontend sends the whole project back, dangling parent included.
			project := Project{ID: 39, Title: "Orphaned project with deleted parent", IsArchived: false, ParentProjectID: Ptr(int64(999999))}
			err := project.Update(s, &user.User{ID: 1})
			require.NoError(t, err)
			require.NoError(t, s.Commit())

			db.AssertExists(t, "projects", map[string]interface{}{"id": 39, "is_archived": false}, false)
		})
		t.Run("move under nonexistent parent", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			project := Project{ID: 1, Title: "Test1", ParentProjectID: Ptr(int64(999999))}
			err := project.Update(s, &user.User{ID: 1})
			require.Error(t, err)
			assert.True(t, IsErrProjectDoesNotExist(err))
		})
		t.Run("unarchive child and detach to root in one request", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			project := Project{ID: 21, Title: "Test21", IsArchived: false, ParentProjectID: Ptr(int64(0))}
			err := project.Update(s, &user.User{ID: 1})
			require.NoError(t, err)
			require.NoError(t, s.Commit())

			db.AssertExists(t, "projects", map[string]interface{}{"id": 21, "is_archived": false, "parent_project_id": 0}, false)
		})
		t.Run("unarchive child and move under unarchived parent", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			project := Project{ID: 21, Title: "Test21", IsArchived: false, ParentProjectID: Ptr(int64(1))}
			err := project.Update(s, &user.User{ID: 1})
			require.NoError(t, err)
			require.NoError(t, s.Commit())

			db.AssertExists(t, "projects", map[string]interface{}{"id": 21, "is_archived": false, "parent_project_id": 1}, false)
		})
		t.Run("unarchive child under unarchived parent", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			project := Project{ID: 40, IsArchived: false}
			can, err := project.CanUpdate(s, &user.User{ID: 1})
			require.NoError(t, err)
			assert.True(t, can)
		})
		t.Run("unarchive root unarchives subtree", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			project := Project{ID: 22, Title: "Test22", IsArchived: false}
			err := project.Update(s, &user.User{ID: 1})
			require.NoError(t, err)
			require.NoError(t, s.Commit())

			db.AssertExists(t, "projects", map[string]interface{}{"id": 22, "is_archived": false}, false)
			db.AssertExists(t, "projects", map[string]interface{}{"id": 21, "is_archived": false}, false)
		})
	})
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
		db.AssertMissing(t, "projects", map[string]interface{}{
			"id": 1,
		})
		// AssertMissing queries raw, so this also covers the soft-deleted task 51
		db.AssertMissing(t, "tasks", map[string]interface{}{
			"project_id": 1,
		})
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
	t.Run("deletes archived parent and its child atomically", func(t *testing.T) {
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

		db.AssertMissing(t, "projects", map[string]interface{}{"id": 22})
		db.AssertMissing(t, "projects", map[string]interface{}{"id": 21})
	})
	t.Run("deletes deeply nested child projects recursively", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Project hierarchy: 27 -> 12 -> 25 -> 26 (all owned by user 6)
		project := Project{ID: 27}
		err := project.Delete(s, &user.User{ID: 6})
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertMissing(t, "projects", map[string]interface{}{"id": 27})
		db.AssertMissing(t, "projects", map[string]interface{}{"id": 12})
		db.AssertMissing(t, "projects", map[string]interface{}{"id": 25})
		db.AssertMissing(t, "projects", map[string]interface{}{"id": 26})
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
		// +1 for the reparent-escalation fixture child (project 43, owner=6).
		assert.Len(t, projects, 28)
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
		// +1 for the reparent-escalation fixture child (project 43) that
		// user 1 inherits Write on via project 10.
		assert.Len(t, ls, 28)
		assert.Equal(t, int64(3), ls[0].ID) // Project 3 has a position of 1 and should be sorted first
		assert.Equal(t, int64(1), ls[1].ID)
		assert.Equal(t, int64(6), ls[2].ID)
		assert.Equal(t, int64(-1), ls[26].ID)
		assert.Equal(t, int64(-2), ls[27].ID)
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

		projectIDs := make([]int64, len(ls))
		for i, p := range ls {
			projectIDs[i] = p.ID
		}
		if db.ParadeDBAvailable() {
			// ParadeDB fuzzy(1, prefix=true) matches far beyond the literal term, so
			// the result set varies by backend — the accessible set does not.
			accessible := []int64{-1, 1, 3, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 19, 21, 22, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 39, 40, 43}
			assert.Subset(t, accessible, projectIDs, "search returned a project outside user 1's accessible set")
		} else {
			// Project 10 (the search target) and the favorites pseudo project -1.
			// Its child 43 does not match and is no longer pulled in as a descendant.
			require.Len(t, ls, 2)
		}
		assert.Contains(t, projectIDs, int64(10))
		assert.NotContains(t, projectIDs, int64(43))
		assert.Contains(t, projectIDs, int64(-1))
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
	t.Run("archived propagation aggregation", func(t *testing.T) {
		// Regression test for #2589. Child projects of archived parents carry
		// is_archived themselves; getAllProjectsForUser must return that flag and
		// hide them when getArchived=false.
		// The CTE must use dialect-agnostic SQL — no CAST(... AS int), which MySQL 8 rejects.

		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		findByID := func(ps []*Project, id int64) *Project {
			for _, p := range ps {
				if p.ID == id {
					return p
				}
			}
			return nil
		}

		// getArchived=true: project 21 (child of archived 22) must appear and carry is_archived=true.
		withArchived, _, err := getAllProjectsForUser(s, 1, &projectOptions{getArchived: true})
		require.NoError(t, err)

		parent := findByID(withArchived, 22)
		require.NotNil(t, parent, "archived parent project 22 must be returned when getArchived=true")
		assert.True(t, parent.IsArchived, "project 22 is archived in fixtures")

		child := findByID(withArchived, 21)
		require.NotNil(t, child, "child project 21 must be returned when getArchived=true")
		assert.True(t, child.IsArchived, "project 21 is archived in fixtures (through its parent 22)")

		// getArchived=false: both rows must be filtered out.
		withoutArchived, _, err := getAllProjectsForUser(s, 1, &projectOptions{getArchived: false})
		require.NoError(t, err)

		assert.Nil(t, findByID(withoutArchived, 22),
			"archived project 22 must be filtered when getArchived=false")
		assert.Nil(t, findByID(withoutArchived, 21),
			"child of archived project (21) must be filtered when getArchived=false")
		assert.Nil(t, findByID(withoutArchived, 40),
			"archived child (40) of unarchived parent must be filtered when getArchived=false")

		// Sanity: a non-archived project owned by user 1 is still present in the filtered list.
		assert.NotNil(t, findByID(withoutArchived, 1),
			"non-archived project 1 must still be present when getArchived=false")
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
	t.Run("child archived through parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		u := &user.User{ID: 1}
		l := &Project{ID: 21}
		_, _, err := l.CanRead(s, u)
		require.NoError(t, err)
		require.NoError(t, l.ReadOne(s, u))
		assert.True(t, l.IsArchived)

		projects, _, err := getAllProjectsForUser(s, u.ID, &projectOptions{getArchived: true})
		require.NoError(t, err)
		var fromList *Project
		for _, p := range projects {
			if p.ID == 21 {
				fromList = p
				break
			}
		}
		require.NotNil(t, fromList)
		assert.True(t, fromList.IsArchived)
	})
}

func TestCheckIsArchived(t *testing.T) {
	t.Run("child project archived individually with non-archived parent", func(t *testing.T) {
		// Project 40 is archived individually (is_archived=true) but its parent
		// (project 1) is not archived. CheckIsArchived must still return an error.
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		p := &Project{ID: 40, ParentProjectID: Ptr(int64(3))}
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
	t.Run("child project archived through parent", func(t *testing.T) {
		// Project 21 was archived through its parent (22), its own row is flagged.
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		p := &Project{ID: 21}
		err := p.CheckIsArchived(s)
		require.Error(t, err)
		assert.True(t, IsErrProjectIsArchived(err))
	})
	t.Run("new project under archived parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		p := &Project{ParentProjectID: Ptr(int64(22))}
		err := p.CheckIsArchived(s)
		require.Error(t, err)
		assert.True(t, IsErrProjectIsArchived(err))
	})
	t.Run("new project under unarchived parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		p := &Project{ParentProjectID: Ptr(int64(1))}
		require.NoError(t, p.CheckIsArchived(s))
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
