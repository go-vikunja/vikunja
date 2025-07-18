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

package webtests

import (
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskCollection(t *testing.T) {
	testHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.TaskCollection{}
		},
		t: t,
	}
	t.Run("ReadAll on project", func(t *testing.T) {

		urlParams := map[string]string{"project": "1"}

		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(nil, urlParams)
			require.NoError(t, err)
			// Not using assert.Equal to avoid having the tests break every time we add new fixtures
			assert.Contains(t, rec.Body.String(), `task #1`)
			assert.Contains(t, rec.Body.String(), `task #2 `)
			assert.Contains(t, rec.Body.String(), `task #3 `)
			assert.Contains(t, rec.Body.String(), `task #4 `)
			assert.Contains(t, rec.Body.String(), `task #5 `)
			assert.Contains(t, rec.Body.String(), `task #6 `)
			assert.Contains(t, rec.Body.String(), `task #7 `)
			assert.Contains(t, rec.Body.String(), `task #8 `)
			assert.Contains(t, rec.Body.String(), `task #9 `)
			assert.Contains(t, rec.Body.String(), `task #10`)
			assert.Contains(t, rec.Body.String(), `task #11`)
			assert.Contains(t, rec.Body.String(), `task #12`)
			assert.NotContains(t, rec.Body.String(), `task #13`)
			assert.NotContains(t, rec.Body.String(), `task #14`)
			assert.NotContains(t, rec.Body.String(), `task #15`) // Shared via team readonly
			assert.NotContains(t, rec.Body.String(), `task #16`) // Shared via team write
			assert.NotContains(t, rec.Body.String(), `task #17`) // Shared via team admin
			assert.NotContains(t, rec.Body.String(), `task #18`) // Shared via user readonly
			assert.NotContains(t, rec.Body.String(), `task #19`) // Shared via user write
			assert.NotContains(t, rec.Body.String(), `task #20`) // Shared via user admin
			assert.NotContains(t, rec.Body.String(), `task #21`) // Shared via namespace team readonly
			assert.NotContains(t, rec.Body.String(), `task #22`) // Shared via namespace team write
			assert.NotContains(t, rec.Body.String(), `task #23`) // Shared via namespace team admin
			assert.NotContains(t, rec.Body.String(), `task #24`) // Shared via namespace user readonly
			assert.NotContains(t, rec.Body.String(), `task #25`) // Shared via namespace user write
			assert.NotContains(t, rec.Body.String(), `task #26`) // Shared via namespace user admin
			assert.Contains(t, rec.Body.String(), `task #27`)
			assert.Contains(t, rec.Body.String(), `task #28`)
			assert.NotContains(t, rec.Body.String(), `task #32`)
		})
		t.Run("Search", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(url.Values{"s": []string{"unique"}}, urlParams)
			require.NoError(t, err)
			assert.NotContains(t, rec.Body.String(), `task #1`)
			assert.NotContains(t, rec.Body.String(), `task #2 `)
			assert.NotContains(t, rec.Body.String(), `task #3 `)
			assert.NotContains(t, rec.Body.String(), `task #4 `)
			assert.NotContains(t, rec.Body.String(), `task #5 `)
			assert.Contains(t, rec.Body.String(), `task #6 `)
			assert.NotContains(t, rec.Body.String(), `task #7 `)
			assert.NotContains(t, rec.Body.String(), `task #8 `)
			assert.NotContains(t, rec.Body.String(), `task #9 `)
			assert.NotContains(t, rec.Body.String(), `task #10`)
			assert.NotContains(t, rec.Body.String(), `task #11`)
			assert.NotContains(t, rec.Body.String(), `task #12`)
			assert.NotContains(t, rec.Body.String(), `task #13`)
			assert.NotContains(t, rec.Body.String(), `task #14`)
		})
		t.Run("Search case insensitive", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(url.Values{"s": []string{"uNIQue"}}, urlParams)
			require.NoError(t, err)
			assert.NotContains(t, rec.Body.String(), `task #1`)
			assert.NotContains(t, rec.Body.String(), `task #2 `)
			assert.NotContains(t, rec.Body.String(), `task #3 `)
			assert.NotContains(t, rec.Body.String(), `task #4 `)
			assert.NotContains(t, rec.Body.String(), `task #5 `)
			assert.Contains(t, rec.Body.String(), `task #6 `)
			assert.NotContains(t, rec.Body.String(), `task #7 `)
			assert.NotContains(t, rec.Body.String(), `task #8 `)
			assert.NotContains(t, rec.Body.String(), `task #9 `)
			assert.NotContains(t, rec.Body.String(), `task #10`)
			assert.NotContains(t, rec.Body.String(), `task #11`)
			assert.NotContains(t, rec.Body.String(), `task #12`)
			assert.NotContains(t, rec.Body.String(), `task #13`)
			assert.NotContains(t, rec.Body.String(), `task #14`)
		})
		t.Run("Sort Order", func(t *testing.T) {
			// TODO: Add more cases
			// should equal priority asc
			t.Run("by priority", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"priority"}}, urlParams)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `{"id":33,"title":"task #33 with percent done","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0.5,"identifier":"test1-17","index":17,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}}]`)
			})
			t.Run("by priority desc", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"priority"}, "order_by": []string{"desc"}}, urlParams)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":3,"title":"task #3 high prio","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":100,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"test1-3","index":3,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}},{"id":4,"title":"task #4 low prio","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":1`)
			})
			t.Run("by priority asc", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"priority"}, "order_by": []string{"asc"}}, urlParams)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `{"id":33,"title":"task #33 with percent done","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0.5,"identifier":"test1-17","index":17,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}}]`)
			})
			// should equal duedate asc
			t.Run("by due_date", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"due_date"}}, urlParams)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":6,"title":"task #6 lower due date","description":"This has something unique","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"2018-11-30T22:25:24Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"test1-6","index":6,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}}`)
			})
			t.Run("by duedate desc", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"due_date"}, "order_by": []string{"desc"}}, urlParams)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":5,"title":"task #5 higher due date","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"2018-12-01T03:58:44Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"test1-5","index":5,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}},{"id":6,"title":"task #6 lower due date`)
			})
			// Due date without unix suffix
			t.Run("by duedate asc without  suffix", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"due_date"}, "order_by": []string{"asc"}}, urlParams)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":6,"title":"task #6 lower due date","description":"This has something unique","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"2018-11-30T22:25:24Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"test1-6","index":6,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}}`)
			})
			t.Run("by due_date without suffix", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"due_date"}}, urlParams)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":6,"title":"task #6 lower due date","description":"This has something unique","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"2018-11-30T22:25:24Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"test1-6","index":6,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}}`)
			})
			t.Run("by duedate desc without  suffix", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"due_date"}, "order_by": []string{"desc"}}, urlParams)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":5,"title":"task #5 higher due date","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"2018-12-01T03:58:44Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"test1-5","index":5,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}},{"id":6,"title":"task #6 lower due date`)
			})
			t.Run("by duedate asc", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"due_date"}, "order_by": []string{"asc"}}, urlParams)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":6,"title":"task #6 lower due date","description":"This has something unique","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"2018-11-30T22:25:24Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"test1-6","index":6,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}}`)
			})
			t.Run("invalid sort parameter", func(t *testing.T) {
				_, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"loremipsum"}}, urlParams)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeInvalidTaskField)
			})
			t.Run("invalid sort order", func(t *testing.T) {
				_, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"id"}, "order_by": []string{"loremipsum"}}, urlParams)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeInvalidSortOrder)
			})
			t.Run("invalid parameter", func(t *testing.T) {
				// Invalid parameter should not sort at all
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort": []string{"loremipsum"}}, urlParams)
				require.NoError(t, err)
				assert.NotContains(t, rec.Body.String(), `[{"id":3,"title":"task #3 high prio","description":"","done":false,"due_date":0,"reminders":null,"repeat_after":0,"repeat_mode":0,"priority":100,"start_date":0,"end_date":0,"assignees":null,"labels":null,"hex_color":"","created":1543626724,"updated":1543626724,"created_by":{"id":0,"name":"","username":"","email":"","created":0,"updated":0}},{"id":4,"title":"task #4 low prio","description":"","done":false,"due_date":0,"repeat_after":0,"repeat_mode":0,"priority":1`)
				assert.NotContains(t, rec.Body.String(), `{"id":4,"title":"task #4 low prio","description":"","done":false,"due_date":0,"reminders":null,"repeat_after":0,"repeat_mode":0,"priority":1,"start_date":0,"end_date":0,"assignees":null,"labels":null,"hex_color":"","created":1543626724,"updated":1543626724,"created_by":{"id":0,"name":"","username":"","email":"","created":0,"updated":0}},{"id":3,"title":"task #3 high prio","description":"","done":false,"due_date":0,"repeat_after":0,"repeat_mode":0,"priority":100,"start_date":0,"end_date":0,"assignees":null,"labels":null,"created":1543626724,"updated":1543626724,"created_by":{"id":0,"name":"","username":"","email":"","created":0,"updated":0}}]`)
				assert.NotContains(t, rec.Body.String(), `[{"id":5,"title":"task #5 higher due date","description":"","done":false,"due_date":1543636724,"reminders":null,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":0,"end_date":0,"assignees":null,"labels":null,"hex_color":"","created":1543626724,"updated":1543626724,"created_by":{"id":0,"name":"","username":"","email":"","created":0,"updated":0}},{"id":6,"title":"task #6 lower due date"`)
				assert.NotContains(t, rec.Body.String(), `{"id":6,"title":"task #6 lower due date","description":"","done":false,"due_date":1543616724,"reminders":null,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":0,"end_date":0,"assignees":null,"labels":null,"hex_color":"","created":1543626724,"updated":1543626724,"created_by":{"id":0,"name":"","username":"","email":"","created":0,"updated":0}},{"id":5,"title":"task #5 higher due date","description":"","done":false,"due_date":1543636724,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":0,"end_date":0,"assignees":null,"labels":null,"created":1543626724,"updated":1543626724,"created_by":{"id":0,"name":"","username":"","email":"","created":0,"updated":0}}]`)
			})
		})
		t.Run("Filter", func(t *testing.T) {
			t.Run("Date range", func(t *testing.T) {
				t.Run("start and end date", func(t *testing.T) {
					rec, err := testHandler.testReadAllWithUser(
						url.Values{
							"filter": []string{"start_date > '2018-12-11T03:46:40+00:00' || end_date < '2018-12-13T11:20:01+00:00' || due_date > '2018-11-29T14:00:00+00:00'"},
						},
						urlParams,
					)
					require.NoError(t, err)
					assert.NotContains(t, rec.Body.String(), `task #1`)
					assert.NotContains(t, rec.Body.String(), `task #2 `)
					assert.NotContains(t, rec.Body.String(), `task #3 `)
					assert.NotContains(t, rec.Body.String(), `task #4 `)
					assert.Contains(t, rec.Body.String(), `task #5 `)
					assert.Contains(t, rec.Body.String(), `task #6 `)
					assert.Contains(t, rec.Body.String(), `task #7 `)
					assert.Contains(t, rec.Body.String(), `task #8 `)
					assert.Contains(t, rec.Body.String(), `task #9 `)
					assert.NotContains(t, rec.Body.String(), `task #10`)
					assert.NotContains(t, rec.Body.String(), `task #11`)
					assert.NotContains(t, rec.Body.String(), `task #12`)
					assert.NotContains(t, rec.Body.String(), `task #13`)
					assert.NotContains(t, rec.Body.String(), `task #14`)
				})
				t.Run("start date only", func(t *testing.T) {
					rec, err := testHandler.testReadAllWithUser(
						url.Values{
							"filter": []string{"start_date > '2018-10-20T01:46:40+00:00'"},
						},
						urlParams,
					)
					require.NoError(t, err)
					assert.NotContains(t, rec.Body.String(), `task #1`)
					assert.NotContains(t, rec.Body.String(), `task #2 `)
					assert.NotContains(t, rec.Body.String(), `task #3 `)
					assert.NotContains(t, rec.Body.String(), `task #4 `)
					assert.NotContains(t, rec.Body.String(), `task #5 `)
					assert.NotContains(t, rec.Body.String(), `task #6 `)
					assert.Contains(t, rec.Body.String(), `task #7 `)
					assert.NotContains(t, rec.Body.String(), `task #8 `)
					assert.Contains(t, rec.Body.String(), `task #9 `)
					assert.NotContains(t, rec.Body.String(), `task #10`)
					assert.NotContains(t, rec.Body.String(), `task #11`)
					assert.NotContains(t, rec.Body.String(), `task #12`)
					assert.NotContains(t, rec.Body.String(), `task #13`)
					assert.NotContains(t, rec.Body.String(), `task #14`)
				})
				t.Run("end date only", func(t *testing.T) {
					rec, err := testHandler.testReadAllWithUser(
						url.Values{
							"filter": []string{"end_date > '2018-12-13T11:20:01+00:00'"},
						},
						urlParams,
					)
					require.NoError(t, err)
					// If no start date but an end date is specified, this should be null
					// since we don't have any tasks in the fixtures with an end date >
					// the current date.
					assert.Equal(t, "[]\n", rec.Body.String())
				})
				t.Run("unix timestamps", func(t *testing.T) {
					rec, err := testHandler.testReadAllWithUser(
						url.Values{
							"filter": []string{"start_date > 1544500000 || end_date < 1513164001 || due_date > 1543500000"},
						},
						urlParams,
					)
					require.NoError(t, err)
					assert.NotContains(t, rec.Body.String(), `task #1`)
					assert.NotContains(t, rec.Body.String(), `task #2 `)
					assert.NotContains(t, rec.Body.String(), `task #3 `)
					assert.NotContains(t, rec.Body.String(), `task #4 `)
					assert.Contains(t, rec.Body.String(), `task #5 `)
					assert.Contains(t, rec.Body.String(), `task #6 `)
					assert.Contains(t, rec.Body.String(), `task #7 `)
					assert.NotContains(t, rec.Body.String(), `task #8 `)
					assert.Contains(t, rec.Body.String(), `task #9 `)
					assert.NotContains(t, rec.Body.String(), `task #10`)
					assert.NotContains(t, rec.Body.String(), `task #11`)
					assert.NotContains(t, rec.Body.String(), `task #12`)
					assert.NotContains(t, rec.Body.String(), `task #13`)
					assert.NotContains(t, rec.Body.String(), `task #14`)
				})
			})
			t.Run("invalid date", func(t *testing.T) {
				_, err := testHandler.testReadAllWithUser(
					url.Values{
						"filter": []string{"due_date > invalid"},
					},
					nil,
				)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeInvalidTaskFilterValue)
			})
		})
		t.Run("saved filter", func(t *testing.T) {
			t.Run("date range", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(
					nil,
					map[string]string{"project": "-2"}, // Actually a saved filter - contains the same filter arguments as the start and end date filter from above
				)
				require.NoError(t, err)
				assert.NotContains(t, rec.Body.String(), `task #1`)
				assert.NotContains(t, rec.Body.String(), `task #2 `)
				assert.NotContains(t, rec.Body.String(), `task #3 `)
				assert.NotContains(t, rec.Body.String(), `task #4 `)
				assert.Contains(t, rec.Body.String(), `task #5 `)
				assert.Contains(t, rec.Body.String(), `task #6 `)
				assert.Contains(t, rec.Body.String(), `task #7 `)
				assert.Contains(t, rec.Body.String(), `task #8 `)
				assert.Contains(t, rec.Body.String(), `task #9 `)
				assert.NotContains(t, rec.Body.String(), `task #10`)
				assert.NotContains(t, rec.Body.String(), `task #11`)
				assert.NotContains(t, rec.Body.String(), `task #12`)
				assert.NotContains(t, rec.Body.String(), `task #13`)
				assert.NotContains(t, rec.Body.String(), `task #14`)
			})
		})
	})

	t.Run("ReadAll for all tasks", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			// Not using assert.Equal to avoid having the tests break every time we add new fixtures
			assert.Contains(t, rec.Body.String(), `task #1`)
			assert.Contains(t, rec.Body.String(), `task #2 `)
			assert.Contains(t, rec.Body.String(), `task #3 `)
			assert.Contains(t, rec.Body.String(), `task #4 `)
			assert.Contains(t, rec.Body.String(), `task #5 `)
			assert.Contains(t, rec.Body.String(), `task #6 `)
			assert.Contains(t, rec.Body.String(), `task #7 `)
			assert.Contains(t, rec.Body.String(), `task #8 `)
			assert.Contains(t, rec.Body.String(), `task #9 `)
			assert.Contains(t, rec.Body.String(), `task #10`)
			assert.Contains(t, rec.Body.String(), `task #11`)
			assert.Contains(t, rec.Body.String(), `task #12`)
			assert.NotContains(t, rec.Body.String(), `task #13`)
			assert.NotContains(t, rec.Body.String(), `task #14`)
			assert.NotContains(t, rec.Body.String(), `task #13`)
			assert.NotContains(t, rec.Body.String(), `task #14`)
			assert.Contains(t, rec.Body.String(), `task #15`) // Shared via team readonly
			assert.Contains(t, rec.Body.String(), `task #16`) // Shared via team write
			assert.Contains(t, rec.Body.String(), `task #17`) // Shared via team admin
			assert.Contains(t, rec.Body.String(), `task #18`) // Shared via user readonly
			assert.Contains(t, rec.Body.String(), `task #19`) // Shared via user write
			assert.Contains(t, rec.Body.String(), `task #20`) // Shared via user admin
			assert.Contains(t, rec.Body.String(), `task #21`) // Shared via namespace team readonly
			assert.Contains(t, rec.Body.String(), `task #22`) // Shared via namespace team write
			assert.Contains(t, rec.Body.String(), `task #23`) // Shared via namespace team admin
			assert.Contains(t, rec.Body.String(), `task #24`) // Shared via namespace user readonly
			assert.Contains(t, rec.Body.String(), `task #25`) // Shared via namespace user write
			assert.Contains(t, rec.Body.String(), `task #26`) // Shared via namespace user admin
			// TODO: Add some cases where the user has access to the project, somhow shared
		})
		t.Run("Search", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(url.Values{"s": []string{"unique"}}, nil)
			require.NoError(t, err)
			assert.NotContains(t, rec.Body.String(), `task #1`)
			assert.NotContains(t, rec.Body.String(), `task #2 `)
			assert.NotContains(t, rec.Body.String(), `task #3 `)
			assert.NotContains(t, rec.Body.String(), `task #4 `)
			assert.NotContains(t, rec.Body.String(), `task #5 `)
			assert.Contains(t, rec.Body.String(), `task #6 `)
			assert.NotContains(t, rec.Body.String(), `task #7 `)
			assert.NotContains(t, rec.Body.String(), `task #8 `)
			assert.NotContains(t, rec.Body.String(), `task #9 `)
			assert.NotContains(t, rec.Body.String(), `task #10`)
			assert.NotContains(t, rec.Body.String(), `task #11`)
			assert.NotContains(t, rec.Body.String(), `task #12`)
			assert.NotContains(t, rec.Body.String(), `task #13`)
			assert.NotContains(t, rec.Body.String(), `task #14`)
		})
		t.Run("Sort Order", func(t *testing.T) {
			// should equal priority asc
			t.Run("by priority", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"priority"}}, nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `{"id":33,"title":"task #33 with percent done","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0.5,"identifier":"test1-17","index":17,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}},{"id":35,"title":"task #35","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":21,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":[{"id":2,"name":"","username":"user2","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}],"labels":[{"id":4,"title":"Label #4 - visible via other task","description":"","hex_color":"","created_by":{"id":2,"name":"","username":"user2","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"},"created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"},{"id":5,"title":"Label #5","description":"","hex_color":"","created_by":{"id":2,"name":"","username":"user2","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"},"created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}],"hex_color":"","percent_done":0,"identifier":"test21-1","index":1,"related_tasks":{"related":[{"id":1,"title":"task #1","description":"Lorem Ipsum","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"","index":1,"related_tasks":null,"attachments":null,"cover_image_attachment_id":0,"is_favorite":true,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":null},{"id":1,"title":"task #1","description":"Lorem Ipsum","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"","index":1,"related_tasks":null,"attachments":null,"cover_image_attachment_id":0,"is_favorite":true,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":null}]},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}},{"id":39,"title":"task #39","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":25,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"#0","index":0,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}}]`)
			})
			t.Run("by priority desc", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"priority"}, "order_by": []string{"desc"}}, nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":3,"title":"task #3 high prio","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":100,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"test1-3","index":3,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}},{"id":4,"title":"task #4 low prio","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":1`)
			})
			t.Run("by priority asc", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"priority"}, "order_by": []string{"asc"}}, nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `{"id":33,"title":"task #33 with percent done","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0.5,"identifier":"test1-17","index":17,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}},{"id":35,"title":"task #35","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":21,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":[{"id":2,"name":"","username":"user2","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}],"labels":[{"id":4,"title":"Label #4 - visible via other task","description":"","hex_color":"","created_by":{"id":2,"name":"","username":"user2","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"},"created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"},{"id":5,"title":"Label #5","description":"","hex_color":"","created_by":{"id":2,"name":"","username":"user2","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"},"created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}],"hex_color":"","percent_done":0,"identifier":"test21-1","index":1,"related_tasks":{"related":[{"id":1,"title":"task #1","description":"Lorem Ipsum","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"","index":1,"related_tasks":null,"attachments":null,"cover_image_attachment_id":0,"is_favorite":true,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":null},{"id":1,"title":"task #1","description":"Lorem Ipsum","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"","index":1,"related_tasks":null,"attachments":null,"cover_image_attachment_id":0,"is_favorite":true,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":null}]},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}},{"id":39,"title":"task #39","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"0001-01-01T00:00:00Z","reminders":null,"project_id":25,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"#0","index":0,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}}]`)
			})
			// should equal duedate asc
			t.Run("by due_date", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"due_date"}}, nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":6,"title":"task #6 lower due date","description":"This has something unique","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"2018-11-30T22:25:24Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"test1-6","index":6,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}},{"id":5,"title":"task #5 higher due date","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"2018-12-01T03:58:44Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"test1-5","index":5,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}}`)
			})
			t.Run("by duedate desc", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"due_date"}, "order_by": []string{"desc"}}, nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":5,"title":"task #5 higher due date","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"2018-12-01T03:58:44Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"test1-5","index":5,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}},{"id":6,"title":"task #6 lower due date`)
			})
			t.Run("by duedate asc", func(t *testing.T) {
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort_by": []string{"due_date"}, "order_by": []string{"asc"}}, nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":6,"title":"task #6 lower due date","description":"This has something unique","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"2018-11-30T22:25:24Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"test1-6","index":6,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}},{"id":5,"title":"task #5 higher due date","description":"","done":false,"done_at":"0001-01-01T00:00:00Z","due_date":"2018-12-01T03:58:44Z","reminders":null,"project_id":1,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":"0001-01-01T00:00:00Z","end_date":"0001-01-01T00:00:00Z","assignees":null,"labels":null,"hex_color":"","percent_done":0,"identifier":"test1-5","index":5,"related_tasks":{},"attachments":null,"cover_image_attachment_id":0,"is_favorite":false,"created":"2018-12-01T01:12:04Z","updated":"2018-12-01T01:12:04Z","bucket_id":0,"position":0,"reactions":null,"created_by":{"id":1,"name":"","username":"user1","created":"2018-12-01T15:13:12Z","updated":"2018-12-02T15:13:12Z"}}`)
			})
			t.Run("invalid parameter", func(t *testing.T) {
				// Invalid parameter should not sort at all
				rec, err := testHandler.testReadAllWithUser(url.Values{"sort": []string{"loremipsum"}}, nil)
				require.NoError(t, err)
				assert.NotContains(t, rec.Body.String(), `[{"id":3,"title":"task #3 high prio","description":"","done":false,"due_date":0,"reminders":null,"repeat_after":0,"repeat_mode":0,"priority":100,"start_date":0,"end_date":0,"assignees":null,"labels":null,"hex_color":"","created":1543626724,"updated":1543626724,"created_by":{"id":0,"name":"","username":"","email":"","created":0,"updated":0}},{"id":4,"title":"task #4 low prio","description":"","done":false,"due_date":0,"repeat_after":0,"repeat_mode":0,"priority":1`)
				assert.NotContains(t, rec.Body.String(), `{"id":4,"title":"task #4 low prio","description":"","done":false,"due_date":0,"reminders":null,"repeat_after":0,"repeat_mode":0,"priority":1,"start_date":0,"end_date":0,"assignees":null,"labels":null,"hex_color":"","created":1543626724,"updated":1543626724,"created_by":{"id":0,"name":"","username":"","email":"","created":0,"updated":0}},{"id":3,"title":"task #3 high prio","description":"","done":false,"due_date":0,"repeat_after":0,"repeat_mode":0,"priority":100,"start_date":0,"end_date":0,"assignees":null,"labels":null,"created":1543626724,"updated":1543626724,"created_by":{"id":0,"name":"","username":"","email":"","created":0,"updated":0}}]`)
				assert.NotContains(t, rec.Body.String(), `[{"id":5,"title":"task #5 higher due date","description":"","done":false,"due_date":1543636724,"reminders":null,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":0,"end_date":0,"assignees":null,"labels":null,"hex_color":"","created":1543626724,"updated":1543626724,"created_by":{"id":0,"name":"","username":"","email":"","created":0,"updated":0}},{"id":6,"title":"task #6 lower due date"`)
				assert.NotContains(t, rec.Body.String(), `{"id":6,"title":"task #6 lower due date","description":"","done":false,"due_date":1543616724,"reminders":null,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":0,"end_date":0,"assignees":null,"labels":null,"hex_color":"","created":1543626724,"updated":1543626724,"created_by":{"id":0,"name":"","username":"","email":"","created":0,"updated":0}},{"id":5,"title":"task #5 higher due date","description":"","done":false,"due_date":1543636724,"repeat_after":0,"repeat_mode":0,"priority":0,"start_date":0,"end_date":0,"assignees":null,"labels":null,"created":1543626724,"updated":1543626724,"created_by":{"id":0,"name":"","username":"","email":"","created":0,"updated":0}}]`)
			})
		})
		t.Run("Filter", func(t *testing.T) {
			t.Run("Date range", func(t *testing.T) {
				t.Run("start and end date", func(t *testing.T) {
					rec, err := testHandler.testReadAllWithUser(
						url.Values{
							"filter": []string{"start_date > '2018-12-11T03:46:40+00:00' || end_date < '2018-12-13T11:20:01+00:00' || due_date > '2018-11-29T14:00:00+00:00'"},
						},
						nil,
					)
					require.NoError(t, err)
					assert.NotContains(t, rec.Body.String(), `task #1`)
					assert.NotContains(t, rec.Body.String(), `task #2 `)
					assert.NotContains(t, rec.Body.String(), `task #3 `)
					assert.NotContains(t, rec.Body.String(), `task #4 `)
					assert.Contains(t, rec.Body.String(), `task #5 `)
					assert.Contains(t, rec.Body.String(), `task #6 `)
					assert.Contains(t, rec.Body.String(), `task #7 `)
					assert.Contains(t, rec.Body.String(), `task #8 `)
					assert.Contains(t, rec.Body.String(), `task #9 `)
					assert.NotContains(t, rec.Body.String(), `task #10`)
					assert.NotContains(t, rec.Body.String(), `task #11`)
					assert.NotContains(t, rec.Body.String(), `task #12`)
					assert.NotContains(t, rec.Body.String(), `task #13`)
					assert.NotContains(t, rec.Body.String(), `task #14`)
				})
				t.Run("start date only", func(t *testing.T) {
					rec, err := testHandler.testReadAllWithUser(
						url.Values{
							"filter": []string{"start_date > '2018-10-20T01:46:40+00:00'"},
						},
						nil,
					)
					require.NoError(t, err)
					assert.NotContains(t, rec.Body.String(), `task #1`)
					assert.NotContains(t, rec.Body.String(), `task #2 `)
					assert.NotContains(t, rec.Body.String(), `task #3 `)
					assert.NotContains(t, rec.Body.String(), `task #4 `)
					assert.NotContains(t, rec.Body.String(), `task #5 `)
					assert.NotContains(t, rec.Body.String(), `task #6 `)
					assert.Contains(t, rec.Body.String(), `task #7 `)
					assert.NotContains(t, rec.Body.String(), `task #8 `)
					assert.Contains(t, rec.Body.String(), `task #9 `)
					assert.NotContains(t, rec.Body.String(), `task #10`)
					assert.NotContains(t, rec.Body.String(), `task #11`)
					assert.NotContains(t, rec.Body.String(), `task #12`)
					assert.NotContains(t, rec.Body.String(), `task #13`)
					assert.NotContains(t, rec.Body.String(), `task #14`)
				})
				t.Run("end date only", func(t *testing.T) {
					rec, err := testHandler.testReadAllWithUser(
						url.Values{
							"filter": []string{"end_date > '2018-12-13T11:20:01+00:00'"},
						},
						nil,
					)
					require.NoError(t, err)
					// If no start date but an end date is specified, this should be null
					// since we don't have any tasks in the fixtures with an end date >
					// the current date.
					assert.Equal(t, "[]\n", rec.Body.String())
				})
			})
			t.Run("invalid date", func(t *testing.T) {
				_, err := testHandler.testReadAllWithUser(
					url.Values{
						"filter": []string{"due_date > invalid"},
					},
					nil,
				)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeInvalidTaskFilterValue)
			})
		})
	})

}
