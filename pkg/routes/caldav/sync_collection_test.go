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

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncTokenPurgeGeneration(t *testing.T) {
	updated := time.Unix(1_700_000_000, 0).UTC()
	purged := updated.Add(time.Hour)
	rr := VikunjaProjectResourceAdapter{
		project: &models.ProjectWithTasksAndBuckets{
			Project: models.Project{
				ID:           36,
				Updated:      updated,
				LastPurgedAt: purged,
			},
		},
	}
	token := syncTokenPrefix + rr.CalculateEtag()
	projectID, timestamp, generation, ok := parseSyncToken(token)
	require.True(t, ok)
	assert.Equal(t, int64(36), projectID)
	assert.Equal(t, updated, timestamp)
	assert.Equal(t, purged.Unix(), generation)

	_, _, generation, ok = parseSyncToken(`data:,"36-1700000000"`)
	require.True(t, ok)
	assert.Zero(t, generation)

	for _, token := range []string{
		`data:,"36-1700000000;purge=abc"`,
		`data:,"36-1700000000;purge=0"`,
		`data:,"36-1700000000;purge=-1"`,
	} {
		_, _, _, ok := parseSyncToken(token)
		assert.False(t, ok)
	}
}
