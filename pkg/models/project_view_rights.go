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

func (pv *ProjectView) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	filterID := GetSavedFilterIDFromProjectID(pv.ProjectID)
	if filterID > 0 {
		sf := &SavedFilter{ID: filterID}
		return sf.CanRead(s, a)
	}

	pp := pv.getProject()
	return pp.CanRead(s, a)
}

func (pv *ProjectView) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	filterID := GetSavedFilterIDFromProjectID(pv.ProjectID)
	if filterID > 0 {
		sf := &SavedFilter{ID: filterID}
		return sf.CanDelete(s, a)
	}

	pp := pv.getProject()
	return pp.IsAdmin(s, a)
}

func (pv *ProjectView) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	filterID := GetSavedFilterIDFromProjectID(pv.ProjectID)
	if filterID > 0 {
		sf := &SavedFilter{ID: filterID}
		return sf.CanUpdate(s, a)
	}

	pp := pv.getProject()
	return pp.IsAdmin(s, a)
}

func (pv *ProjectView) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	filterID := GetSavedFilterIDFromProjectID(pv.ProjectID)
	if filterID > 0 {
		sf := &SavedFilter{ID: filterID}
		return sf.CanUpdate(s, a)
	}

	pp := pv.getProject()
	return pp.IsAdmin(s, a)
}

func (pv *ProjectView) getProject() (pp *Project) {
	return &Project{ID: pv.ProjectID}
}
