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
	"xorm.io/xorm"
)

type LabelService struct {
	DB *xorm.Engine
}

func NewLabelService(db *xorm.Engine) *LabelService {
	return &LabelService{DB: db}
}

func (ls *LabelService) Create(s *xorm.Session, label *models.Label, u *user.User) error {
	if u == nil {
		return ErrAccessDenied
	}
	label.CreatedByID = u.ID
	_, err := s.Insert(label)
	return err
}

func (ls *LabelService) Get(s *xorm.Session, id int64, u *user.User) (*models.Label, error) {
	label := new(models.Label)
	has, err := s.ID(id).Get(label)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrLabelNotFound
	}

	can, err := ls.Can(s, label, u).Read()
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrAccessDenied
	}

	return label, nil
}

type LabelPermissions struct {
	s     *xorm.Session
	label *models.Label
	user  *user.User
}

func (ls *LabelService) Can(s *xorm.Session, label *models.Label, u *user.User) *LabelPermissions {
	return &LabelPermissions{s: s, label: label, user: u}
}

func (lp *LabelPermissions) Read() (bool, error) {
	if lp.user == nil {
		return false, nil
	}
	return lp.label.CreatedByID == lp.user.ID, nil
}
