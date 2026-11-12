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

package jobs

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func TestQueueReplaySelectionAndDryRun(t *testing.T) {
	e := queueTestEngine(t)
	config.Key("jobs.queues." + t.Name() + ".maxattempts").Set(12)
	now := time.Now().UTC().Truncate(time.Second)
	rows := []Job{
		{Queue: t.Name(), Kind: "delivery", State: StateFailed, FinishedAt: now.Add(-2 * time.Hour)},
		{Queue: t.Name(), Kind: "delivery", State: StateFailed, FinishedAt: now.Add(-90 * time.Minute)},
		{Queue: t.Name(), Kind: "delivery", State: StateFailed, FinishedAt: now.Add(-time.Hour)},
		{Queue: t.Name(), Kind: "different", State: StateFailed, FinishedAt: now.Add(-90 * time.Minute)},
		{Queue: t.Name() + "-other", Kind: "delivery", State: StateFailed, FinishedAt: now.Add(-90 * time.Minute)},
		{Queue: t.Name(), Kind: "delivery", State: StateAvailable, FinishedAt: now.Add(-90 * time.Minute)},
		{Queue: t.Name(), Kind: "delivery", State: StateFailed},
	}
	session := e.NewSession()
	defer session.Close()
	ids := make([]int64, 0, len(rows))
	for i := range rows {
		rows[i].Payload = "{}"
		rows[i].RunAt = now
		rows[i].Attempt = 5
		rows[i].MaxAttempts = 5
		_, err := session.Insert(&rows[i])
		require.NoError(t, err)
		ids = append(ids, rows[i].ID)
	}
	t.Cleanup(func() { _, err := e.Where(builder.In("id", ids)).Delete(new(Job)); require.NoError(t, err) })
	queue := NewQueue(e, "operator")
	filter := ReplayFilter{
		Queue:        t.Name(),
		Kind:         "delivery",
		ParkedAfter:  now.Add(-2 * time.Hour),
		ParkedBefore: now.Add(-time.Hour),
		Limit:        1,
	}
	preview, err := queue.ReplaySelected(t.Context(), filter, true)
	require.NoError(t, err)
	assert.EqualValues(t, 2, preview.Matched)
	assert.Zero(t, preview.Replayed)
	fractional := filter
	fractional.ParkedAfter = fractional.ParkedAfter.Add(500 * time.Millisecond)
	fractionalPreview, err := queue.ReplaySelected(t.Context(), fractional, true)
	require.NoError(t, err)
	assert.EqualValues(t, 1, fractionalPreview.Matched, "fractional bounds must not truncate into the previous second")
	var unchanged Job
	_, err = session.ID(rows[0].ID).Get(&unchanged)
	require.NoError(t, err)
	assert.Equal(t, StateFailed, unchanged.State)
	result, err := queue.ReplaySelected(t.Context(), filter, false)
	require.NoError(t, err)
	assert.EqualValues(t, 2, result.Matched)
	assert.EqualValues(t, 1, result.Replayed)
	var replayed Job
	_, err = session.ID(rows[0].ID).Get(&replayed)
	require.NoError(t, err)
	assert.Equal(t, StateAvailable, replayed.State)
	assert.Equal(t, 5, replayed.Attempt)
	assert.Equal(t, 17, replayed.MaxAttempts)
	assert.True(t, replayed.FinishedAt.IsZero())
	result, err = queue.ReplaySelected(t.Context(), filter, false)
	require.NoError(t, err)
	assert.EqualValues(t, 1, result.Matched)
	assert.EqualValues(t, 1, result.Replayed)
	result, err = queue.ReplaySelected(t.Context(), filter, false)
	require.NoError(t, err)
	assert.Zero(t, result.Matched)
	assert.Zero(t, result.Replayed)
}

func TestQueueReplayRejectsUnboundedOrInvalidSelectors(t *testing.T) {
	now := time.Now()
	for _, filter := range []ReplayFilter{
		{},
		{ID: -1},
		{Queue: "mail", Limit: 1001},
		{Queue: "mail", Limit: -1},
		{ID: 1, Queue: "mail"},
		{Queue: "mail", ParkedAfter: now, ParkedBefore: now.Add(-time.Second)},
	} {
		require.Error(t, filter.Validate())
	}
	for _, filter := range []ReplayFilter{
		{ID: 1},
		{Queue: "mail"},
		{Kind: "delivery"},
		{ParkedAfter: now.Add(-time.Hour), ParkedBefore: now},
	} {
		assert.NoError(t, filter.Validate())
	}
}
