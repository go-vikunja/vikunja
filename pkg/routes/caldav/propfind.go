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
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"xorm.io/xorm"
)

const maxPropfindProperties = 200

// propfindRequest is what a PROPFIND body asks for (RFC 4918 §9.1).
type propfindRequest interface{ isPropfindRequest() }

// propRequest names the properties to answer, without duplicates.
type propRequest struct{ names []xml.Name }

// allPropRequest is <allprop/> or an empty body; include names properties allprop leaves out.
type allPropRequest struct{ include []xml.Name }

// propNameRequest asks for the names of all properties, without values.
type propNameRequest struct{}

func (propRequest) isPropfindRequest()     {}
func (allPropRequest) isPropfindRequest()  {}
func (propNameRequest) isPropfindRequest() {}

func parsePropfind(body string) (propfindRequest, error) {
	var props propCollector
	err := walkXML(body, func(path []xml.Name, tok xml.Token) error {
		if t, isStart := tok.(xml.StartElement); isStart {
			return props.visit(path, t)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return props.request(), nil
}

// walkXML calls visit for every token inside the root element, with the open elements
// enclosing it, root first. An empty body visits nothing.
func walkXML(body string, visit func(path []xml.Name, tok xml.Token) error) error {
	d := xml.NewDecoder(strings.NewReader(body))
	var path []xml.Name
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) && len(path) == 0 {
			return nil
		}
		if err != nil {
			return fmt.Errorf("%w: %w", errMalformedBody, err)
		}
		switch t := tok.(type) {
		case xml.CharData:
			if len(path) == 0 {
				if len(bytes.TrimSpace(t)) > 0 {
					return fmt.Errorf("%w: text outside the root element", errMalformedBody)
				}
				continue
			}
			if err := visit(path, t); err != nil {
				return err
			}
		case xml.StartElement:
			if err := visit(path, t); err != nil {
				return err
			}
			path = append(path, t.Name)
		case xml.EndElement:
			path = path[:len(path)-1]
			if len(path) == 0 {
				return nil
			}
			if err := visit(path, t); err != nil {
				return err
			}
		}
	}
}

// propCollector collects what the DAV:prop, DAV:allprop, DAV:include and DAV:propname
// children of a request's root ask for.
type propCollector struct {
	named, include       []xml.Name
	sawProp, sawPropName bool
	seen                 map[xml.Name]struct{}
}

func (p *propCollector) visit(path []xml.Name, t xml.StartElement) error {
	if len(path) == 1 && t.Name.Space == nsDAV {
		switch t.Name.Local {
		case "prop":
			p.sawProp = true
		case "propname":
			p.sawPropName = true
		}
	}
	if len(path) != 2 || path[1].Space != nsDAV || (path[1].Local != "prop" && path[1].Local != "include") {
		return nil
	}
	if _, dup := p.seen[t.Name]; dup {
		return nil
	}
	if len(p.seen) == maxPropfindProperties {
		return fmt.Errorf("%w: more than %d properties", errMalformedBody, maxPropfindProperties)
	}
	if p.seen == nil {
		p.seen = map[xml.Name]struct{}{}
	}
	p.seen[t.Name] = struct{}{}
	if path[1].Local == "prop" {
		p.named = append(p.named, t.Name)
	} else {
		p.include = append(p.include, t.Name)
	}
	return nil
}

func (p *propCollector) request() propfindRequest {
	switch {
	case p.sawProp:
		return propRequest{names: p.named}
	case p.sawPropName:
		return propNameRequest{}
	}
	return allPropRequest{include: p.include}
}

// propfindLoader loads the resources a PROPFIND answers, once access is granted.
type propfindLoader func(s *xorm.Session, deep bool) ([]davResource, error)

func servePropfind(c *echo.Context, u *user.User, t target) error {
	s := db.NewSession()
	defer s.Close()

	// Before the body is looked at: a bad body must not tell an unreadable collection from a missing one.
	load, ctx, err := authorizePropfind(s, u, t)
	if err != nil {
		return writeError(c, err)
	}

	body, err := io.ReadAll(io.LimitReader(c.Request().Body, maxPropRequestBodySize+1))
	if err != nil {
		return writeError(c, fmt.Errorf("%w: %w", errMalformedBody, err))
	}
	if len(body) > maxPropRequestBodySize {
		return writeError(c, errBodyTooLarge)
	}
	logRequest(c.Request().Header, body)
	req, err := parsePropfind(string(body))
	if err != nil {
		return writeError(c, err)
	}

	// Pinned: anything but 1, including infinity, is Depth 0.
	resources, err := load(s, c.Request().Header.Get("Depth") == "1")
	if err != nil {
		return writeError(c, err)
	}
	if err := s.Commit(); err != nil {
		return writeError(c, err)
	}

	ms := multistatus{
		responses: make([]msResponse, len(resources)),
		minimal:   c.Request().Header.Get("Prefer") == "return=minimal",
	}
	for i, r := range resources {
		ms.responses[i] = r.answer(ctx, req)
	}
	return writeMultistatus(c, ms)
}

func authorizePropfind(s *xorm.Session, u *user.User, t target) (propfindLoader, propContext, error) {
	ctx := propContext{principal: principalPathForUser(u.Username)}

	switch t := t.(type) {
	case entryTarget:
		return func(*xorm.Session, bool) ([]davResource, error) {
			return []davResource{&entryResource{href: DavBasePath + "/"}}, nil
		}, ctx, nil
	case principalTarget:
		return func(*xorm.Session, bool) ([]davResource, error) {
			return []davResource{&principalResource{href: t.href}}, nil
		}, ctx, nil
	case homeSetTarget:
		return func(s *xorm.Session, deep bool) ([]davResource, error) {
			if !deep {
				return []davResource{&homeSetResource{href: ProjectBasePath}}, nil
			}
			return loadHomeSet(s, u)
		}, ctx, nil
	case collectionTarget:
		readable, err := t.checkRead(s, u)
		if err != nil {
			return nil, propContext{}, err
		}
		return func(s *xorm.Session, deep bool) ([]davResource, error) {
			return loadCollectionResources(s, u, readable, collectionHref(t.project.ID), deep)
		}, ctx, nil
	case taskTarget:
		if _, err := t.collection.checkRead(s, u); err != nil {
			return nil, propContext{}, err
		}
		// A task has no members, so Depth: 1 answers like Depth: 0.
		return func(s *xorm.Session, _ bool) ([]davResource, error) {
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
			return []davResource{&taskResource{href: taskURL(t.collection.project.ID, task), task: task, owner: owner}}, nil
		}, ctx, nil
	}
	return nil, propContext{}, fmt.Errorf("unhandled caldav target %T", t)
}

func loadHomeSet(s *xorm.Session, u *user.User) ([]davResource, error) {
	// Page 0: everything, no paging.
	result, _, _, err := (&models.Project{}).ReadAll(s, u, "", 0, 0)
	if err != nil {
		return nil, err
	}
	projects, ok := result.([]*models.Project)
	if !ok {
		return nil, fmt.Errorf("unexpected project result type %T", result)
	}

	resources := make([]davResource, 0, len(projects)+1)
	resources = append(resources, &homeSetResource{href: ProjectBasePath})
	for _, p := range projects {
		resources = append(resources, &collectionResource{
			href:    collectionHref(p.ID),
			project: &models.ProjectWithTasksAndBuckets{Project: *p},
			// ReadAll loaded all owners at once.
			owner: ownerName(u, p),
		})
	}
	return resources, nil
}

func loadCollectionResources(s *xorm.Session, u *user.User, c readableCollection, href string, deep bool) ([]davResource, error) {
	tasks, err := readCollectionTasks(s, u, c)
	if err != nil {
		return nil, err
	}

	// readCollectionTasks loaded the owner.
	owner := ownerName(u, &c.project.Project)
	resources := []davResource{&collectionResource{href: href, project: c.project, tasks: tasks, owner: owner}}
	if !deep {
		return resources, nil
	}

	emitted := make(map[string]bool, len(tasks))
	for _, t := range tasks {
		// tasks.uid has no unique index, so two same-UID tasks would map to one href.
		href := taskURL(c.project.ID, &t.Task)
		if emitted[href] {
			continue
		}
		emitted[href] = true
		resources = append(resources, &taskResource{href: href, task: &t.Task, owner: owner})
	}
	return resources, nil
}

// Favorites and saved filters belong to the user looking at them.
func ownerName(u *user.User, p *models.Project) string {
	if models.IsPseudoProjectID(p.ID) {
		return u.Username
	}
	if p.Owner == nil {
		return ""
	}
	return p.Owner.Username
}

func loadOwner(s *xorm.Session, u *user.User, p *models.Project) (string, error) {
	if p.Owner == nil && !models.IsPseudoProjectID(p.ID) {
		owner, err := user.GetUserByID(s, p.OwnerID)
		if err != nil && !user.IsErrUserDoesNotExist(err) {
			return "", err
		}
		p.Owner = owner
	}
	return ownerName(u, p), nil
}

func collectionHref(projectID int64) string {
	return ProjectBasePath + "/" + strconv.FormatInt(projectID, 10)
}
