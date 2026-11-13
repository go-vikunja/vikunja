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
	"encoding/json"
	"os"
	"os/exec"
	"slices"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/jobs"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOverdueProducerWebhookEligibility(t *testing.T) {
	cases := []struct {
		name     string
		hooks    []Webhook
		ancestor bool
		singles  []int64
		batch    bool
	}{
		{name: "no matching hook"},
		{name: "one of three projects", hooks: []Webhook{{ProjectID: 3, Events: []string{"task.overdue"}}}, singles: []int64{102}},
		{name: "ancestor subscription", ancestor: true, hooks: []Webhook{{ProjectID: 90001, Events: []string{"task.overdue"}}}, singles: []int64{101}},
		{name: "overlapping subscriptions", ancestor: true, hooks: []Webhook{{ProjectID: 90001, Events: []string{"task.overdue"}}, {ProjectID: 90002, Events: []string{"task.overdue"}}, {UserID: 1, Events: []string{"task.overdue"}}}, singles: []int64{101, 102, 103}},
		{name: "unrelated project", hooks: []Webhook{{ProjectID: 5, Events: []string{"task.overdue"}}}},
		{name: "topic mismatch", hooks: []Webhook{{ProjectID: 1, Events: []string{"task.updated"}}}},
		{name: "other recipient", hooks: []Webhook{{UserID: 2, Events: []string{"task.overdue", "tasks.overdue"}}}},
		{name: "user single", hooks: []Webhook{{UserID: 1, Events: []string{"task.overdue"}}}, singles: []int64{101, 102, 103}},
		{name: "user batch", hooks: []Webhook{{UserID: 1, Events: []string{"tasks.overdue"}}}, batch: true},
		{name: "project batch cannot receive user-only payload", hooks: []Webhook{{ProjectID: 1, Events: []string{"tasks.overdue"}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
			defer events.CleanupPending(s)
			defer jobs.DiscardPending(s)
			_, err := s.Where("id > 0").Delete(new(jobs.Job))
			require.NoError(t, err)
			events.Unfake()
			defer events.Fake()
			RegisterEventForWebhook(&TaskOverdueEvent{})
			RegisterEventForWebhook(&TasksOverdueEvent{})
			ut := eligibilityTestRecipient()
			if tc.ancestor {
				insertTestProject(t, s, &Project{ID: 90001, Title: "parent", OwnerID: 1})
				insertTestProject(t, s, &Project{ID: 90002, Title: "child", OwnerID: 1, ParentProjectID: Ptr(int64(90001))})
				ut.assigned[101].ProjectID = 90002
			}
			for _, hook := range tc.hooks {
				hook.CreatedByID = 1
				hook.TargetURL = "https://example.com"
				_, err := s.Insert(&hook)
				require.NoError(t, err)
			}
			require.NoError(t, dispatchOverdueReminders(s, ut))
			rows := []jobs.Job{}
			require.NoError(t, s.In("kind", []string{"task.overdue.webhook.listener", "tasks.overdue.webhook.listener"}).Find(&rows))
			singles := []int64{}
			batches := 0
			for _, row := range rows {
				assert.Equal(t, "bulk-webhooks", row.Queue)
				if row.Kind == "task.overdue.webhook.listener" {
					var payload TaskOverdueEvent
					require.NoError(t, json.Unmarshal([]byte(row.Payload), &payload))
					require.NotNil(t, payload.Task)
					singles = append(singles, payload.Task.ID)
					assert.Equal(t, int64(1), payload.User.ID)
					assert.Equal(t, payload.Task.ProjectID, payload.Project.ID)
				} else {
					batches++
					var payload TasksOverdueEvent
					require.NoError(t, json.Unmarshal([]byte(row.Payload), &payload))
					assert.Len(t, payload.Tasks, 3)
					assert.Len(t, payload.Projects, 3)
					assert.Equal(t, int64(1), payload.User.ID)
				}
			}
			assert.ElementsMatch(t, tc.singles, singles)
			wantBatches := 0
			if tc.batch {
				wantBatches = 1
			}
			assert.Equal(t, wantBatches, batches)
		})
	}
}

func eligibilityTestRecipient() *userWithTasks {
	return &userWithTasks{user: &user.User{ID: 1}, followed: map[int64]*Task{}, assigned: map[int64]*Task{
		101: {ID: 101, ProjectID: 1}, 102: {ID: 102, ProjectID: 3}, 103: {ID: 103, ProjectID: 4},
	}}
}

func TestOverdueProducerPluginListeners(t *testing.T) {
	// The global listener registry has no unregister; isolate plugin registration.
	mode := os.Getenv("VIKUNJA_TEST_OVERDUE_PLUGIN")
	if mode == "" {
		executable, err := os.Executable()
		require.NoError(t, err)
		for _, mode := range []string{"single", "batch", "both"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.CommandContext(t.Context(), executable, "-test.run=^TestOverdueProducerPluginListeners$")
				cmd.Env = append(os.Environ(), "VIKUNJA_TEST_OVERDUE_PLUGIN="+mode)
				output, err := cmd.CombinedOutput()
				require.NoError(t, err, "%s", output)
			})
		}
		return
	}

	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	defer events.CleanupPending(s)
	events.ClearDispatchedEvents()
	wantSingles, wantBatch := 0, 0
	if mode == "single" || mode == "both" {
		events.RegisterListener("task.overdue", &testOverdueListener{})
		wantSingles = 3
	}
	if mode == "batch" || mode == "both" {
		events.RegisterListener("tasks.overdue", &testOverdueListener{})
		wantBatch = 1
	}
	require.NoError(t, dispatchOverdueReminders(s, eligibilityTestRecipient()))
	require.NoError(t, s.Commit())
	events.DispatchPending(t.Context(), s)
	assert.Len(t, events.GetDispatchedEvents("task.overdue"), wantSingles)
	assert.Len(t, events.GetDispatchedEvents("tasks.overdue"), wantBatch)
}

func TestOverdueProducerManyProjects(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	ut := &userWithTasks{user: &user.User{ID: 1}, assigned: map[int64]*Task{}}
	for id := int64(100000); id < 132770; id++ {
		ut.assigned[id] = &Task{ID: id, ProjectID: id}
	}
	ancestors := make([]ProjectAncestor, 1001)
	for i := range ancestors {
		ancestors[i] = ProjectAncestor{ProjectID: 100000 + int64(i), AncestorID: 200000 + int64(i), Depth: 1}
	}
	for batch := range slices.Chunk(ancestors, 200) {
		_, err := s.Insert(batch)
		require.NoError(t, err)
	}
	_, err := s.Insert(&Webhook{ProjectID: 201000, CreatedByID: 1, TargetURL: "https://example.com", Events: []string{"task.overdue"}})
	require.NoError(t, err)
	singles, batch, err := overdueEventsForRecipient(s, ut)
	require.NoError(t, err)
	require.Len(t, singles, 1)
	assert.Equal(t, int64(101000), singles[0].ID)
	assert.False(t, batch)
}

func TestOverdueProducerEmptyProjectIDs(t *testing.T) {
	assert.Equal(t, "NULL", overdueProjectIDList(nil))
	assert.Equal(t, "-1,0,9223372036854775807", overdueProjectIDList([]int64{-1, 0, 9223372036854775807}))
}
