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

package caldav

// Tests for #3483: saved filters and the favorites project are exposed as CalDAV
// collections, but they aggregate tasks living in real projects.

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/samedi/caldav-go/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

// Tasks 40-46 in the fixtures are the only ones with a caldav uid and the only ones
// due after 2023-01-01; they belong to user 15's projects 36 and 38.
const caldavFilterMatchingTasks = `due_date > '2023-01-01T00:00:00+00:00'`
const caldavFilterMatchingNothing = `done = true && due_date > '2023-01-01T00:00:00+00:00'`
const caldavFilterTaskCount = 7

// Saved filter 1 in the fixtures belongs to user 1.
const foreignFilterProjectID = -2

// Project 20 in the fixtures is owned by user 13, user 1 has no access to it.
const inaccessibleProjectID = 20

var caldavFilterUser = &user.User{ID: 15, Username: "user15"}
var caldavOtherUser = &user.User{ID: 1, Username: "user1"}

// Goes through the model instead of a raw insert, which would skip the default views TaskCollection.ReadAll branches on.
func createSavedFilter(t *testing.T, u *user.User, filter string) int64 {
	t.Helper()

	s := db.NewSession()
	defer s.Close()

	sf := &models.SavedFilter{
		Title:   "caldav test filter",
		Filters: &models.TaskCollection{Filter: filter},
	}
	require.NoError(t, sf.Create(s, u))
	require.NoError(t, s.Commit())

	pseudoID := -sf.ID - 1
	require.Equal(t, sf.ID, models.GetSavedFilterIDFromProjectID(pseudoID))
	return pseudoID
}

func storageFor(u *user.User, projectID int64) *VikunjaCaldavProjectStorage {
	return &VikunjaCaldavProjectStorage{
		project: &models.ProjectWithTasksAndBuckets{Project: models.Project{ID: projectID}},
		user:    u,
	}
}

func filterStorage(pseudoProjectID int64) *VikunjaCaldavProjectStorage {
	return storageFor(caldavFilterUser, pseudoProjectID)
}

func projectPath(projectID int64) string {
	return ProjectBasePath + "/" + strconv.FormatInt(projectID, 10) + "/"
}

func TestSavedFilterCollection_ChildHrefs(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)
	rpath := projectPath(pseudoID)

	resources, err := filterStorage(pseudoID).GetResources(rpath, true)
	require.NoError(t, err)
	require.Len(t, resources, caldavFilterTaskCount+1)

	// resources[0] is the collection itself; caldav-go strips its trailing slash.
	for _, r := range resources[1:] {
		assert.True(t, strings.HasPrefix(r.Path, rpath),
			"href %q must be nested under the collection %q", r.Path, rpath)
	}
}

// Marking a task done through a filter must work; only creating new tasks in a pseudo
// collection stays forbidden (#3483).
func TestPseudoCollection_Writes(t *testing.T) {
	const newTaskContent = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid-caldav-new
DTSTAMP:20230301T073337Z
SUMMARY:New task
END:VTODO
END:VCALENDAR`

	const doneTaskContent = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid-caldav-test
DTSTAMP:20230301T073337Z
SUMMARY:Done through the filter
STATUS:COMPLETED
END:VTODO
END:VCALENDAR`

	// The PUT and DELETE flows in caldav-go both resolve the resource first, which is what
	// puts the full task - including its real project - into the storage.
	resolve := func(t *testing.T, storage *VikunjaCaldavProjectStorage, rpath string) {
		t.Helper()

		storage.task = &models.Task{UID: "uid-caldav-test"}
		_, found, err := storage.GetResource(rpath)
		require.NoError(t, err)
		require.True(t, found)
	}

	t.Run("create through a filter is forbidden", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)
		storage := filterStorage(pseudoID)
		storage.task = &models.Task{UID: "uid-caldav-new"}

		_, err := storage.CreateResource(projectPath(pseudoID)+"uid-caldav-new.ics", newTaskContent)
		require.ErrorIs(t, err, errs.ForbiddenError)
		db.AssertMissing(t, "tasks", map[string]interface{}{"uid": "uid-caldav-new"})
	})

	t.Run("create through favorites is forbidden", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		storage := filterStorage(models.FavoritesPseudoProjectID)
		storage.task = &models.Task{UID: "uid-caldav-new"}

		_, err := storage.CreateResource(projectPath(models.FavoritesPseudoProjectID)+"uid-caldav-new.ics", newTaskContent)
		require.ErrorIs(t, err, errs.ForbiddenError)
		db.AssertMissing(t, "tasks", map[string]interface{}{"uid": "uid-caldav-new"})
	})

	t.Run("update of a filter member succeeds and leaves it in its project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)
		rpath := projectPath(pseudoID) + "uid-caldav-test.ics"
		storage := filterStorage(pseudoID)
		resolve(t, storage, rpath)

		_, err := storage.UpdateResource(rpath, doneTaskContent)
		require.NoError(t, err)
		db.AssertCount(t, "tasks", builder.Eq{
			"id":         40,
			"title":      "Done through the filter",
			"done":       true,
			"project_id": 36,
		}, 1)
	})

	t.Run("delete of a filter member succeeds", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)
		rpath := projectPath(pseudoID) + "uid-caldav-test.ics"
		storage := filterStorage(pseudoID)
		resolve(t, storage, rpath)

		require.NoError(t, storage.DeleteResource(rpath))
		// Deletion only stamps deleted_at, so mere existence of the row proves nothing.
		db.AssertCount(t, "tasks", builder.And(builder.Eq{"id": 40}, builder.IsNull{"deleted_at"}), 0)
	})

	t.Run("update of a task the filter does not contain is not found", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingNothing)
		storage := filterStorage(pseudoID)
		storage.task = &models.Task{ID: 40, UID: "uid-caldav-test", ProjectID: 36}

		_, err := storage.UpdateResource(projectPath(pseudoID)+"uid-caldav-test.ics", doneTaskContent)
		require.ErrorIs(t, err, errs.ResourceNotFoundError)
		db.AssertCount(t, "tasks", builder.Eq{"id": 40, "title": "Title Caldav Test"}, 1)
	})

	t.Run("delete of a task the filter does not contain is not found", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingNothing)
		storage := filterStorage(pseudoID)
		storage.task = &models.Task{ID: 40, UID: "uid-caldav-test", ProjectID: 36}

		err := storage.DeleteResource(projectPath(pseudoID) + "uid-caldav-test.ics")
		require.ErrorIs(t, err, errs.ResourceNotFoundError)
		db.AssertCount(t, "tasks", builder.And(builder.Eq{"id": 40}, builder.IsNull{"deleted_at"}), 1)
	})

	t.Run("a read-only share still refuses updates", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		// Project 3 is owned by user 3 and shared read-only with user 1.
		storage := storageFor(caldavOtherUser, 3)
		storage.task = &models.Task{UID: "uid-caldav-new"}

		_, err := storage.UpdateResource(projectPath(3)+"uid-caldav-new.ics", newTaskContent)
		require.ErrorIs(t, err, errs.ForbiddenError)
		db.AssertMissing(t, "tasks", map[string]interface{}{"uid": "uid-caldav-new"})
	})
}

// canReadCollection used to have to tolerate ErrProjectDoesNotExist because CanRead
// errored out for instance admins on a pseudo id; d5332ac3b fixed that in the model.
func TestCanReadCollection_InstanceAdmin(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	defer license.ResetForTests()

	s := db.NewSession()
	defer s.Close()

	_, err := s.ID(int64(2)).Cols("is_admin").Update(&user.User{IsAdmin: true})
	require.NoError(t, err)

	admin := &user.User{ID: 2, IsAdmin: true}

	t.Run("favorites", func(t *testing.T) {
		can, err := storageFor(admin, models.FavoritesPseudoProjectID).canReadCollection(s)
		require.NoError(t, err)
		assert.True(t, can)
	})

	t.Run("a saved filter of another user", func(t *testing.T) {
		can, err := storageFor(admin, foreignFilterProjectID).canReadCollection(s)
		require.NoError(t, err)
		assert.False(t, can)
	})
}

// The reduced Allow header must keep advertising PUT and DELETE for pseudo collections,
// without turning OPTIONS into an oracle for collections the user cannot read.
func TestProjectHandlerOPTIONS(t *testing.T) {
	options := func(t *testing.T, u *user.User, projectID int64) *httptest.ResponseRecorder {
		t.Helper()

		rec := httptest.NewRecorder()
		c := echo.New().NewContext(httptest.NewRequest(http.MethodOptions, projectPath(projectID), nil), rec)
		c.Set("userBasicAuth", u)
		c.SetPathValues(echo.PathValues{{Name: "project", Value: strconv.FormatInt(projectID, 10)}})
		require.NoError(t, ProjectHandler(c))
		return rec
	}

	t.Run("a saved filter advertises PUT and DELETE", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)
		rec := options(t, caldavFilterUser, pseudoID)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Allow"), "PUT")
		assert.Contains(t, rec.Header().Get("Allow"), "DELETE")
	})

	t.Run("favorites advertise PUT and DELETE", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		rec := options(t, caldavFilterUser, models.FavoritesPseudoProjectID)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Allow"), "PUT")
		assert.Contains(t, rec.Header().Get("Allow"), "DELETE")
	})

	t.Run("a read-only share does not", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		// Project 3 is owned by user 3 and shared read-only with user 1.
		rec := options(t, caldavOtherUser, 3)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Header().Get("Allow"), "PUT")
		assert.NotContains(t, rec.Header().Get("Allow"), "DELETE")
	})

	t.Run("an archived project does not", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		archiveProject(t, 36)

		rec := options(t, caldavFilterUser, 36)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Header().Get("Allow"), "PUT")
		assert.NotContains(t, rec.Header().Get("Allow"), "DELETE")
	})

	t.Run("a saved filter of another user is not found", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		assert.Equal(t, http.StatusNotFound, options(t, caldavFilterUser, foreignFilterProjectID).Code)
	})

	t.Run("a project without access is not found", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		assert.Equal(t, http.StatusNotFound, options(t, caldavOtherUser, inaccessibleProjectID).Code)
	})
}

// ProjectHandler answers OPTIONS with a reduced Allow header whenever this call denies for
// a real project, because caldav-go would otherwise advertise PUT and DELETE unconditionally.
func TestCollectionWritability(t *testing.T) {
	canWrite := func(t *testing.T, u *user.User, projectID int64) bool {
		t.Helper()

		s := db.NewSession()
		defer s.Close()

		project := &models.ProjectWithTasksAndBuckets{Project: models.Project{ID: projectID}}
		can, err := denyArchived(project.CanWrite(s, u))
		require.NoError(t, err)
		return can
	}

	t.Run("a writable project allows writes", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		assert.True(t, canWrite(t, caldavFilterUser, 36))
	})

	t.Run("a read-only share denies writes", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		// Project 3 is owned by user 3 and shared read-only with user 1.
		assert.False(t, canWrite(t, caldavOtherUser, 3))
	})

	t.Run("an archived project denies writes", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		archiveProject(t, 38)

		assert.False(t, canWrite(t, caldavFilterUser, 38))
	})

	t.Run("a saved filter denies writes", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)
		assert.False(t, canWrite(t, caldavFilterUser, pseudoID))
	})
}

// caldav-go maps only its own errs sentinels, so a propagated permission error answers
// 500 while an unknown id answers 404 - which tells the client the collection exists.
func TestGetResources_UnreadableCollectionIsNotFound(t *testing.T) {
	t.Run("a saved filter of another user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		_, err := filterStorage(foreignFilterProjectID).GetResources(projectPath(foreignFilterProjectID), true)
		require.ErrorIs(t, err, errs.ResourceNotFoundError)
	})

	t.Run("a project the user has no access to", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		_, err := storageFor(caldavOtherUser, inaccessibleProjectID).GetResources(projectPath(inaccessibleProjectID), true)
		require.ErrorIs(t, err, errs.ResourceNotFoundError)
	})
}

func TestGetResource_PseudoCollectionMembership(t *testing.T) {
	t.Run("serves a task the filter contains", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)
		storage := filterStorage(pseudoID)
		storage.task = &models.Task{UID: "uid-caldav-test"}

		_, found, err := storage.GetResource(projectPath(pseudoID) + "uid-caldav-test.ics")
		require.NoError(t, err)
		assert.True(t, found)
	})

	t.Run("hides a task the filter does not contain", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingNothing)
		storage := filterStorage(pseudoID)
		storage.task = &models.Task{UID: "uid-caldav-test"}

		_, _, err := storage.GetResource(projectPath(pseudoID) + "uid-caldav-test.ics")
		require.ErrorIs(t, err, errs.ResourceNotFoundError)
	})

	t.Run("still rejects a mismatching real project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		storage := filterStorage(38)
		storage.task = &models.Task{UID: "uid-caldav-test"}

		_, _, err := storage.GetResource("/dav/projects/38/uid-caldav-test.ics")
		require.ErrorIs(t, err, errs.ResourceNotFoundError)
	})

	t.Run("a filter of another user is a miss, not an error", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		storage := filterStorage(foreignFilterProjectID)
		storage.task = &models.Task{UID: "uid-caldav-test"}

		_, _, err := storage.GetResource(projectPath(foreignFilterProjectID) + "uid-caldav-test.ics")
		require.ErrorIs(t, err, errs.ResourceNotFoundError)
	})

	t.Run("a task in an inaccessible project is a miss", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavOtherUser, caldavFilterMatchingTasks)
		storage := storageFor(caldavOtherUser, pseudoID)
		storage.task = &models.Task{UID: "uid-caldav-test"}

		_, _, err := storage.GetResource(projectPath(pseudoID) + "uid-caldav-test.ics")
		require.ErrorIs(t, err, errs.ResourceNotFoundError)
	})
}

func TestGetResourcesByList_PseudoCollection(t *testing.T) {
	t.Run("returns tasks of the filter under the filter path", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)
		href := projectPath(pseudoID) + "uid-caldav-test.ics"

		resources, err := filterStorage(pseudoID).GetResourcesByList([]string{href})
		require.NoError(t, err)
		require.Len(t, resources, 1)
		assert.Equal(t, href, resources[0].Path)
	})

	t.Run("drops tasks the filter does not contain", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingNothing)

		resources, err := filterStorage(pseudoID).GetResourcesByList([]string{
			projectPath(pseudoID) + "uid-caldav-test.ics",
		})
		require.NoError(t, err)
		assert.Empty(t, resources)
	})

	t.Run("drops hrefs of a foreign pseudo collection", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)
		otherPseudoID := pseudoID - 1

		resources, err := filterStorage(pseudoID).GetResourcesByList([]string{
			projectPath(otherPseudoID) + "uid-caldav-test.ics",
		})
		require.NoError(t, err)
		assert.Empty(t, resources)
	})

	t.Run("a filter of another user drops out without aborting the report", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		resources, err := filterStorage(foreignFilterProjectID).GetResourcesByList([]string{
			projectPath(foreignFilterProjectID) + "uid-caldav-test.ics",
		})
		require.NoError(t, err)
		assert.Empty(t, resources)
	})

	t.Run("drops a task in an inaccessible project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavOtherUser, caldavFilterMatchingTasks)

		resources, err := storageFor(caldavOtherUser, pseudoID).GetResourcesByList([]string{
			projectPath(pseudoID) + "uid-caldav-test.ics",
		})
		require.NoError(t, err)
		assert.Empty(t, resources)
	})
}

func TestCollectionContainsAll_Chunked(t *testing.T) {
	// All tasks the filter matches, interleaved with one the user cannot see, so a chunk
	// boundary cannot be hidden by everything matching anyway.
	candidates := []int64{1, 40, 41, 42, 43, 44, 45, 46}
	allMembers := map[int64]bool{40: true, 41: true, 42: true, 43: true, 44: true, 45: true, 46: true}

	// A full-size fixture would need 500+ tasks; lowering the bound exercises the same code.
	useChunkSize := func(t *testing.T, size int) {
		t.Helper()

		previous := membershipQueryChunkSize
		membershipQueryChunkSize = size
		t.Cleanup(func() { membershipQueryChunkSize = previous })
	}

	membersOf := func(t *testing.T, storage *VikunjaCaldavProjectStorage, taskIDs []int64) map[int64]bool {
		t.Helper()

		s := db.NewSession()
		defer s.Close()

		members, err := storage.collectionContainsAll(s, taskIDs)
		require.NoError(t, err)
		return members
	}

	t.Run("resolves every member of a candidate set spanning several chunks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		useChunkSize(t, 3)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)

		require.Len(t, slices.Collect(slices.Chunk(candidates, membershipQueryChunkSize)), 3,
			"the candidates must span more than one query")
		assert.Equal(t, allMembers, membersOf(t, filterStorage(pseudoID), candidates))
	})

	t.Run("one query per candidate resolves the same members", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		useChunkSize(t, 1)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)

		assert.Equal(t, allMembers, membersOf(t, filterStorage(pseudoID), candidates))
	})

	t.Run("a single chunk resolves the same members", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)

		require.Len(t, slices.Collect(slices.Chunk(candidates, membershipQueryChunkSize)), 1)
		assert.Equal(t, allMembers, membersOf(t, filterStorage(pseudoID), candidates))
	})

	t.Run("no candidates means no query", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		useChunkSize(t, 3)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)

		assert.Empty(t, membersOf(t, filterStorage(pseudoID), nil))
	})

	// Whichever chunk trips the permission error, the answer is "no members" - not the
	// members earlier chunks already matched.
	t.Run("an unreadable filter yields no members at all", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		useChunkSize(t, 3)

		assert.Empty(t, membersOf(t, filterStorage(foreignFilterProjectID), candidates))
	})
}

// Project 21 in the fixtures is not archived itself, but its parent 22 is.
const inheritedArchivedProjectID = 21

func archiveProject(t *testing.T, projectID int64) {
	t.Helper()

	s := db.NewSession()
	defer s.Close()

	_, err := s.Where(builder.Eq{"id": projectID}).
		Cols("is_archived").
		Update(&models.Project{IsArchived: true})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

func favoriteTask(t *testing.T, taskID int64, u *user.User) {
	t.Helper()

	s := db.NewSession()
	defer s.Close()

	_, err := s.Insert(&models.Favorite{EntityID: taskID, UserID: u.ID, Kind: models.FavoriteKindTask})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

// Project.CanWrite returns ErrProjectIsArchived as its error; propagating it made
// caldav-go answer 500 instead of 403.
func TestArchivedProject_WritesAreForbidden(t *testing.T) {
	const taskContent = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid-caldav-new
DTSTAMP:20230301T073337Z
SUMMARY:New task
END:VTODO
END:VCALENDAR`

	t.Run("create is forbidden", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		archiveProject(t, 36)

		storage := storageFor(caldavFilterUser, 36)
		storage.task = &models.Task{UID: "uid-caldav-new"}

		_, err := storage.CreateResource(projectPath(36)+"uid-caldav-new.ics", taskContent)
		require.ErrorIs(t, err, errs.ForbiddenError)
		db.AssertMissing(t, "tasks", map[string]interface{}{"uid": "uid-caldav-new"})
	})

	t.Run("update is forbidden", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		archiveProject(t, 36)

		storage := storageFor(caldavFilterUser, 36)
		storage.task = &models.Task{ID: 40, UID: "uid-caldav-test"}

		_, err := storage.UpdateResource(projectPath(36)+"uid-caldav-test.ics", taskContent)
		require.ErrorIs(t, err, errs.ForbiddenError)
		db.AssertCount(t, "tasks", builder.Eq{"id": 40, "title": "Title Caldav Test"}, 1)
	})

	t.Run("delete is forbidden", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		archiveProject(t, 36)

		storage := storageFor(caldavFilterUser, 36)
		storage.task = &models.Task{ID: 40, UID: "uid-caldav-test"}

		err := storage.DeleteResource(projectPath(36) + "uid-caldav-test.ics")
		require.ErrorIs(t, err, errs.ForbiddenError)
		// Deletion only stamps deleted_at, so mere existence of the row proves nothing.
		db.AssertCount(t, "tasks", builder.And(builder.Eq{"id": 40}, builder.IsNull{"deleted_at"}), 1)
	})

	// Favorites, unlike a saved filter, keep aggregating tasks of archived projects.
	t.Run("a task reached through favorites is forbidden too", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		favoriteTask(t, 40, caldavFilterUser)
		archiveProject(t, 36)

		rpath := projectPath(models.FavoritesPseudoProjectID) + "uid-caldav-test.ics"
		storage := filterStorage(models.FavoritesPseudoProjectID)
		storage.task = &models.Task{ID: 40, UID: "uid-caldav-test", ProjectID: 36}

		_, err := storage.UpdateResource(rpath, taskContent)
		require.ErrorIs(t, err, errs.ForbiddenError)
		db.AssertCount(t, "tasks", builder.Eq{"id": 40, "title": "Title Caldav Test"}, 1)

		require.ErrorIs(t, storage.DeleteResource(rpath), errs.ForbiddenError)
		db.AssertCount(t, "tasks", builder.And(builder.Eq{"id": 40}, builder.IsNull{"deleted_at"}), 1)
	})

	// A saved filter only ever spans non-archived projects, so the task stops being a member.
	t.Run("a task reached through a saved filter is not found", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		archiveProject(t, 36)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)
		storage := filterStorage(pseudoID)
		storage.task = &models.Task{ID: 40, UID: "uid-caldav-test", ProjectID: 36}

		_, err := storage.UpdateResource(projectPath(pseudoID)+"uid-caldav-test.ics", taskContent)
		require.ErrorIs(t, err, errs.ResourceNotFoundError)
		db.AssertCount(t, "tasks", builder.Eq{"id": 40, "title": "Title Caldav Test"}, 1)
	})

	t.Run("archival inherited from a parent project is forbidden too", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		storage := storageFor(caldavOtherUser, inheritedArchivedProjectID)
		storage.task = &models.Task{UID: "uid-caldav-new"}

		_, err := storage.CreateResource(projectPath(inheritedArchivedProjectID)+"uid-caldav-new.ics", taskContent)
		require.ErrorIs(t, err, errs.ForbiddenError)
		db.AssertMissing(t, "tasks", map[string]interface{}{"uid": "uid-caldav-new"})
	})
}

// caldav-go dispatches a PUT whose resource GetResource could not find to CreateResource,
// so CreateResource decides the answer for stale hrefs and unreadable collections alike.
func TestCreateResource_MissesBeforeDenials(t *testing.T) {
	const newTaskContent = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid-caldav-new
DTSTAMP:20230301T073337Z
SUMMARY:New task
END:VTODO
END:VCALENDAR`

	const existingTaskContent = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid-caldav-test
DTSTAMP:20230301T073337Z
SUMMARY:Recreated through the filter
END:VTODO
END:VCALENDAR`

	create := func(t *testing.T, storage *VikunjaCaldavProjectStorage, uid, content string) error {
		t.Helper()

		storage.task = &models.Task{UID: uid}
		// caldav-go only reaches CreateResource once GetResource reported a miss.
		_, found, err := storage.GetResource(projectPath(storage.project.ID) + uid + ".ics")
		require.ErrorIs(t, err, errs.ResourceNotFoundError)
		require.False(t, found)

		storage.task = &models.Task{UID: uid}
		_, err = storage.CreateResource(projectPath(storage.project.ID)+uid+".ics", content)
		return err
	}

	t.Run("a task the pseudo collection does not contain is not found", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingNothing)

		err := create(t, filterStorage(pseudoID), "uid-caldav-test", existingTaskContent)
		require.ErrorIs(t, err, errs.ResourceNotFoundError)
		db.AssertCount(t, "tasks", builder.Eq{"uid": "uid-caldav-test"}, 1)
	})

	t.Run("a saved filter of another user is not found", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		err := create(t, filterStorage(foreignFilterProjectID), "uid-caldav-new", newTaskContent)
		require.ErrorIs(t, err, errs.ResourceNotFoundError)
		db.AssertCount(t, "tasks", builder.Eq{"uid": "uid-caldav-new"}, 0)
	})

	t.Run("a project the user has no access to is not found", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		err := create(t, storageFor(caldavOtherUser, inaccessibleProjectID), "uid-caldav-new", newTaskContent)
		require.ErrorIs(t, err, errs.ResourceNotFoundError)
		db.AssertCount(t, "tasks", builder.Eq{"uid": "uid-caldav-new"}, 0)
	})

	t.Run("a readable project without write access stays forbidden", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		// Project 3 is owned by user 3 and shared read-only with user 1.
		err := create(t, storageFor(caldavOtherUser, 3), "uid-caldav-new", newTaskContent)
		require.ErrorIs(t, err, errs.ForbiddenError)
		db.AssertCount(t, "tasks", builder.Eq{"uid": "uid-caldav-new"}, 0)
	})

	t.Run("a new task in a writable project is created", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		err := create(t, storageFor(caldavFilterUser, 36), "uid-caldav-new", newTaskContent)
		require.NoError(t, err)
		db.AssertCount(t, "tasks", builder.Eq{
			"uid":        "uid-caldav-new",
			"title":      "New task",
			"project_id": 36,
		}, 1)
	})
}
