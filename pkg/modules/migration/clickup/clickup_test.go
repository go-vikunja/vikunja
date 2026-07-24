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

package clickup

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseClickupTimestamp(t *testing.T) {
	tm, ok := parseClickupTimestamp("1609459200000") // 2021-01-01T00:00:00Z
	require.True(t, ok)
	assert.Equal(t, 2021, tm.UTC().Year())

	_, ok = parseClickupTimestamp("")
	assert.False(t, ok)

	_, ok = parseClickupTimestamp("not-a-number")
	assert.False(t, ok)
}

func TestConvertTaskToVikunja(t *testing.T) {
	m := &Migration{Code: "test-token"}

	ct := &clickupTask{
		ID:          "abc123",
		Name:        "A migrated task",
		Description: "Some description",
		Status:      clickupStatus{Status: "closed", Type: "closed"},
		DateClosed:  "1609459200000",
		DueDate:     "1609459200000",
		Priority:    &clickupPriority{Priority: "urgent"},
		Tags:        []clickupTag{{Name: "bug"}, {Name: "urgent-fix"}},
	}

	task := m.convertTaskToVikunja(ct, 42)

	assert.Equal(t, "A migrated task", task.Title)
	assert.Equal(t, "Some description", task.Description)
	assert.Equal(t, int64(42), task.BucketID)
	assert.True(t, task.Done)
	assert.Equal(t, 2021, task.DoneAt.UTC().Year())
	assert.Equal(t, int64(4), task.Priority)
	assert.Equal(t, 2021, task.DueDate.UTC().Year())
	require.Len(t, task.Labels, 2)
	assert.Equal(t, "bug", task.Labels[0].Title)
	assert.Equal(t, "urgent-fix", task.Labels[1].Title)
	assert.Empty(t, task.Attachments)
}

func TestConvertTaskToVikunjaOpenTaskIsNotDone(t *testing.T) {
	m := &Migration{Code: "test-token"}

	ct := &clickupTask{
		ID:     "open-task",
		Name:   "Still open",
		Status: clickupStatus{Status: "in progress", Type: "custom"},
	}

	task := m.convertTaskToVikunja(ct, 1)
	assert.False(t, task.Done)
	assert.Equal(t, int64(0), task.Priority)
}

func TestConvertTaskToVikunjaDownloadsAttachments(t *testing.T) {
	const fileContent = "attachment bytes"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(fileContent))
	}))
	defer server.Close()

	// The SSRF-safe client used by migration.DownloadFile rejects non-routable
	// IPs by default; allow them for the duration of this test so it can hit
	// the local httptest server instead of depending on a real external host.
	prevAllowNonRoutable := config.OutgoingRequestsAllowNonRoutableIPs.GetBool()
	config.OutgoingRequestsAllowNonRoutableIPs.Set("true")
	t.Cleanup(func() {
		config.OutgoingRequestsAllowNonRoutableIPs.Set(prevAllowNonRoutable)
	})

	m := &Migration{Code: "test-token"}
	ct := &clickupTask{
		ID:   "with-attachment",
		Name: "Has an attachment",
		Attachments: []clickupAttachment{
			{ID: "att1", Title: "screenshot.png", URL: server.URL},
		},
	}

	task := m.convertTaskToVikunja(ct, 1)
	require.Len(t, task.Attachments, 1)
	assert.Equal(t, "screenshot.png", task.Attachments[0].File.Name)
	assert.Equal(t, uint64(len(fileContent)), task.Attachments[0].File.Size)
}

func TestResolveSubtaskRelations(t *testing.T) {
	parent := &models.TaskWithComments{Task: models.Task{Title: "parent"}}
	child := &models.TaskWithComments{Task: models.Task{Title: "child"}}
	orphan := &models.TaskWithComments{Task: models.Task{Title: "orphan, parent not in this token's scope"}}

	tasksByClickupID := map[string]*models.Task{
		"parent-id": &parent.Task,
		"child-id":  &child.Task,
	}
	pending := []subtaskLink{
		{task: child, parentID: "parent-id"},
		{task: orphan, parentID: "missing-parent-id"},
	}

	resolveSubtaskRelations(pending, tasksByClickupID)

	require.Contains(t, parent.RelatedTasks, models.RelationKindSubtask)
	require.Len(t, parent.RelatedTasks[models.RelationKindSubtask], 1)
	assert.Equal(t, "child", parent.RelatedTasks[models.RelationKindSubtask][0].Title)
	assert.Empty(t, orphan.RelatedTasks)
}

func TestName(t *testing.T) {
	assert.Equal(t, "clickup", (&Migration{}).Name())
}

func TestIsNotAnOAuthMigrator(t *testing.T) {
	// ClickUp authenticates with a pasted personal token, not an OAuth
	// redirect - not implementing migration.OAuthMigrator is what keeps the
	// /auth route from being registered for it.
	var m interface{} = &Migration{}
	_, isOAuth := m.(migration.OAuthMigrator)
	assert.False(t, isOAuth)
	_, isMigrator := m.(migration.Migrator)
	assert.True(t, isMigrator)
}

func TestDecodeResponseRejectsNonSuccessStatus(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Body:       io.NopCloser(strings.NewReader(`{"err":"Token invalid","ECODE":"OAUTH_025"}`)),
	}

	r := &teamsResponse{}
	err := decodeResponse(resp, r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	assert.Contains(t, err.Error(), "Token invalid")
	assert.Empty(t, r.Teams)
}

func TestDecodeResponseDecodesSuccess(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"teams":[{"id":"1","name":"A Team"}]}`)),
	}

	r := &teamsResponse{}
	require.NoError(t, decodeResponse(resp, r))
	require.Len(t, r.Teams, 1)
	assert.Equal(t, "A Team", r.Teams[0].Name)
}
