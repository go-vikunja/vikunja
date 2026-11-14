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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/samedi/caldav-go/data"
	"xorm.io/xorm"
)

const (
	// A multiget names every task it wants, so it is the largest body a client sends.
	maxReportBodySize = 1 << 20
	maxMultigetHrefs  = 5000
	maxReportElements = 2 * maxMultigetHrefs
	maxReportDepth    = 16
)

var errUnsupportedReport = errors.New("caldav: unsupported report")

// reportRequest is what a REPORT body asks for.
type reportRequest interface{ isReportRequest() }

// multigetRequest is a calendar-multiget (RFC 4791 §7.9).
type multigetRequest struct {
	props propfindRequest
	// Without duplicates, in request order.
	hrefs []string
}

// calendarQueryRequest is a calendar-query (RFC 4791 §7.8).
type calendarQueryRequest struct {
	props propfindRequest
	// The filter element as caldav-go's filter parser reads it.
	filterXML string
}

// syncCollectionRequest is a sync-collection (RFC 6578 §3.2).
type syncCollectionRequest struct {
	props propfindRequest
	token string
}

func (multigetRequest) isReportRequest()       {}
func (calendarQueryRequest) isReportRequest()  {}
func (syncCollectionRequest) isReportRequest() {}

func parseReport(body string) (reportRequest, error) {
	var (
		root     xml.Name
		props    propCollector
		hrefs    []string
		seenHref = map[string]bool{}
		token    string
		elements int
		// The href or sync-token being read.
		text *strings.Builder
	)
	err := walkXML(body, func(p []xml.Name, tok xml.Token) error {
		switch t := tok.(type) {
		case xml.StartElement:
			if len(p) == 0 {
				root = t.Name
				return nil
			}
			elements++
			if elements > maxReportElements {
				return fmt.Errorf("%w: more than %d elements", errBodyTooLarge, maxReportElements)
			}
			if len(p) >= maxReportDepth {
				return fmt.Errorf("%w: nested deeper than %d elements", errBodyTooLarge, maxReportDepth)
			}
			if len(p) == 1 && (t.Name == davName("href") || t.Name == davName("sync-token")) {
				text = &strings.Builder{}
			}
			return props.visit(p, t)
		case xml.CharData:
			if len(p) == 2 && text != nil {
				text.Write(t)
			}
		case xml.EndElement:
			if len(p) != 1 || text == nil {
				return nil
			}
			switch t.Name {
			case davName("href"):
				href := text.String()
				if !seenHref[href] {
					if len(hrefs) == maxMultigetHrefs {
						return fmt.Errorf("%w: more than %d hrefs", errBodyTooLarge, maxMultigetHrefs)
					}
					seenHref[href] = true
					hrefs = append(hrefs, href)
				}
			case davName("sync-token"):
				token = strings.TrimSpace(text.String())
			}
			text = nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	switch root {
	case xml.Name{}:
		return nil, fmt.Errorf("%w: empty body", errMalformedBody)
	case calDAVName("calendar-multiget"):
		return multigetRequest{props: props.request(), hrefs: hrefs}, nil
	case calDAVName("calendar-query"):
		return calendarQueryRequest{props: props.request(), filterXML: libraryFilterXML(body)}, nil
	case davName("sync-collection"):
		return syncCollectionRequest{props: props.request(), token: token}, nil
	}
	return nil, errUnsupportedReport
}

// Pinned: what caldav-go handed its filter parser, "<></>" without a filter element.
func libraryFilterXML(body string) string {
	var raw struct {
		Filter struct {
			XMLName xml.Name
			Inner   string `xml:",innerxml"`
		} `xml:"urn:ietf:params:xml:ns:caldav filter"`
	}
	_ = xml.Unmarshal([]byte(body), &raw)
	return "<" + raw.Filter.XMLName.Local + ">" + raw.Filter.Inner + "</" + raw.Filter.XMLName.Local + ">"
}

// reportAnswer answers a parsed REPORT on a resource the user was granted access to.
type reportAnswer func(c *echo.Context, s *xorm.Session, req reportRequest) error

func serveReport(c *echo.Context, u *user.User, t target) error {
	s := db.NewSession()
	defer s.Close()

	// Before the body is looked at: a bad body must not tell an unreadable collection from a missing one.
	answer, err := authorizeReport(c, s, u, t)
	if err != nil {
		return writeError(c, err)
	}

	body, err := io.ReadAll(io.LimitReader(c.Request().Body, maxReportBodySize+1))
	if err != nil {
		return writeError(c, fmt.Errorf("%w: %w", errMalformedBody, err))
	}
	if len(body) > maxReportBodySize {
		return writeError(c, errBodyTooLarge)
	}
	logRequest(c.Request().Header, body)

	req, err := parseReport(string(body))
	if errors.Is(err, errUnsupportedReport) || (errors.Is(err, errMalformedBody) && !errors.Is(err, errTooManyRequestedProps)) {
		// Pinned: caldav-go answered any report it did not know, or could not read, with an empty 412.
		return c.NoContent(http.StatusPreconditionFailed)
	}
	if err != nil {
		return writeError(c, err)
	}
	return answer(c, s, req)
}

func authorizeReport(c *echo.Context, s *xorm.Session, u *user.User, t target) (reportAnswer, error) {
	ctx := propContext{principal: principalPathForUser(u.Username)}

	switch t := t.(type) {
	case entryTarget, principalTarget:
		// Pinned: caldav-go answered a query or multiget here with the requested resource and
		// none of its properties.
		href := path.Clean(c.Request().URL.Path)
		return func(c *echo.Context, _ *xorm.Session, req reportRequest) error {
			var props propfindRequest
			switch req := req.(type) {
			case multigetRequest:
				props = req.props
			case calendarQueryRequest:
				props = req.props
			default:
				return c.NoContent(http.StatusPreconditionFailed)
			}
			named, _ := props.(propRequest)
			return writeMultistatus(c, multistatus{
				responses: []msResponse{propResponse(href, nil, named.names)},
				minimal:   prefersMinimal(c),
			})
		}, nil
	case homeSetTarget:
		// Pinned: caldav-go found no resource at the home set.
		return nil, errNotFound
	case collectionTarget:
		readable, err := t.checkRead(s, u)
		if err != nil {
			return nil, err
		}
		return func(c *echo.Context, s *xorm.Session, req reportRequest) error {
			var responses []msResponse
			var err error
			switch req := req.(type) {
			case multigetRequest:
				responses, err = answerMultiget(s, u, readable, c.Request().URL.Path, req, ctx)
			case calendarQueryRequest:
				responses, err = answerCalendarQuery(s, u, readable, req, ctx)
			case syncCollectionRequest:
				return handleSyncCollectionReport(c, req, newStorage(u, t))
			}
			if err != nil {
				return writeError(c, err)
			}
			if err := s.Commit(); err != nil {
				return writeError(c, err)
			}
			return writeMultistatus(c, multistatus{responses: responses, minimal: prefersMinimal(c)})
		}, nil
	case taskTarget:
		if _, err := t.collection.checkRead(s, u); err != nil {
			return nil, err
		}
		task, err := findTask(s, u, t.collection, t.uid)
		if err != nil {
			return nil, err
		}
		if task == nil {
			return nil, errNotFound
		}
		owner, err := loadOwner(s, u, &t.collection.project.Project)
		if err != nil {
			return nil, err
		}
		resource := &taskResource{href: taskURL(t.collection.project.ID, task), task: task, owner: owner}
		// Pinned: caldav-go answered a query or multiget on a task with the task, whatever
		// the filter or the hrefs.
		return func(c *echo.Context, s *xorm.Session, req reportRequest) error {
			var props propfindRequest
			switch req := req.(type) {
			case multigetRequest:
				props = req.props
			case calendarQueryRequest:
				props = req.props
			default:
				return c.NoContent(http.StatusPreconditionFailed)
			}
			if err := s.Commit(); err != nil {
				return writeError(c, err)
			}
			return writeMultistatus(c, multistatus{
				responses: []msResponse{resource.answer(ctx, reportProps(props))},
				minimal:   prefersMinimal(c),
			})
		}, nil
	}
	return nil, fmt.Errorf("unhandled caldav target %T", t)
}

// Pinned: caldav-go answered allprop and propname in a REPORT with no properties at all.
func reportProps(req propfindRequest) propfindRequest {
	if named, ok := req.(propRequest); ok {
		return named
	}
	return propRequest{}
}

func prefersMinimal(c *echo.Context) bool {
	return c.Request().Header.Get("Prefer") == "return=minimal"
}

func answerMultiget(s *xorm.Session, u *user.User, c readableCollection, requestPath string, req multigetRequest, ctx propContext) ([]msResponse, error) {
	// Pinned: caldav-go dropped every href that does not start with the request path.
	prefix := path.Clean(requestPath)
	hrefs := make([]string, 0, len(req.hrefs))
	for _, href := range req.hrefs {
		if strings.HasPrefix(href, prefix) {
			hrefs = append(hrefs, href)
		}
	}

	tasks, err := multigetTasks(s, u, c, hrefs)
	if err != nil {
		return nil, err
	}
	owner, err := loadOwner(s, u, &c.project.Project)
	if err != nil {
		return nil, err
	}

	props := reportProps(req.props)
	responses := make([]msResponse, 0, len(hrefs))
	for _, href := range hrefs {
		task, found := tasks[href]
		if !found {
			responses = append(responses, statusResponse{href: href, status: http.StatusNotFound})
			continue
		}
		responses = append(responses, (&taskResource{href: href, task: task, owner: owner}).answer(ctx, props))
	}
	return responses, nil
}

// multigetTasks resolves hrefs to the tasks they name in c, keyed by the task's href. An href
// that is malformed, names a task the user cannot read, or one c does not hold is missing.
func multigetTasks(s *xorm.Session, u *user.User, c readableCollection, hrefs []string) (map[string]*models.Task, error) {
	found := map[string]*models.Task{}

	// Path format: /dav/projects/{projectID}/{uid}.ics.
	// Remember the href's project ID per uid so the consistency check below
	// can drop tasks requested via the wrong project (GHSA-48ch-p4gq-x46x).
	var uids []string
	uidURLProjects := map[string]int64{}
	for _, href := range hrefs {
		parts := strings.Split(href, "/")
		if len(parts) < 5 {
			continue
		}
		// Skip malformed hrefs: without these guards an empty/non-numeric
		// project would bypass the consistency check, and an empty UID
		// would query for tasks with empty/NULL uids.
		if !strings.HasSuffix(parts[4], ".ics") {
			continue
		}
		// echo does not decode path params either, so hrefs travel percent-encoded end to end.
		uid, err := url.PathUnescape(strings.TrimSuffix(parts[4], ".ics"))
		if err != nil || uid == "" {
			continue
		}
		urlProjectID, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			continue
		}
		uids = append(uids, uid)
		uidURLProjects[uid] = urlProjectID
	}
	if len(uids) == 0 {
		return found, nil
	}

	tasks, err := tasksByUIDs(s, u, uids)
	if err != nil {
		return nil, err
	}

	// A task reached through a pseudo collection lives in another project; membership
	// stands in for the project check. One lookup for all of them.
	viaPseudo := func(urlProjectID int64) bool {
		return models.IsPseudoProjectID(urlProjectID) && urlProjectID == c.project.ID
	}
	var candidates []int64
	for _, t := range tasks {
		urlProjectID, ok := uidURLProjects[t.UID]
		if ok && urlProjectID != t.ProjectID && viaPseudo(urlProjectID) {
			candidates = append(candidates, t.ID)
		}
	}
	members, err := c.containsAll(s, u, candidates)
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		urlProjectID, ok := uidURLProjects[t.UID]
		if !ok {
			continue
		}
		// Closes the URL-path leak for calendar-multiget REPORTs (GHSA-48ch-p4gq-x46x).
		if urlProjectID != t.ProjectID && (!viaPseudo(urlProjectID) || !members[t.ID]) {
			continue
		}
		// tasks.uid has no unique index, so two same-UID tasks would map to one href.
		href := taskURL(urlProjectID, t)
		if _, dup := found[href]; !dup {
			found[href] = t
		}
	}
	return found, nil
}

// One bind variable per uid, so chunked like the membership lookup.
func tasksByUIDs(s *xorm.Session, u *user.User, uids []string) ([]*models.Task, error) {
	var tasks []*models.Task
	for chunk := range slices.Chunk(uids, membershipQueryChunkSize) {
		found, err := models.GetTasksByUIDs(s, chunk, u)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, found...)
	}
	return tasks, nil
}

func answerCalendarQuery(s *xorm.Session, u *user.User, c readableCollection, req calendarQueryRequest, ctx propContext) ([]msResponse, error) {
	tasks, err := readCollectionTasks(s, u, c)
	if err != nil {
		return nil, err
	}
	owner := ownerName(u, &c.project.Project)
	props := reportProps(req.props)

	if tasks == nil {
		// Pinned: caldav-go answered a query on a collection without tasks with the collection.
		collection := &collectionResource{href: collectionHref(c.project.ID), project: c.project, owner: owner}
		return []msResponse{collection.answer(ctx, props)}, nil
	}

	// Pinned: caldav-go failed with a 500 on a filter it could not parse, or none at all.
	filter, err := data.ParseResourceFilters(req.filterXML)
	if err != nil {
		return nil, err
	}

	var responses []msResponse
	for _, t := range tasks {
		task := &t.Task
		href := taskURL(c.project.ID, task)
		adapter := &VikunjaProjectResourceAdapter{project: c.project, task: task}
		r := data.NewResource(href, adapter)
		if !matchesTodoFilter(filter, &r) {
			continue
		}
		// Pinned: two tasks with the same UID answer under the same href.
		resource := &taskResource{href: href, task: task, owner: owner, rendered: adapter.content}
		responses = append(responses, resource.answer(ctx, props))
	}
	return responses, nil
}
