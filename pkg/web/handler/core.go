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

package handler

import (
	"context"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/web"

	"github.com/labstack/echo/v5"
)

// DoCreate runs the permission check + model Create + commit pipeline for a
// CObject. Framework-agnostic: callable from both Echo (CreateWeb) and Huma.
// Caller is responsible for body/path binding and validation before calling.
func DoCreate(_ context.Context, obj CObject, a web.Auth) error {
	s := db.NewSession()
	defer func() {
		if err := s.Close(); err != nil {
			log.Errorf("Could not close session: %s", err)
		}
	}()

	canCreate, err := obj.CanCreate(s, a)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return err
	}
	if !canCreate {
		_ = s.Rollback()
		events.CleanupPending(s)
		log.Warningf("Tried to create while not having the permissions for it (User: %v)", a)
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := obj.Create(s, a); err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return err
	}

	if err := s.Commit(); err != nil {
		events.CleanupPending(s)
		return err
	}

	events.DispatchPending(s)
	return nil
}

// DoReadOne runs the permission check + model ReadOne + commit pipeline for a
// CObject. obj should have its identifying fields set before call. On success,
// obj is fully populated. maxPermission is exposed via the x-max-permission
// header in the Echo wrapper; Huma wrapper may ignore it.
func DoReadOne(_ context.Context, obj CObject, a web.Auth) (maxPermission int, err error) {
	s := db.NewSession()
	defer func() {
		if cerr := s.Close(); cerr != nil {
			log.Errorf("Could not close session: %s", cerr)
		}
	}()

	canRead, maxPermission, err := obj.CanRead(s, a)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return 0, err
	}
	if !canRead {
		_ = s.Rollback()
		events.CleanupPending(s)
		log.Warningf("Tried to read while not having the permissions for it (User: %v)", a)
		return 0, echo.NewHTTPError(http.StatusForbidden, "You don't have the permission to see this")
	}

	if err := obj.ReadOne(s, a); err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return 0, err
	}

	if err := s.Commit(); err != nil {
		events.CleanupPending(s)
		return 0, err
	}

	events.DispatchPending(s)
	return maxPermission, nil
}
