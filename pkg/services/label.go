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

package services

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// LabelService is a service for managing labels.
type LabelService struct{}

// NewLabelService returns a new LabelService.
func NewLabelService() *LabelService {
	return &LabelService{}
}

func (ls *LabelService) GetAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}
	return (&models.Label{}).ReadAll(s, u, search, page, perPage)
}

func (ls *LabelService) Get(s *xorm.Session, labelID int64, a web.Auth) (*models.Label, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}

	l := &models.Label{ID: labelID}
	can, _, err := l.CanRead(s, u)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, echo.ErrForbidden
	}

	if err := l.ReadOne(s); err != nil {
		return nil, err
	}

	return l, nil
}

func (ls *LabelService) Create(s *xorm.Session, l *models.Label, a web.Auth) (*models.Label, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}
	if err := l.Create(s, u); err != nil {
		return nil, err
	}
	return l, nil
}

func (ls *LabelService) Update(s *xorm.Session, l *models.Label, a web.Auth) (*models.Label, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}

	can, err := l.CanUpdate(s, u)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, echo.ErrForbidden
	}

	if err := l.Update(s); err != nil {
		return nil, err
	}
	return l, nil
}

func (ls *LabelService) Delete(s *xorm.Session, labelID int64, a web.Auth) error {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	l := &models.Label{ID: labelID}
	can, err := l.CanDelete(s, u)
	if err != nil {
		return err
	}
	if !can {
		return echo.ErrForbidden
	}

	return l.Delete(s)
}
