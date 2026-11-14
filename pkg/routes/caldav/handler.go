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
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"code.vikunja.io/api/pkg/db"

	caldav2 "code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v5"
	"github.com/samedi/caldav-go"
	"github.com/samedi/caldav-go/data"
	"github.com/samedi/caldav-go/lib"
)

func getBasicAuthUserFromContext(c *echo.Context) (*user.User, error) {
	u, is := c.Get("userBasicAuth").(*user.User)
	if !is {
		return &user.User{}, fmt.Errorf("user is not user element, is %s", reflect.TypeOf(c.Get("userBasicAuth")))
	}
	return u, nil
}

// Per request, never through caldav-go's setup globals: those let concurrent
// requests of different users answer each other with the wrong storage.
func caldavConfig(storage data.Storage, principalPath string) caldav.Config {
	return caldav.Config{
		Storage: storage,
		// caldav-go renders current-user-principal as "/<name>/", so pass the path
		// without surrounding slashes or the href ends up with a double slash.
		User:                &data.CalUser{Name: strings.Trim(principalPath, "/")},
		SupportedComponents: []string{lib.VCALENDAR, lib.VTODO},
	}
}

// EntryHandler serves /dav/.
func EntryHandler(c *echo.Context) error { return handle(c, resolveEntry) }

// PrincipalHandler serves /dav/principals/<user>/ and /.well-known/caldav.
func PrincipalHandler(c *echo.Context) error { return handle(c, resolvePrincipal) }

// ProjectHandler serves the home set /dav/projects and the collections below it.
func ProjectHandler(c *echo.Context) error { return handle(c, resolveProjects) }

// TaskHandler serves a single task below a collection.
func TaskHandler(c *echo.Context) error { return handle(c, resolveTask) }

func handle(c *echo.Context, resolve targetResolver) error {
	u, err := getBasicAuthUserFromContext(c)
	if err != nil {
		return writeError(c, err)
	}

	t, err := resolve(c, u)
	if err != nil {
		return writeError(c, err)
	}

	if c.Request().Method == http.MethodOptions {
		return serveOptions(c, u, t)
	}

	switch t := t.(type) {
	case entryTarget, principalTarget:
		readAndLogBody(c)
		return serveLibrary(c, newStorage(u, t), principalPathForUser(u.Username))
	case homeSetTarget, collectionTarget:
		return serveProjects(c, u, t)
	case taskTarget:
		return serveTask(c, u, t)
	}
	return writeError(c, fmt.Errorf("unhandled caldav target %T", t))
}

func serveLibrary(c *echo.Context, storage *VikunjaCaldavProjectStorage, principalPath string) error {
	response := caldav.HandleRequestWithConfig(c.Request(), caldavConfig(storage, principalPath))
	response.Write(c.Response())
	return nil
}

func readAndLogBody(c *echo.Context) []byte {
	body, _ := io.ReadAll(c.Request().Body)
	c.Request().Body = io.NopCloser(bytes.NewBuffer(body))
	logRequest(c.Request().Header, body)
	return body
}

// serveProjects serves the home set and collections.
func serveProjects(c *echo.Context, u *user.User, t target) error {
	storage := newStorage(u, t)
	body := readAndLogBody(c)

	if c.Request().Method == http.MethodPut && strings.HasPrefix(string(body), `BEGIN:VCALENDAR`) {
		s := db.NewSession()
		canRead, err := storage.collection().canRead(s, u)
		s.Close()
		if err != nil {
			return writeError(c, err)
		}
		if !canRead {
			return writeError(c, errNotFound)
		}

		if err := caldav2.CheckVCalendarLimits(string(body)); err != nil {
			return writeError(c, err)
		}

		storage.task, _, err = caldav2.ParseTaskFromVTODO(string(body))
		if err != nil {
			log.Warningf("[CALDAV] Failed to parse task: %v", err)
			return writeError(c, models.ErrInvalidData{Message: "Invalid task"})
		}
	}

	// RFC 6578: intercept sync-collection REPORT before caldav-go sees it.
	// caldav-go returns 412 (empty body) for unknown REPORT types, which causes
	// iOS to silently stop syncing after the initial account add.
	if c.Request().Method == "REPORT" && strings.Contains(string(body), "sync-collection") {
		log.Debugf("[CALDAV] sync-collection REPORT for project %d by user %s", storage.project.ID, u.Username)
		return handleSyncCollectionReport(c, string(body), storage)
	}

	collection, isCollection := t.(collectionTarget)
	if !isCollection {
		// Pinned: no PROPPATCH special case for the home set.
		return serveLibrary(c, storage, ProjectHomeSetPath)
	}

	// caldav-go has no PROPPATCH dispatch case and falls back to a blanket 501;
	// intercept it here, same as the sync-collection REPORT above.
	if c.Request().Method == "PROPPATCH" {
		log.Debugf("[CALDAV] PROPPATCH for project %d by user %s", collection.project.ID, u.Username)
		return handlePropPatch(c, string(body), collection.project, u, propPatchCollection)
	}

	return serveLibrary(c, storage, ProjectHomeSetPath)
}

func serveTask(c *echo.Context, u *user.User, t taskTarget) error {
	// A task resource has no settable properties; it gets a 207 refusal
	// rather than caldav-go's blanket 501.
	if c.Request().Method == "PROPPATCH" {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return writeError(c, fmt.Errorf("%w: %w", errMalformedBody, err))
		}
		c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

		return handlePropPatch(c, string(body), t.collection.project, u, propPatchTask)
	}

	storage := newStorage(u, t)
	if c.Request().Method == http.MethodPut {
		body, err := io.ReadAll(io.LimitReader(c.Request().Body, caldav2.MaxVCalendarSize+1))
		if err != nil {
			return writeError(c, fmt.Errorf("%w: %w", errMalformedBody, err))
		}
		if err := caldav2.CheckVCalendarLimits(string(body)); err != nil {
			return writeError(c, storage.rejectBody(err))
		}
		c.Request().Body = io.NopCloser(bytes.NewBuffer(body))
	}

	return serveLibrary(c, storage, principalPathForUser(u.Username))
}

const maxLoggedBodySize = 4096

func logRequest(header http.Header, body []byte) {
	if len(body) > maxLoggedBodySize {
		body = body[:maxLoggedBodySize]
	}
	log.Debugf("[CALDAV] Request Body: %s", body)
	log.Debugf("[CALDAV] Request Headers: %v", log.SafeHeaders(header))
}

// Body checks precede caldav-go's access checks, so an unreadable collection must answer like a missing one.
func (vcls *VikunjaCaldavProjectStorage) rejectBody(checkErr error) error {
	// The empty project of /dav/projects//<uid>.ics, see resolveTask.
	if vcls.project.ID == 0 {
		return checkErr
	}

	s := db.NewSession()
	defer s.Close()
	canRead, err := vcls.collection().canRead(s, vcls.user)
	if err != nil {
		return err
	}
	if !canRead {
		return errNotFound
	}
	return checkErr
}
