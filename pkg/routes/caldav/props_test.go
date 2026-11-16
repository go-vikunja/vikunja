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

// absent marks a property the resource must not have.
var absent propValue = textValue("\x00absent")

func assertProps(t *testing.T, resolve func(xml.Name) (propValue, bool), want map[xml.Name]propValue) {
	t.Helper()
	for name, wantValue := range want {
		value, found := resolve(name)
		if wantValue == absent {
			assert.False(t, found, "%v must not be found, got %v", name, value)
			continue
		}
		if assert.True(t, found, "%v must be found", name) {
			assert.Equal(t, wantValue, value, "%v", name)
		}
	}
}

func TestPropertyResolution(t *testing.T) {
	updated := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	project := &models.ProjectWithTasksAndBuckets{Project: models.Project{ID: 36, Title: "A & B", Updated: updated}}
	task := &models.Task{ID: 40, UID: "uid", Title: "Task <1>", Updated: updated}
	tasks := []*models.TaskWithComments{{Task: *task}}

	// Every kind is asked for every property, so a missing row is an untested property.
	t.Run("entry", func(t *testing.T) {
		r := &entryResource{href: "/dav/"}
		assertProps(t, func(n xml.Name) (propValue, bool) { return entryProps.resolve(testCtx, r, n) }, map[xml.Name]propValue{
			propResourceType:         elementsValue{collectionType},
			propDisplayName:          textValue("dav"),
			propCurrentUserPrincipal: hrefsValue{testPrincipal},
			propCalendarHomeSet:      hrefsValue{ProjectHomeSetPath},
			propPrincipalURL:         hrefsValue{testPrincipal},
			propPrincipalCollections: hrefsValue{"/dav/"},
			propCalendarUserAddress:  hrefsValue{"/dav/"},
			propSupportedComponents:  supportedComponents,
			propGetContentType:       textValue(contentTypeCollection),
			propOwner:                absent,
			propGetETag:              absent,
			propGetCTag:              absent,
			propSyncToken:            absent,
			propGetLastModified:      absent,
			propGetContentLength:     absent,
			propCalendarData:         absent,
		})
	})

	t.Run("principal", func(t *testing.T) {
		r := &principalResource{href: testPrincipal}
		assertProps(t, func(n xml.Name) (propValue, bool) { return principalProps.resolve(testCtx, r, n) }, map[xml.Name]propValue{
			propResourceType:         elementsValue{collectionType, principalType},
			propDisplayName:          textValue("user15"),
			propCurrentUserPrincipal: hrefsValue{testPrincipal},
			propCalendarHomeSet:      hrefsValue{ProjectHomeSetPath},
			propPrincipalURL:         hrefsValue{testPrincipal},
			propPrincipalCollections: hrefsValue{testPrincipal},
			propCalendarUserAddress:  hrefsValue{testPrincipal},
			propSupportedComponents:  supportedComponents,
			propGetContentType:       textValue(contentTypeCollection),
			propOwner:                absent,
			propGetETag:              absent,
			propGetCTag:              absent,
			propSyncToken:            absent,
			propGetLastModified:      absent,
			propGetContentLength:     absent,
			propCalendarData:         absent,
		})
	})

	t.Run("home set", func(t *testing.T) {
		r := &homeSetResource{href: ProjectBasePath}
		assertProps(t, func(n xml.Name) (propValue, bool) { return homeSetProps.resolve(testCtx, r, n) }, map[xml.Name]propValue{
			propResourceType:         elementsValue{collectionType},
			propDisplayName:          textValue("projects"),
			propCurrentUserPrincipal: hrefsValue{testPrincipal},
			propCalendarHomeSet:      absent,
			propPrincipalURL:         absent,
			propPrincipalCollections: hrefsValue{ProjectBasePath},
			propCalendarUserAddress:  hrefsValue{ProjectBasePath},
			propSupportedComponents:  supportedComponents,
			propGetContentType:       textValue(contentTypeCollection),
			propOwner:                absent,
			propGetETag:              absent,
			propGetCTag:              absent,
			propSyncToken:            absent,
			propGetLastModified:      absent,
			propGetContentLength:     absent,
			propCalendarData:         absent,
		})
	})

	collectionWant := func(r *collectionResource) map[xml.Name]propValue {
		etag := collectionETag(project, r.tasks)
		return map[xml.Name]propValue{
			propResourceType:         elementsValue{collectionType, calendarType},
			propDisplayName:          textValue("A & B"),
			propCurrentUserPrincipal: hrefsValue{testPrincipal},
			propCalendarHomeSet:      absent,
			propPrincipalURL:         absent,
			propPrincipalCollections: hrefsValue{r.href},
			propCalendarUserAddress:  hrefsValue{r.href},
			propSupportedComponents:  supportedComponents,
			propGetContentType:       textValue(contentTypeCollection),
			propOwner:                hrefsValue{"/dav/principals/owner/"},
			propGetETag:              textValue(etag),
			propGetCTag:              textValue(etag),
			propSyncToken:            textValue("data:," + etag),
			propGetLastModified:      textValue(lastModified(updated)),
			propGetContentLength:     absent,
		}
	}

	t.Run("collection listed in the home set", func(t *testing.T) {
		r := &collectionResource{href: "/dav/projects/36", project: project, owner: "owner"}
		want := collectionWant(r)
		want[propCalendarData] = absent
		assertProps(t, func(n xml.Name) (propValue, bool) { return collectionProps.resolve(testCtx, r, n) }, want)
	})

	t.Run("collection with its tasks", func(t *testing.T) {
		r := &collectionResource{href: "/dav/projects/36", project: project, tasks: tasks, owner: "owner"}
		want := collectionWant(r)
		want[propCalendarData] = textValue(renderCollection(project, tasks))
		assertProps(t, func(n xml.Name) (propValue, bool) { return collectionProps.resolve(testCtx, r, n) }, want)
	})

	t.Run("collection whose owner is gone", func(t *testing.T) {
		r := &collectionResource{href: "/dav/projects/36", project: project}
		_, found := collectionProps.resolve(testCtx, r, propOwner)
		assert.False(t, found)
	})

	t.Run("task", func(t *testing.T) {
		const href = "/dav/projects/36/uid.ics"
		r := &taskResource{href: href, task: task, owner: "owner"}
		content := renderTask(task)
		assertProps(t, func(n xml.Name) (propValue, bool) { return taskProps.resolve(testCtx, r, n) }, map[xml.Name]propValue{
			propResourceType:         nil,
			propDisplayName:          textValue("Task <1>"),
			propCurrentUserPrincipal: hrefsValue{testPrincipal},
			propCalendarHomeSet:      absent,
			propPrincipalURL:         absent,
			propPrincipalCollections: hrefsValue{href},
			propCalendarUserAddress:  hrefsValue{href},
			propSupportedComponents:  absent,
			propGetContentType:       textValue(contentTypeTask),
			propOwner:                hrefsValue{"/dav/principals/owner/"},
			propGetETag:              textValue(taskETag(task)),
			propGetCTag:              textValue(taskETag(task)),
			propSyncToken:            absent,
			propGetLastModified:      textValue(lastModified(updated)),
			propGetContentLength:     textValue(strconv.Itoa(len(content))),
			propCalendarData:         textValue(content),
		})
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

	t.Run("named properties in request order, missing ones with their namespace", func(t *testing.T) {
		r := (&taskResource{href: "/t.ics", task: task}).answer(testCtx, propRequest{names: []xml.Name{propGetETag, unknown, propSyncToken, propDisplayName}})
		assert.Equal(t, map[int][]xml.Name{
			http.StatusOK:       {propGetETag, propDisplayName},
			http.StatusNotFound: {unknown, propSyncToken},
		}, foundNames(r))
	})

	t.Run("allprop leaves calendar-data out and lists nothing as missing", func(t *testing.T) {
		names := foundNames((&taskResource{href: "/t.ics", task: task}).answer(testCtx, allPropRequest{}))
		assert.NotContains(t, names[http.StatusOK], propCalendarData)
		assert.Contains(t, names[http.StatusOK], propGetContentLength)
		assert.Empty(t, names[http.StatusNotFound])
	})

	t.Run("allprop with include", func(t *testing.T) {
		names := foundNames((&taskResource{href: "/t.ics", task: task}).answer(testCtx, allPropRequest{include: []xml.Name{propCalendarData, propGetETag, unknown}}))
		assert.Equal(t, 1, countName(names[http.StatusOK], propCalendarData))
		assert.Equal(t, 1, countName(names[http.StatusOK], propGetETag))
		assert.Equal(t, []xml.Name{unknown}, names[http.StatusNotFound])
	})

	t.Run("propname answers names without values", func(t *testing.T) {
		r := (&taskResource{href: "/t.ics", task: task}).answer(testCtx, propNameRequest{}).(propstatResponse)
		require.Len(t, r.propstats, 1)
		for _, p := range r.propstats[0].props {
			assert.Nil(t, p.value, "%v", p.name)
			assert.NotEqual(t, propCalendarData, p.name)
		}
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
