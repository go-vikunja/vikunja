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
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"xorm.io/xorm"
)

const (
	// A multiget names every task it wants, so it is the largest body a client sends.
	maxReportBodySize = 1 << 20
	maxMultigetHrefs  = 5000
	maxReportElements = 2 * maxMultigetHrefs
	maxReportDepth    = 16
)

var errUnsupportedReport = preconditionError{
	precondition: davName("supported-report"),
	reason:       "unsupported report",
}

type reportRequest interface{ isReportRequest() }

type multigetRequest struct {
	props propfindRequest
	// Without duplicates, in request order.
	hrefs []string
}

type calendarQueryRequest struct {
	props  propfindRequest
	filter calendarFilter
}

type syncCollectionRequest struct {
	props propfindRequest
	token string
}

func (multigetRequest) isReportRequest()       {}
func (calendarQueryRequest) isReportRequest()  {}
func (syncCollectionRequest) isReportRequest() {}

func parseReport(body string) (reportRequest, error) {
	p := reportParser{seenHref: map[string]bool{}}
	if err := walkXML(body, p.visit); err != nil {
		return nil, err
	}

	switch p.root {
	case xml.Name{}:
		return nil, fmt.Errorf("%w: empty body", errMalformedBody)
	case calDAVName("calendar-multiget"):
		return multigetRequest{
			props: p.props.request(),
			hrefs: p.hrefs,
		}, nil
	case calDAVName("calendar-query"):
		// RFC 4791 §9.5: the filter is required.
		if p.filter == nil {
			return nil, fmt.Errorf("%w: calendar-query without a filter", errMalformedBody)
		}
		filter, err := parseFilter(p.filter)
		if err != nil {
			return nil, err
		}
		return calendarQueryRequest{
			props:  p.props.request(),
			filter: filter,
		}, nil
	case davName("sync-collection"):
		return syncCollectionRequest{
			props: p.props.request(),
			token: p.token,
		}, nil
	}
	return nil, errUnsupportedReport
}

type reportParser struct {
	root     xml.Name
	props    propCollector
	hrefs    []string
	seenHref map[string]bool
	token    string
	elements int
	text     *strings.Builder
	// The filter element, and the open elements in it.
	filter         *xmlElement
	filterStack    []*xmlElement
	filterElements int
}

func (p *reportParser) visit(path []xml.Name, tok xml.Token) error {
	switch t := tok.(type) {
	case xml.StartElement:
		return p.start(path, t)
	case xml.CharData:
		if len(p.filterStack) > 0 {
			p.filterStack[len(p.filterStack)-1].text.Write(t)
		}
		if len(path) == 2 && p.text != nil {
			p.text.Write(t)
		}
	case xml.EndElement:
		if len(p.filterStack) > 0 {
			p.filterStack = p.filterStack[:len(p.filterStack)-1]
			return nil
		}
		if len(path) == 1 && p.text != nil {
			return p.endText(t.Name)
		}
	}
	return nil
}

func (p *reportParser) start(path []xml.Name, t xml.StartElement) error {
	if len(path) == 0 {
		p.root = t.Name
		return nil
	}
	p.elements++
	if p.elements > maxReportElements {
		return fmt.Errorf("%w: more than %d elements", errBodyTooLarge, maxReportElements)
	}
	if len(path) >= maxReportDepth {
		return fmt.Errorf("%w: nested deeper than %d elements", errBodyTooLarge, maxReportDepth)
	}
	if len(p.filterStack) > 0 || (len(path) == 1 && t.Name == calDAVName("filter")) {
		return p.openFilterElement(t)
	}
	if len(path) == 1 && (t.Name == davName("href") || t.Name == davName("sync-token")) {
		p.text = &strings.Builder{}
	}
	return p.props.visit(path, t)
}

func (p *reportParser) endText(name xml.Name) error {
	text := p.text.String()
	p.text = nil
	switch name {
	case davName("href"):
		if p.seenHref[text] {
			return nil
		}
		if len(p.hrefs) == maxMultigetHrefs {
			return fmt.Errorf("%w: more than %d hrefs", errBodyTooLarge, maxMultigetHrefs)
		}
		p.seenHref[text] = true
		p.hrefs = append(p.hrefs, text)
	case davName("sync-token"):
		p.token = strings.TrimSpace(text)
	}
	return nil
}

func (p *reportParser) openFilterElement(t xml.StartElement) error {
	p.filterElements++
	if p.filterElements > maxFilterElements {
		return fmt.Errorf("%w: more than %d filter elements", errBodyTooLarge, maxFilterElements)
	}
	el := &xmlElement{
		name:  t.Name,
		attrs: t.Copy().Attr,
	}
	if len(p.filterStack) == 0 {
		if p.filter != nil {
			return invalidFilter("more than one filter")
		}
		p.filter = el
	} else {
		parent := p.filterStack[len(p.filterStack)-1]
		parent.children = append(parent.children, el)
	}
	p.filterStack = append(p.filterStack, el)
	return nil
}

type reportAnswer func(c *echo.Context, s *xorm.Session, req reportRequest) error

func serveReport(c *echo.Context, u *user.User, t target) error {
	switch t.(type) {
	case entryTarget, principalTarget, homeSetTarget:
		return methodNotAllowed(c, u, t)
	}

	// Before the session: a trickled body must not pin a DB connection.
	body, bodyErr := readBody(c, maxReportBodySize)

	s := db.NewSession()
	defer s.Close()

	answer, err := authorizeReport(s, u, t)
	if err != nil {
		return writeError(c, err)
	}
	// After the access check: a bad body must not tell an unreadable collection from a missing one.
	if bodyErr != nil {
		return writeError(c, bodyErr)
	}

	req, err := parseReport(string(body))
	if err != nil {
		return writeError(c, err)
	}
	return answer(c, s, req)
}

func authorizeReport(s *xorm.Session, u *user.User, t target) (reportAnswer, error) {
	ctx := propContext{principal: principalPathForUser(u.Username)}

	switch t := t.(type) {
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
				responses, err = answerMultiget(s, u, readable, req, ctx)
			case calendarQueryRequest:
				responses, err = answerCalendarQuery(s, u, readable, req, ctx)
			case syncCollectionRequest:
				return handleSyncCollectionReport(c, s, u, readable, req, ctx)
			}
			if err != nil {
				return writeError(c, err)
			}
			if err := s.Commit(); err != nil {
				return writeError(c, err)
			}
			return writeMultistatus(c, multistatus{
				responses: responses,
				minimal:   prefersMinimal(c.Request().Header),
			})
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
		resource := &taskResource{
			href:  taskURL(t.collection.project.ID, task),
			task:  task,
			owner: owner,
		}
		// Pinned: a query or multiget on a task answers the task, whatever the filter or hrefs.
		return func(c *echo.Context, s *xorm.Session, req reportRequest) error {
			var props propfindRequest
			switch req := req.(type) {
			case multigetRequest:
				props = req.props
			case calendarQueryRequest:
				props = req.props
			default:
				return writeError(c, errUnsupportedReport)
			}
			if err := s.Commit(); err != nil {
				return writeError(c, err)
			}
			return writeMultistatus(c, multistatus{
				responses: []msResponse{resource.answer(ctx, props)},
				minimal:   prefersMinimal(c.Request().Header),
			})
		}, nil
	}
	return nil, fmt.Errorf("unhandled caldav target %T", t)
}

func answerMultiget(s *xorm.Session, u *user.User, c readableCollection, req multigetRequest, ctx propContext) ([]msResponse, error) {
	// Only members are looked up; anything else is not found here (RFC 4791 §7.9).
	members := collectionHref(c.project.ID) + "/"
	hrefs := make([]string, 0, len(req.hrefs))
	for _, href := range req.hrefs {
		if strings.HasPrefix(href, members) {
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

	responses := make([]msResponse, 0, len(req.hrefs))
	for _, href := range req.hrefs {
		task, found := tasks[href]
		if !found {
			responses = append(responses, statusResponse{
				href:   href,
				status: http.StatusNotFound,
			})
			continue
		}
		resource := &taskResource{
			href:  href,
			task:  task,
			owner: owner,
		}
		responses = append(responses, resource.answer(ctx, req.props))
	}
	return responses, nil
}

func multigetTasks(s *xorm.Session, u *user.User, c readableCollection, hrefs []string) (map[string]*models.Task, error) {
	found := map[string]*models.Task{}

	var uids []string
	uidURLProjects := map[string]int64{}
	for _, href := range hrefs {
		parts := strings.Split(href, "/")
		if len(parts) < 5 {
			continue
		}
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

	// Via a pseudo collection a task lives in another project; membership stands in for the project check.
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
		// GHSA-48ch-p4gq-x46x: drop tasks requested via the wrong project.
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

// Chunked: one bind variable per uid.
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
	props := req.props

	if tasks == nil {
		// Pinned: a query on a collection without tasks answers the collection itself.
		collection := &collectionResource{
			href:    collectionHref(c.project.ID),
			project: c.project,
			owner:   owner,
		}
		return []msResponse{collection.answer(ctx, props)}, nil
	}

	var responses []msResponse
	for _, t := range tasks {
		href := taskURL(c.project.ID, &t.Task)
		// Pinned: two tasks with the same UID answer under the same href.
		candidate := &queryCandidate{resource: &taskResource{
			href:  href,
			task:  &t.Task,
			owner: owner,
		}}
		if req.filter.matches(candidate) {
			responses = append(responses, candidate.resource.answer(ctx, props))
		}
	}
	return responses, nil
}
