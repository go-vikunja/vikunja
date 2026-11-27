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
	"net/url"
	"path"
	"slices"
	"strconv"
	"time"

	caldav2 "code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/models"
)

type propContext struct {
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
	switch req := req.(type) {
	case propRequest:
		found, missing := t.resolveAll(ctx, r, req.names)
		return propResponse(href, found, missing)
	case allPropRequest:
		var found []prop
		answered := map[xml.Name]bool{}
		for _, def := range t {
			if explicitOnly[def.name] {
				continue
			}
			if value, has := def.resolve(ctx, r); has {
				found = append(found, prop{
					name:  def.name,
					value: value,
				})
				answered[def.name] = true
			}
		}
		included, missing := t.resolveAll(ctx, r, slices.DeleteFunc(slices.Clone(req.include), func(n xml.Name) bool { return answered[n] }))
		return propResponse(href, append(found, included...), missing)
	case propNameRequest:
		var names []prop
		for _, def := range t {
			if explicitOnly[def.name] {
				continue
			}
			if _, has := def.resolve(ctx, r); has {
				names = append(names, prop{name: def.name})
			}
		}
		return propResponse(href, names, nil)
	}
	return propResponse(href, nil, nil)
}

func (t propTable[R]) resolveAll(ctx propContext, r R, names []xml.Name) (found []prop, missing []xml.Name) {
	for _, name := range names {
		if value, has := t.resolve(ctx, r, name); has {
			found = append(found, prop{
				name:  name,
				value: value,
			})
			continue
		}
		missing = append(missing, name)
	}
	return found, missing
}

// Not owed to allprop (RFC 4918 §9.1, RFC 4791 §9.6); it would send every task body on Depth: 1.
var explicitOnly = map[xml.Name]bool{propCalendarData: true}

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
	propGetCTag              = xml.Name{
		Space: nsCalendarServer,
		Local: "getctag",
	}
	propCalendarHomeSet     = calDAVName("calendar-home-set")
	propCalendarUserAddress = calDAVName("calendar-user-address-set")
	propSupportedComponents = calDAVName("supported-calendar-component-set")
	propCalendarData        = calDAVName("calendar-data")
)

var (
	collectionType = element{name: davName("collection")}
	principalType  = element{name: davName("principal")}
	calendarType   = element{name: calDAVName("calendar")}

	supportedComponents = elementsValue{
		{
			name: calDAVName("comp"),
			attrs: []attr{
				{
					name:  "name",
					value: "VCALENDAR",
				},
			},
		},
		{
			name: calDAVName("comp"),
			attrs: []attr{
				{
					name:  "name",
					value: "VTODO",
				},
			},
		},
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

func modTime(t time.Time) (propValue, bool) {
	if t.IsZero() {
		return nil, false
	}
	return textValue(lastModified(t)), true
}

func ownerPrincipal(username string) (propValue, bool) {
	if username == "" {
		return nil, false
	}
	return hrefsValue{principalPathForUser(username)}, true
}

// hrefed is a resource whose displayname is the last segment of its href.
type hrefed interface{ ownHref() string }

func principalCollections[R any]() propDef[R] {
	return propDef[R]{
		name:    propPrincipalCollections,
		resolve: always[R](hrefsValue{PrincipalBasePath + "/"}),
	}
}

// Without scheduling there is no mailto: address to give, so the principal is the address.
func calendarUserAddress[R any]() propDef[R] {
	return propDef[R]{
		name:    propCalendarUserAddress,
		resolve: currentUserPrincipal[R],
	}
}

// RFC 4791 §6.2.1, RFC 3744 §4.2: on /dav/ and the home set too, where clients start discovery (#2417).
func principalDiscovery[R any]() propTable[R] {
	return propTable[R]{
		// No trailing slash: existing clients stored it that way.
		{
			name:    propCalendarHomeSet,
			resolve: always[R](hrefsValue{ProjectBasePath}),
		},
		{
			name:    propPrincipalURL,
			resolve: currentUserPrincipal[R],
		},
	}
}

func lastSegment[R hrefed](_ propContext, r R) (propValue, bool) {
	name, err := url.PathUnescape(path.Base(r.ownHref()))
	return textValue(name), err == nil
}

// entryResource is /dav/.
type entryResource struct{ href string }

func (r *entryResource) ownHref() string { return r.href }

var entryProps = append(propTable[*entryResource]{
	{
		name:    propResourceType,
		resolve: always[*entryResource](elementsValue{collectionType}),
	},
	{
		name:    propDisplayName,
		resolve: lastSegment[*entryResource],
	},
	{
		name:    propCurrentUserPrincipal,
		resolve: currentUserPrincipal[*entryResource],
	},
	{
		name:    propSupportedComponents,
		resolve: always[*entryResource](supportedComponents),
	},
	{
		name:    propGetContentType,
		resolve: always[*entryResource](textValue(contentTypeCollection)),
	},
	principalCollections[*entryResource](),
	calendarUserAddress[*entryResource](),
}, principalDiscovery[*entryResource]()...)

func (r *entryResource) answer(ctx propContext, req propfindRequest) msResponse {
	return entryProps.answer(ctx, r, r.href, req)
}

type principalResource struct{ href string }

func (r *principalResource) ownHref() string { return r.href }

var principalProps = append(propTable[*principalResource]{
	{
		name: propResourceType,
		resolve: always[*principalResource](elementsValue{
			collectionType,
			principalType,
		}),
	},
	{
		name:    propDisplayName,
		resolve: lastSegment[*principalResource],
	},
	{
		name:    propCurrentUserPrincipal,
		resolve: currentUserPrincipal[*principalResource],
	},
	{
		name:    propSupportedComponents,
		resolve: always[*principalResource](supportedComponents),
	},
	{
		name:    propGetContentType,
		resolve: always[*principalResource](textValue(contentTypeCollection)),
	},
	principalCollections[*principalResource](),
	calendarUserAddress[*principalResource](),
}, principalDiscovery[*principalResource]()...)

func (r *principalResource) answer(ctx propContext, req propfindRequest) msResponse {
	return principalProps.answer(ctx, r, r.href, req)
}

// homeSetResource is /dav/projects.
type homeSetResource struct{ href string }

func (r *homeSetResource) ownHref() string { return r.href }

var homeSetProps = append(propTable[*homeSetResource]{
	{
		name:    propResourceType,
		resolve: always[*homeSetResource](elementsValue{collectionType}),
	},
	{
		name:    propDisplayName,
		resolve: lastSegment[*homeSetResource],
	},
	{
		name:    propCurrentUserPrincipal,
		resolve: currentUserPrincipal[*homeSetResource],
	},
	{
		name:    propSupportedComponents,
		resolve: always[*homeSetResource](supportedComponents),
	},
	{
		name:    propGetContentType,
		resolve: always[*homeSetResource](textValue(contentTypeCollection)),
	},
	principalCollections[*homeSetResource](),
}, principalDiscovery[*homeSetResource]()...)

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

func (r *collectionResource) etag() string { return collectionETag(r.project, r.tasks) }

var collectionProps = propTable[*collectionResource]{
	{
		name: propResourceType,
		resolve: always[*collectionResource](elementsValue{
			collectionType,
			calendarType,
		}),
	},
	{
		name: propDisplayName,
		resolve: func(_ propContext, r *collectionResource) (propValue, bool) {
			return textValue(r.project.Title), true
		},
	},
	{
		name: calendarColorProp,
		resolve: func(_ propContext, r *collectionResource) (propValue, bool) {
			return textIfSet(caldav2.AppleColor(r.project.HexColor))
		},
	},
	{
		name:    propCurrentUserPrincipal,
		resolve: currentUserPrincipal[*collectionResource],
	},
	{
		name:    propSupportedComponents,
		resolve: always[*collectionResource](supportedComponents),
	},
	{
		name:    propGetContentType,
		resolve: always[*collectionResource](textValue(contentTypeCollection)),
	},
	{
		name:    propOwner,
		resolve: func(_ propContext, r *collectionResource) (propValue, bool) { return ownerPrincipal(r.owner) },
	},
	{
		name:    propGetETag,
		resolve: func(_ propContext, r *collectionResource) (propValue, bool) { return textValue(r.etag()), true },
	},
	{
		name:    propGetCTag,
		resolve: func(_ propContext, r *collectionResource) (propValue, bool) { return textValue(r.etag()), true },
	},
	{
		name: propSyncToken,
		resolve: func(_ propContext, r *collectionResource) (propValue, bool) {
			return textValue(syncTokenPrefix + r.etag()), true
		},
	},
	{
		name:    propGetLastModified,
		resolve: func(_ propContext, r *collectionResource) (propValue, bool) { return modTime(r.project.Updated) },
	},
	{
		name: propCalendarData,
		resolve: func(_ propContext, r *collectionResource) (propValue, bool) {
			if r.tasks == nil {
				return nil, false
			}
			return textIfSet(caldav2.GetCaldavTodosForTasks(r.project, r.tasks))
		},
	},
	principalCollections[*collectionResource](),
}

func (r *collectionResource) answer(ctx propContext, req propfindRequest) msResponse {
	return collectionProps.answer(ctx, r, r.href, req)
}

// taskResource is a task as a member of the collection its href names.
type taskResource struct {
	href  string
	task  *models.Task
	owner string
	// Rendering costs as much as a GET, so it happens at most once per resource.
	rendered *string
}

func (r *taskResource) content() string {
	if r.rendered == nil {
		content := renderTask(r.task)
		r.rendered = &content
	}
	return *r.rendered
}

var taskProps = propTable[*taskResource]{
	{
		name:    propResourceType,
		resolve: always[*taskResource](nil),
	},
	{
		name:    propDisplayName,
		resolve: func(_ propContext, r *taskResource) (propValue, bool) { return textValue(r.task.Title), true },
	},
	{
		name:    propCurrentUserPrincipal,
		resolve: currentUserPrincipal[*taskResource],
	},
	{
		name:    propGetContentType,
		resolve: always[*taskResource](textValue(contentTypeTask)),
	},
	{
		name:    propOwner,
		resolve: func(_ propContext, r *taskResource) (propValue, bool) { return ownerPrincipal(r.owner) },
	},
	{
		name:    propGetETag,
		resolve: func(_ propContext, r *taskResource) (propValue, bool) { return textValue(taskETag(r.task)), true },
	},
	{
		name:    propGetCTag,
		resolve: func(_ propContext, r *taskResource) (propValue, bool) { return textValue(taskETag(r.task)), true },
	},
	{
		name:    propGetLastModified,
		resolve: func(_ propContext, r *taskResource) (propValue, bool) { return modTime(r.task.Updated) },
	},
	{
		name: propGetContentLength,
		resolve: func(_ propContext, r *taskResource) (propValue, bool) {
			return textValue(strconv.Itoa(len(r.content()))), true
		},
	},
	{
		name:    propCalendarData,
		resolve: func(_ propContext, r *taskResource) (propValue, bool) { return textIfSet(r.content()) },
	},
	principalCollections[*taskResource](),
}

func (r *taskResource) answer(ctx propContext, req propfindRequest) msResponse {
	return taskProps.answer(ctx, r, r.href, req)
}
