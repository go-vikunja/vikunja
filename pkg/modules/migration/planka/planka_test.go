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

package planka

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// plankaFixtures are the testdata payloads a Planka v2 instance would answer with.
var plankaFixtures = map[string]string{
	"/api/projects":            "projects.json",
	"/api/boards/10":           "board_10.json",
	"/api/boards/11":           "board_11.json",
	"/api/boards/20":           "board_20.json",
	"/api/boards/30":           "board_v1.json",
	"/api/lists/1003/cards":    "list_1003_cards.json",
	"/api/cards/2000/comments": "card_2000_comments.json",
}

func TestFetchAll(t *testing.T) {
	f, srv := newFake(t)
	f.validAPIKey = "key"
	f.fixtures = plankaFixtures

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	require.NoError(t, c.login(t.Context(), "key", "", ""))

	data, err := fetchAll(c)
	require.NoError(t, err)

	assert.Equal(t, "1", data.CurrentUserID)
	assert.Equal(t, "Base Group", data.BaseCustomFieldGroups["900"].Name)
	assert.Equal(t, "Other Person", data.Users["2"].Name)

	require.Len(t, data.Projects, 2)
	multi := data.Projects[0]
	assert.Equal(t, "Multi Board Project", multi.Name)
	require.Len(t, multi.Boards, 2)
	assert.Equal(t, "Board A", multi.Boards[0].Name, "boards sorted by position")
	assert.Equal(t, "Board B", multi.Boards[1].Name)

	boardA := multi.Boards[0]
	assert.Len(t, boardA.Cards, 5, "4 board cards + 1 archived card")
	assert.Equal(t, "Archived card", boardA.Cards[4].Name)
	assert.Len(t, boardA.CardLabels, 4, "archived card labels merged in")

	comments := boardA.Comments["2000"]
	require.Len(t, comments, 2)
	assert.Equal(t, "6000", comments[0].ID, "oldest first")
	assert.Equal(t, "6001", comments[1].ID)

	single := data.Projects[1]
	require.Len(t, single.Boards, 1)
	assert.Equal(t, "Only Board", single.Boards[0].Name)
}

func TestFetchBoardV1Unsupported(t *testing.T) {
	f, srv := newFake(t)
	f.validAPIKey = "key"
	f.fixtures = plankaFixtures

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	require.NoError(t, c.login(t.Context(), "key", "", ""))

	_, err = fetchBoard(c, "30")
	var errVersion *ErrUnsupportedVersion
	require.ErrorAs(t, err, &errVersion)
}

func TestConvertPlankaToVikunja(t *testing.T) {
	f, srv := newFake(t)
	f.validAPIKey = "key"
	f.fixtures = plankaFixtures

	c, err := newClient(srv.URL)
	require.NoError(t, err)
	require.NoError(t, c.login(t.Context(), "key", "", ""))

	data, err := fetchAll(c)
	require.NoError(t, err)

	downloads := []string{}
	projects, err := convertPlankaToVikunja(data, func(a *plankaAttachment) (*bytes.Buffer, error) {
		downloads = append(downloads, a.ID)
		if a.ID == "3001" {
			return nil, errors.New("boom")
		}
		return bytes.NewBufferString("PNG!"), nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"3000", "3001"}, downloads, "link attachments are not downloaded")

	require.Len(t, projects, 5)

	t.Run("hierarchy", func(t *testing.T) {
		root, parent, boardA, boardB, single := projects[0], projects[1], projects[2], projects[3], projects[4]

		assert.Equal(t, "Migrated from Planka", root.Title)
		assert.Nil(t, root.ParentProjectID)

		assert.Equal(t, "Multi Board Project", parent.Title)
		assert.Equal(t, "Two boards", parent.Description)
		require.NotNil(t, parent.ParentProjectID)
		assert.Equal(t, root.ID, *parent.ParentProjectID)
		assert.Empty(t, parent.Tasks)

		assert.Equal(t, "Board A", boardA.Title)
		require.NotNil(t, boardA.ParentProjectID)
		assert.Equal(t, parent.ID, *boardA.ParentProjectID)
		assert.Equal(t, "Board B", boardB.Title)
		require.NotNil(t, boardB.ParentProjectID)
		assert.Equal(t, parent.ID, *boardB.ParentProjectID)

		assert.Equal(t, "Single Board Project", single.Title, "single-board project keeps the project name")
		require.NotNil(t, single.ParentProjectID)
		assert.Equal(t, root.ID, *single.ParentProjectID)
		require.Len(t, single.Tasks, 1)
		assert.Equal(t, "Lonely card", single.Tasks[0].Title)
		require.Len(t, single.Buckets, 1)
		assert.Equal(t, single.Buckets[0].ID, single.Tasks[0].BucketID)
	})

	boardA := projects[2]

	t.Run("buckets", func(t *testing.T) {
		require.Len(t, boardA.Buckets, 4, "trash list is skipped")
		titles := []string{}
		for _, b := range boardA.Buckets {
			titles = append(titles, b.Title)
		}
		assert.Equal(t, []string{"To Do", "In Progress", "Done", "Archive"}, titles)

		ids := map[int64]bool{}
		for _, p := range projects {
			for _, b := range p.Buckets {
				assert.False(t, ids[b.ID], "bucket ids must be unique across projects")
				ids[b.ID] = true
			}
		}
	})

	tasksByTitle := map[string]*models.TaskWithComments{}
	for _, task := range boardA.Tasks {
		tasksByTitle[task.Title] = task
	}
	require.Len(t, tasksByTitle, 5)
	bucketByTitle := map[string]int64{}
	for _, b := range boardA.Buckets {
		bucketByTitle[b.Title] = b.ID
	}

	t.Run("order and done state", func(t *testing.T) {
		assert.Equal(t, "Zeroth card", boardA.Tasks[0].Title, "cards sorted by position")
		assert.Equal(t, "First card", boardA.Tasks[1].Title)

		assert.False(t, tasksByTitle["Zeroth card"].Done)
		assert.Equal(t, bucketByTitle["To Do"], tasksByTitle["Zeroth card"].BucketID)
		assert.True(t, tasksByTitle["Closed card"].Done, "cards in closed lists are done")
		assert.Equal(t, bucketByTitle["Done"], tasksByTitle["Closed card"].BucketID)
		assert.True(t, tasksByTitle["Card with checklist"].Done, "isClosed cards are done")
		assert.True(t, tasksByTitle["Archived card"].Done, "archived cards are done")
		assert.Equal(t, bucketByTitle["Archive"], tasksByTitle["Archived card"].BucketID)
	})

	first := tasksByTitle["First card"]

	t.Run("card fields", func(t *testing.T) {
		assert.Equal(t, time.Date(2024, 3, 1, 10, 0, 0, 0, time.UTC), first.DueDate.UTC())
		assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), first.Created.UTC())
		assert.Contains(t, first.Description, "<strong>markdown</strong>")
	})

	t.Run("labels", func(t *testing.T) {
		require.Len(t, first.Labels, 2)
		assert.Equal(t, "Bug", first.Labels[0].Title)
		assert.Equal(t, "e83855", first.Labels[0].HexColor)
		assert.Equal(t, "Weird", first.Labels[1].Title)
		assert.Empty(t, first.Labels[1].HexColor, "unknown colors map to no color")

		checklist := tasksByTitle["Card with checklist"]
		require.Len(t, checklist.Labels, 1)
		assert.Equal(t, "sunny-grass", checklist.Labels[0].Title, "unnamed labels fall back to the color name")

		archived := tasksByTitle["Archived card"]
		require.Len(t, archived.Labels, 1)
		assert.Equal(t, "Bug", archived.Labels[0].Title)
	})

	t.Run("checklist", func(t *testing.T) {
		desc := tasksByTitle["Card with checklist"].Description
		assert.Contains(t, desc, "Steps")
		assert.Contains(t, desc, `data-type="taskList"`)
		assert.Contains(t, desc, `data-checked="true"`)
		assert.Contains(t, desc, `data-checked="false"`)
		assert.Less(t, indexOf(desc, "first step"), indexOf(desc, "second step"), "checklist items sorted by position")
	})

	t.Run("custom fields and links", func(t *testing.T) {
		assert.Contains(t, first.Description, "<h2>Details</h2>")
		assert.Contains(t, first.Description, "<td>Priority</td>")
		assert.Contains(t, first.Description, "<td>High</td>")
		assert.Contains(t, first.Description, "<h2>Base Group</h2>", "group name derived from base group")
		assert.Contains(t, first.Description, "<td>Estimate</td>")
		assert.Contains(t, first.Description, "<td>3d</td>")
		assert.Contains(t, first.Description, `<a href="https://vikunja.io">Vikunja</a>`)
	})

	t.Run("attachments and cover", func(t *testing.T) {
		require.Len(t, first.Attachments, 1, "failed download is skipped")
		att := first.Attachments[0]
		assert.Equal(t, "cover.png", att.File.Name)
		assert.Equal(t, "image/png", att.File.Mime)
		assert.Equal(t, uint64(4), att.File.Size)
		assert.Equal(t, []byte("PNG!"), att.File.FileContent)
		assert.NotZero(t, att.ID)
		assert.Equal(t, att.ID, first.CoverImageAttachmentID)
	})

	t.Run("comments", func(t *testing.T) {
		require.Len(t, first.Comments, 2)
		assert.Contains(t, first.Comments[0].Comment, "<em>Other Person</em>:")
		assert.Contains(t, first.Comments[0].Comment, "Hello from <em>someone else</em>")
		assert.Equal(t, time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), first.Comments[0].Created.UTC())
		assert.NotContains(t, first.Comments[1].Comment, "<em>Me</em>", "own comments are not prefixed")
		assert.Contains(t, first.Comments[1].Comment, "Reply from me")
	})

	t.Run("no other user data leaks", func(t *testing.T) {
		out, err := json.Marshal(projects)
		require.NoError(t, err)
		assert.NotContains(t, string(out), "example.com")
		assert.NotContains(t, string(out), `"assignees":[{`)
		for _, p := range projects {
			for _, task := range p.Tasks {
				assert.Empty(t, task.Assignees)
			}
		}
	})
}

func indexOf(s, sub string) int {
	return bytes.Index([]byte(s), []byte(sub))
}

func TestMigrate(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	f, srv := newFake(t)
	f.validAPIKey = "key"
	f.fixtures = plankaFixtures
	u := &user.User{ID: 1}

	m := &Migrator{URL: srv.URL + "/api/", Token: "key"}
	require.NoError(t, m.Migrate(u))

	db.AssertExists(t, "projects", map[string]interface{}{"title": "Migrated from Planka", "owner_id": u.ID}, false)
	s := db.NewSession()
	defer s.Close()
	root := &models.Project{}
	_, err := s.Where("title = ?", "Migrated from Planka").Get(root)
	require.NoError(t, err)

	db.AssertExists(t, "projects", map[string]interface{}{"title": "Multi Board Project", "owner_id": u.ID, "parent_project_id": root.ID}, false)
	parent := &models.Project{}
	_, err = s.Where("title = ?", "Multi Board Project").Get(parent)
	require.NoError(t, err)
	db.AssertExists(t, "projects", map[string]interface{}{"title": "Board A", "owner_id": u.ID, "parent_project_id": parent.ID}, false)
	db.AssertExists(t, "projects", map[string]interface{}{"title": "Board B", "owner_id": u.ID, "parent_project_id": parent.ID}, false)
	db.AssertExists(t, "projects", map[string]interface{}{"title": "Single Board Project", "owner_id": u.ID, "parent_project_id": root.ID}, false)
	db.AssertExists(t, "tasks", map[string]interface{}{"title": "First card", "created_by_id": u.ID, "done": false}, false)
	db.AssertExists(t, "tasks", map[string]interface{}{"title": "Archived card", "created_by_id": u.ID, "done": true}, false)
	db.AssertExists(t, "buckets", map[string]interface{}{"title": "Archive", "created_by_id": u.ID}, false)
	db.AssertExists(t, "labels", map[string]interface{}{"title": "Bug", "hex_color": "e83855", "created_by_id": u.ID}, false)
	db.AssertExists(t, "task_comments", map[string]interface{}{"comment": "<p>Reply from me</p>", "author_id": u.ID}, false)
	db.AssertExists(t, "task_comments", map[string]interface{}{"comment": "<p><em>Other Person</em>:</p>\n<p>Hello from <em>someone else</em></p>", "author_id": u.ID}, false)
	db.AssertExists(t, "files", map[string]interface{}{"name": "cover.png", "created_by_id": u.ID}, false)
}
