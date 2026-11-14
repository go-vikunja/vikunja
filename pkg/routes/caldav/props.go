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
	"path"
	"slices"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/models"
)

// propContext is what all resources of one request share.
type propContext struct {
	// The requesting user's principal.
	principal string
}

// davResource is a resource whose properties are loaded and can be answered without a query.
type davResource interface {
	answer(ctx propContext, req propfindRequest) msResponse
}

type propResolver[R any] func(ctx propContext, r R) (propValue, bool)

type propDef[R any] struct {
	name    xml.Name
	resolve propResolver[R]
}

// propTable lists the properties a resource kind has; anything else is not found.
type propTable[R any] []propDef[R]

func (t propTable[R]) resolve(ctx propContext, r R, name xml.Name) (propValue, bool) {
	for _, def := range t {
		if def.name == name {
			return def.resolve(ctx, r)
		}
	}
	return nil, false
}

func (t propTable[R]) answer(ctx propContext, r R, href string, req propfindRequest) msResponse {
	named, ok := req.(propRequest)
	if !ok || len(named.names) == 0 {
		// Pinned: caldav-go only answers properties asked for by name.
		return hrefOnlyResponse{href: href}
	}

	var found []prop
	var missing []xml.Name
	for _, name := range named.names {
		if value, has := t.resolve(ctx, r, name); has {
			found = append(found, prop{name: name, value: value})
			continue
		}
		// Pinned: caldav-go dropped namespaces it had no prefix for.
		if _, known := nsPrefixes[name.Space]; !known {
			name.Space = ""
		}
		missing = append(missing, name)
	}
	return propResponse(href, found, missing)
}

var (
	propResourceType         = davName("resourcetype")
	propDisplayName          = davName("displayname")
	propCurrentUserPrincipal = davName("current-user-principal")
	propPrincipalURL         = davName("principal-URL")
	propPrincipalCollections = davName("principal-collection-set")
	propOwner                = davName("owner")
	propGetContentType       = davName("getcontenttype")
	propGetContentLength     = davName("getcontentlength")
	propGetETag              = davName("getetag")
	propGetLastModified      = davName("getlastmodified")
	propSyncToken            = davName("sync-token")
	propGetCTag              = xml.Name{Space: nsCalendarServer, Local: "getctag"}
	propCalendarHomeSet      = calDAVName("calendar-home-set")
	propCalendarUserAddress  = calDAVName("calendar-user-address-set")
	propSupportedComponents  = calDAVName("supported-calendar-component-set")
	propCalendarData         = calDAVName("calendar-data")
)

var (
	collectionType = element{name: davName("collection")}
	principalType  = element{name: davName("principal")}
	calendarType   = element{name: calDAVName("calendar")}

	supportedComponents = elementsValue{
		{name: calDAVName("comp"), attrs: []attr{{name: "name", value: "VCALENDAR"}}},
		{name: calDAVName("comp"), attrs: []attr{{name: "name", value: "VTODO"}}},
	}
)

func always[R any](v propValue) propResolver[R] {
	return func(propContext, R) (propValue, bool) { return v, true }
}

func currentUserPrincipal[R any](ctx propContext, _ R) (propValue, bool) {
	return hrefsValue{ctx.principal}, true
}

func textIfSet(s string) (propValue, bool) {
	return textValue(s), s != ""
}

// Pinned: caldav-go formats in the server's zone and labels it GMT.
func modTime(t time.Time) (propValue, bool) {
	if t.IsZero() {
		return nil, false
	}
	return textValue(t.Format(http.TimeFormat)), true
}

func ownerPrincipal(username string) (propValue, bool) {
	if username == "" {
		return nil, false
	}
	return hrefsValue{principalPathForUser(username)}, true
}

// hrefed is any resource that answers some properties with its own href.
type hrefed interface{ ownHref() string }

func ownHref[R hrefed](_ propContext, r R) (propValue, bool) {
	return hrefsValue{r.ownHref()}, true
}

// Pinned: caldav-go answers these with the href of whatever resource was asked.
func selfReferencing[R hrefed]() propTable[R] {
	return propTable[R]{
		{propPrincipalCollections, ownHref[R]},
		{propCalendarUserAddress, ownHref[R]},
	}
}

// RFC 4791 §6.2.1 and RFC 4918 §16: these belong to the principal; /dav/ answers them too
// because clients start discovery there.
func principalDiscovery[R any]() propTable[R] {
	return propTable[R]{
		{propCalendarHomeSet, always[R](hrefsValue{ProjectHomeSetPath})},
		{propPrincipalURL, func(ctx propContext, _ R) (propValue, bool) { return hrefsValue{ctx.principal}, true }},
	}
}

func lastSegment[R hrefed](_ propContext, r R) (propValue, bool) {
	return textValue(path.Base(r.ownHref())), true
}

// entryResource is /dav/.
type entryResource struct{ href string }

func (r *entryResource) ownHref() string { return r.href }

var entryProps = append(propTable[*entryResource]{
	{propResourceType, always[*entryResource](elementsValue{collectionType})},
	{propDisplayName, lastSegment[*entryResource]},
	{propCurrentUserPrincipal, currentUserPrincipal[*entryResource]},
	{propSupportedComponents, always[*entryResource](supportedComponents)},
	{propGetContentType, always[*entryResource](textValue(contentTypeCollection))},
}, slices.Concat(principalDiscovery[*entryResource](), selfReferencing[*entryResource]())...)

func (r *entryResource) answer(ctx propContext, req propfindRequest) msResponse {
	return entryProps.answer(ctx, r, r.href, req)
}

// principalResource is the requesting user's principal.
type principalResource struct{ href string }

func (r *principalResource) ownHref() string { return r.href }

var principalProps = append(propTable[*principalResource]{
	{propResourceType, always[*principalResource](elementsValue{collectionType, principalType})},
	{propDisplayName, lastSegment[*principalResource]},
	{propCurrentUserPrincipal, currentUserPrincipal[*principalResource]},
	{propSupportedComponents, always[*principalResource](supportedComponents)},
	{propGetContentType, always[*principalResource](textValue(contentTypeCollection))},
}, slices.Concat(principalDiscovery[*principalResource](), selfReferencing[*principalResource]())...)

func (r *principalResource) answer(ctx propContext, req propfindRequest) msResponse {
	return principalProps.answer(ctx, r, r.href, req)
}

// homeSetResource is /dav/projects.
type homeSetResource struct{ href string }

func (r *homeSetResource) ownHref() string { return r.href }

var homeSetProps = append(propTable[*homeSetResource]{
	{propResourceType, always[*homeSetResource](elementsValue{collectionType})},
	{propDisplayName, lastSegment[*homeSetResource]},
	{propCurrentUserPrincipal, currentUserPrincipal[*homeSetResource]},
	{propSupportedComponents, always[*homeSetResource](supportedComponents)},
	{propGetContentType, always[*homeSetResource](textValue(contentTypeCollection))},
}, selfReferencing[*homeSetResource]()...)

func (r *homeSetResource) answer(ctx propContext, req propfindRequest) msResponse {
	return homeSetProps.answer(ctx, r, r.href, req)
}

// collectionResource is a project, the Favorites pseudo project or a saved filter.
type collectionResource struct {
	href    string
	project *models.ProjectWithTasksAndBuckets
	// nil when the tasks were not loaded or there are none.
	tasks []*models.TaskWithComments
	// Username; empty if the owner no longer exists.
	owner string
}

func (r *collectionResource) ownHref() string { return r.href }

func (r *collectionResource) etag() string { return collectionETag(r.project, r.tasks) }

var collectionProps = append(propTable[*collectionResource]{
	{propResourceType, always[*collectionResource](elementsValue{collectionType, calendarType})},
	{propDisplayName, func(_ propContext, r *collectionResource) (propValue, bool) {
		return textValue(r.project.Title), true
	}},
	{propCurrentUserPrincipal, currentUserPrincipal[*collectionResource]},
	{propSupportedComponents, always[*collectionResource](supportedComponents)},
	{propGetContentType, always[*collectionResource](textValue(contentTypeCollection))},
	{propOwner, func(_ propContext, r *collectionResource) (propValue, bool) { return ownerPrincipal(r.owner) }},
	{propGetETag, func(_ propContext, r *collectionResource) (propValue, bool) { return textValue(r.etag()), true }},
	{propGetCTag, func(_ propContext, r *collectionResource) (propValue, bool) { return textValue(r.etag()), true }},
	{propSyncToken, func(_ propContext, r *collectionResource) (propValue, bool) {
		return textValue(syncTokenPrefix + r.etag()), true
	}},
	{propGetLastModified, func(_ propContext, r *collectionResource) (propValue, bool) { return modTime(r.project.Updated) }},
	{propCalendarData, func(_ propContext, r *collectionResource) (propValue, bool) {
		if r.tasks == nil {
			return nil, false
		}
		return textIfSet(renderCollection(r.project, r.tasks))
	}},
}, selfReferencing[*collectionResource]()...)

func (r *collectionResource) answer(ctx propContext, req propfindRequest) msResponse {
	return collectionProps.answer(ctx, r, r.href, req)
}

// taskResource is a task as a member of the collection its href names.
type taskResource struct {
	href string
	task *models.Task
	// The collection's owner.
	owner string
	// Rendering costs as much as a GET, so it happens at most once per resource.
	rendered *string
}

func (r *taskResource) ownHref() string { return r.href }

func (r *taskResource) content() string {
	if r.rendered == nil {
		content := renderTask(r.task)
		r.rendered = &content
	}
	return *r.rendered
}

var taskProps = append(propTable[*taskResource]{
	{propResourceType, always[*taskResource](nil)},
	{propDisplayName, func(_ propContext, r *taskResource) (propValue, bool) { return textValue(r.task.Title), true }},
	{propCurrentUserPrincipal, currentUserPrincipal[*taskResource]},
	{propGetContentType, always[*taskResource](textValue(contentTypeTask))},
	{propOwner, func(_ propContext, r *taskResource) (propValue, bool) { return ownerPrincipal(r.owner) }},
	{propGetETag, func(_ propContext, r *taskResource) (propValue, bool) { return textValue(taskETag(r.task)), true }},
	{propGetCTag, func(_ propContext, r *taskResource) (propValue, bool) { return textValue(taskETag(r.task)), true }},
	{propGetLastModified, func(_ propContext, r *taskResource) (propValue, bool) { return modTime(r.task.Updated) }},
	{propGetContentLength, func(_ propContext, r *taskResource) (propValue, bool) {
		return textValue(strconv.Itoa(len(r.content()))), true
	}},
	{propCalendarData, func(_ propContext, r *taskResource) (propValue, bool) { return textIfSet(r.content()) }},
}, selfReferencing[*taskResource]()...)

func (r *taskResource) answer(ctx propContext, req propfindRequest) msResponse {
	return taskProps.answer(ctx, r, r.href, req)
}
