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
	"slices"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/samedi/caldav-go/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
