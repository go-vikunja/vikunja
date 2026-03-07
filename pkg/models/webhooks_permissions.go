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

package models

import (
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

func (w *Webhook) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	// User-level webhook: user owns it
	if w.UserID > 0 {
		return w.UserID == a.GetID(), int(PermissionRead), nil
	}

	// Project-level webhook: delegate to project
	p := &Project{ID: w.ProjectID}
	return p.CanRead(s, a)
}

func (w *Webhook) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return w.canDoWebhook(s, a)
}

func (w *Webhook) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return w.canDoWebhook(s, a)
}

func (w *Webhook) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return w.canDoWebhook(s, a)
}

func (w *Webhook) canDoWebhook(s *xorm.Session, a web.Auth) (bool, error) {
	_, isShareAuth := a.(*LinkSharing)
	if isShareAuth {
		return false, nil
	}

	// User-level webhook: user owns it or is creating new
	if w.UserID > 0 || w.ProjectID == 0 {
		return w.UserID == 0 || w.UserID == a.GetID(), nil
	}

	// Project-level webhook: delegate to project
	p := &Project{ID: w.ProjectID}
	return p.CanUpdate(s, a)
}
