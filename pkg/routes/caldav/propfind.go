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

// Same cap caldav-go had.
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
	d := xml.NewDecoder(strings.NewReader(body))
	var named, include []xml.Name
	var sawProp, sawPropName bool
	seen := map[xml.Name]struct{}{}
	var path []xml.Name
tokens:
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) && len(path) == 0 {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errMalformedBody, err)
		}
		switch t := tok.(type) {
		case xml.CharData:
			if len(path) == 0 && len(bytes.TrimSpace(t)) > 0 {
				return nil, fmt.Errorf("%w: text outside the root element", errMalformedBody)
			}
		case xml.StartElement:
			if len(path) == 1 && t.Name.Space == nsDAV {
				switch t.Name.Local {
				case "prop":
					sawProp = true
				case "propname":
					sawPropName = true
				}
			}
			if len(path) == 2 && path[1].Space == nsDAV && (path[1].Local == "prop" || path[1].Local == "include") {
				if _, dup := seen[t.Name]; !dup {
					if len(seen) == maxPropfindProperties {
						return nil, fmt.Errorf("%w: more than %d properties", errMalformedBody, maxPropfindProperties)
					}
					seen[t.Name] = struct{}{}
					if path[1].Local == "prop" {
						named = append(named, t.Name)
					} else {
						include = append(include, t.Name)
					}
				}
			}
			path = append(path, t.Name)
		case xml.EndElement:
			path = path[:len(path)-1]
			if len(path) == 0 {
				break tokens
			}
		}
	}

	switch {
	case sawProp:
		return propRequest{names: named}, nil
	case sawPropName:
		return propNameRequest{}, nil
	}
	return allPropRequest{include: include}, nil
}

// hrefOnlyResponse is what caldav-go answered when no property was asked for by name.
type hrefOnlyResponse struct{ href string }

func (r hrefOnlyResponse) writeResponse(w *xmlWriter, _ bool) {
	w.open(davName("response"))
	w.textElement(davName("href"), r.href)
	w.close(davName("response"))
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
	principal := propContext{currentUserPrincipal: principalPathForUser(u.Username)}
	// Pinned: caldav-go was configured with the home set as the user of these targets.
	homeSet := propContext{currentUserPrincipal: ProjectHomeSetPath}

	switch t := t.(type) {
	case entryTarget:
		return func(*xorm.Session, bool) ([]davResource, error) {
			return []davResource{&entryResource{href: DavBasePath}}, nil
		}, principal, nil
	case principalTarget:
		// Pinned: the principal answers as the home set.
		return func(*xorm.Session, bool) ([]davResource, error) {
			return []davResource{&homeSetResource{href: ProjectBasePath}}, nil
		}, principal, nil
	case homeSetTarget:
		return func(s *xorm.Session, deep bool) ([]davResource, error) {
			if !deep {
				return []davResource{&homeSetResource{href: ProjectBasePath}}, nil
			}
			return loadHomeSet(s, u)
		}, homeSet, nil
	case collectionTarget:
		readable, err := t.checkRead(s, u)
		if err != nil {
			return nil, propContext{}, err
		}
		return func(s *xorm.Session, deep bool) ([]davResource, error) {
			return loadCollectionResources(s, u, readable, collectionHref(t.project.ID), deep)
		}, homeSet, nil
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
			return []davResource{&taskResource{href: taskURL(t.collection.project.ID, task), task: task}}, nil
		}, principal, nil
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
		})
	}
	return resources, nil
}

func loadCollectionResources(s *xorm.Session, u *user.User, c readableCollection, href string, deep bool) ([]davResource, error) {
	tasks, err := readCollectionTasks(s, u, c)
	if err != nil {
		return nil, err
	}

	resources := []davResource{&collectionResource{href: href, project: c.project, tasks: tasks}}
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
		resources = append(resources, &taskResource{href: href, task: &t.Task})
	}
	return resources, nil
}

func collectionHref(projectID int64) string {
	return ProjectBasePath + "/" + strconv.FormatInt(projectID, 10)
}
