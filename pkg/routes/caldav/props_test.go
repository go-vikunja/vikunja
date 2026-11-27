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
	"encoding/xml"
	"net/http"
	"strconv"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPrincipal = "/dav/principals/user15/"

var testCtx = propContext{principal: testPrincipal}

func TestCollectionPropsNotLoaded(t *testing.T) {
	project := &models.ProjectWithTasksAndBuckets{Project: models.Project{ID: 36, Title: "A & B"}}

	t.Run("collection listed in the home set has no calendar-data", func(t *testing.T) {
		r := &collectionResource{href: "/dav/projects/36", project: project, owner: "owner"}
		_, found := collectionProps.resolve(testCtx, r, propCalendarData)
		assert.False(t, found)
	})

	t.Run("collection whose owner is gone", func(t *testing.T) {
		r := &collectionResource{href: "/dav/projects/36", project: project}
		_, found := collectionProps.resolve(testCtx, r, propOwner)
		assert.False(t, found)
	})
}

func TestLastModifiedIsUTC(t *testing.T) {
	updated := time.Date(2024, 6, 15, 23, 30, 0, 0, time.FixedZone("UTC+2", 2*60*60))
	task := &taskResource{task: &models.Task{Updated: updated}}

	value, found := taskProps.resolve(propContext{}, task, propGetLastModified)
	assert.True(t, found)
	assert.Equal(t, textValue("Sat, 15 Jun 2024 21:30:00 GMT"), value)
}

func TestTaskContentIsRenderedOnce(t *testing.T) {
	task := &models.Task{UID: "uid-rendered-once", Title: "first"}
	r := &taskResource{task: task}

	content, found := taskProps.resolve(testCtx, r, propCalendarData)
	require.True(t, found)
	require.Contains(t, string(content.(textValue)), "SUMMARY:first")

	// A second render would pick up the new title.
	task.Title = "second"
	again, _ := taskProps.resolve(testCtx, r, propCalendarData)
	assert.Equal(t, content, again)
	length, _ := taskProps.resolve(testCtx, r, propGetContentLength)
	assert.Equal(t, textValue(strconv.Itoa(len(content.(textValue)))), length)
}

func TestPropTableAnswer(t *testing.T) {
	task := &models.Task{ID: 40, UID: "uid", Title: "Task", Updated: time.Now()}
	unknown := xml.Name{Space: "urn:example", Local: "x"}

	foundNames := func(r msResponse) map[int][]xml.Name {
		out := map[int][]xml.Name{}
		for _, ps := range r.(propstatResponse).propstats {
			for _, p := range ps.props {
				out[ps.status] = append(out[ps.status], p.name)
			}
		}
		return out
	}

	t.Run("allprop with include", func(t *testing.T) {
		names := foundNames((&taskResource{href: "/t.ics", task: task}).answer(testCtx, allPropRequest{include: []xml.Name{propCalendarData, propGetETag, unknown}}))
		assert.Equal(t, 1, countName(names[http.StatusOK], propCalendarData))
		assert.Equal(t, 1, countName(names[http.StatusOK], propGetETag))
		assert.Equal(t, []xml.Name{unknown}, names[http.StatusNotFound])
	})
}

func countName(names []xml.Name, name xml.Name) int {
	n := 0
	for _, got := range names {
		if got == name {
			n++
		}
	}
	return n
}
