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
	"net/url"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

// target is the resource a request addresses, resolved once per request.
type target interface{ isTarget() }

// entryTarget is /dav/.
type entryTarget struct{}

// principalTarget is the authenticated user's own principal, or /.well-known/caldav.
type principalTarget struct {
	// What the principal answers as: the canonical principal path, or the well-known path asked for.
	href string
}

// homeSetTarget is /dav/projects.
type homeSetTarget struct{}

// collectionTarget is /dav/projects/<id>: a project, the Favorites pseudo project or a
// saved filter. Loaded without a permission check.
type collectionTarget struct {
	project *models.ProjectWithTasksAndBuckets
}

// taskTarget is /dav/projects/<id>/<uid>.ics. The task itself is not loaded.
type taskTarget struct {
	collection collectionTarget
	uid        string
}

func (entryTarget) isTarget()      {}
func (principalTarget) isTarget()  {}
func (homeSetTarget) isTarget()    {}
func (collectionTarget) isTarget() {}
func (taskTarget) isTarget()       {}

type targetResolver func(c *echo.Context, u *user.User) (target, error)

func resolveEntry(*echo.Context, *user.User) (target, error) {
	return entryTarget{}, nil
}

func resolvePrincipal(c *echo.Context, u *user.User) (target, error) {
	path := c.Request().URL.Path
	if isWellKnownPath(path) {
		return principalTarget{href: path}, nil
	}
	if !isOwnPrincipalPath(path, u.Username) {
		return nil, errNotFound
	}
	return principalTarget{href: principalPathForUser(u.Username)}, nil
}

func isWellKnownPath(path string) bool {
	return strings.TrimSuffix(path, "/") == "/.well-known/caldav"
}

// Sub-paths and foreign usernames must 404 (RFC 4918) instead of falling
// through to the merged principal pseudo-resource, which would silently serve
// the authenticated user's data and act as a username enumeration oracle.
func isOwnPrincipalPath(path, username string) bool {
	rest := strings.TrimPrefix(strings.TrimPrefix(path, PrincipalBasePath), "/")
	return strings.TrimSuffix(rest, "/") == username
}

func resolveProjects(c *echo.Context, _ *user.User) (target, error) {
	param := c.Param("project")
	if param == "" {
		return homeSetTarget{}, nil
	}
	project, err := loadCollection(param)
	if err != nil {
		return nil, err
	}
	return collectionTarget{project: project}, nil
}

func resolveTask(c *echo.Context, _ *user.User) (target, error) {
	// /dav/projects//<uid>.ics routes here too, with an empty project param.
	project, err := loadCollection(c.Param("project"))
	if err != nil {
		return nil, err
	}

	// Whether c.Param("task") arrives decoded depends on router config and on which
	// bytes were escaped; the escaped path is unambiguous, so decode that instead.
	esc := c.Request().URL.EscapedPath()
	seg := esc[strings.LastIndex(esc, "/")+1:]
	uid, err := url.PathUnescape(strings.TrimSuffix(seg, ".ics"))
	if err != nil || uid == "" {
		return nil, errNotFound
	}

	return taskTarget{collection: collectionTarget{project: project}, uid: uid}, nil
}

func loadCollection(param string) (*models.ProjectWithTasksAndBuckets, error) {
	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return nil, models.ErrProjectDoesNotExist{}
	}

	if id == models.FavoritesPseudoProjectID {
		return &models.ProjectWithTasksAndBuckets{Project: models.FavoritesPseudoProject}, nil
	}

	s := db.NewSession()
	defer s.Close()

	if id < models.FavoritesPseudoProjectID {
		sf, err := models.GetSavedFilterSimpleByID(s, models.GetSavedFilterIDFromProjectID(id))
		if err != nil {
			return nil, err
		}
		return &models.ProjectWithTasksAndBuckets{Project: *sf.ToProject()}, nil
	}

	p, err := models.GetProjectSimpleByID(s, id)
	if err != nil {
		return nil, err
	}
	return &models.ProjectWithTasksAndBuckets{Project: *p}, nil
}

// newStorage is the only way the caldav-go storage gets built.
func newStorage(u *user.User, t target) *VikunjaCaldavProjectStorage {
	storage := &VikunjaCaldavProjectStorage{user: u}
	switch t := t.(type) {
	case homeSetTarget:
		storage.project = &models.ProjectWithTasksAndBuckets{}
	case collectionTarget:
		storage.project = t.project
	case taskTarget:
		storage.project = t.collection.project
		storage.task = &models.Task{UID: t.uid}
	}
	return storage
}
