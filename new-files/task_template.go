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
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

type TaskTemplate struct {
	ID          int64          `xorm:"autoincr not null unique pk" json:"id" param:"template"`
	Title       string         `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	Description string         `xorm:"longtext null" json:"description"`
	Priority    int64          `xorm:"bigint null" json:"priority"`
	HexColor    string         `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|7)" maxLength:"7"`
	PercentDone float64        `xorm:"DOUBLE null" json:"percent_done"`
	RepeatAfter int64          `xorm:"bigint null" json:"repeat_after"`
	RepeatMode  TaskRepeatMode `xorm:"not null default 0" json:"repeat_mode"`
	LabelIDs    []int64        `xorm:"json null" json:"label_ids"`
	OwnerID     int64          `xorm:"bigint not null INDEX" json:"-"`
	Owner       *user.User     `xorm:"-" json:"owner" valid:"-"`
	Created     time.Time      `xorm:"created not null" json:"created"`
	Updated     time.Time      `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*TaskTemplate) TableName() string {
	return "task_templates"
}

func getTaskTemplateSimpleByID(s *xorm.Session, id int64) (tt *TaskTemplate, err error) {
	tt = &TaskTemplate{}
	exists, err := s.Where("id = ?", id).Get(tt)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &ErrTaskTemplateDoesNotExist{ID: id}
	}
	return
}

func (tt *TaskTemplate) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	can, err := tt.canDoTemplate(s, a)
	return can, int(PermissionAdmin), err
}

func (tt *TaskTemplate) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	fresh := &TaskTemplate{ID: tt.ID}
	return fresh.canDoTemplate(s, a)
}

func (tt *TaskTemplate) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return tt.canDoTemplate(s, a)
}

func (tt *TaskTemplate) CanCreate(_ *xorm.Session, a web.Auth) (bool, error) {
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}
	return a.GetID() > 0, nil
}

func (tt *TaskTemplate) canDoTemplate(s *xorm.Session, a web.Auth) (bool, error) {
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}
	ttt, err := getTaskTemplateSimpleByID(s, tt.ID)
	if err != nil {
		return false, err
	}
	if ttt.OwnerID != a.GetID() {
		return false, nil
	}
	*tt = *ttt
	return true, nil
}

func (tt *TaskTemplate) Create(s *xorm.Session, a web.Auth) (err error) {
	tt.OwnerID = a.GetID()
	tt.ID = 0
	if tt.LabelIDs == nil {
		tt.LabelIDs = []int64{}
	}
	_, err = s.Insert(tt)
	return
}

func (tt *TaskTemplate) ReadOne(s *xorm.Session, _ web.Auth) error {
	ttt, err := getTaskTemplateSimpleByID(s, tt.ID)
	if err != nil {
		return err
	}
	*tt = *ttt
	tt.Owner, _ = user.GetUserByID(s, tt.OwnerID)
	return nil
}

func (tt *TaskTemplate) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	if _, is := a.(*LinkSharing); is {
		return nil, 0, 0, nil
	}

	templates := []*TaskTemplate{}
	query := s.Where("owner_id = ?", a.GetID())
	if search != "" {
		query = query.And("title LIKE ?", "%"+search+"%")
	}
	if perPage > 0 {
		query = query.Limit(perPage, (page-1)*perPage)
	}
	err = query.OrderBy("created desc").Find(&templates)
	if err != nil {
		return nil, 0, 0, err
	}

	totalCount, err := s.Where("owner_id = ?", a.GetID()).Count(&TaskTemplate{})
	if err != nil {
		return nil, 0, 0, err
	}

	for _, t := range templates {
		t.Owner, _ = user.GetUserByID(s, t.OwnerID)
	}

	return templates, len(templates), totalCount, nil
}

func (tt *TaskTemplate) Update(s *xorm.Session, _ web.Auth) error {
	_, err := s.ID(tt.ID).Cols(
		"title", "description", "priority", "hex_color",
		"percent_done", "repeat_after", "repeat_mode", "label_ids",
	).Update(tt)
	return err
}

func (tt *TaskTemplate) Delete(s *xorm.Session, _ web.Auth) error {
	_, err := s.ID(tt.ID).Delete(&TaskTemplate{})
	return err
}

type ErrTaskTemplateDoesNotExist struct {
	ID int64
}

func IsErrTaskTemplateDoesNotExist(err error) bool {
	_, ok := err.(*ErrTaskTemplateDoesNotExist)
	return ok
}

func (err *ErrTaskTemplateDoesNotExist) Error() string {
	return "Task template does not exist"
}

func (err *ErrTaskTemplateDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: 404,
		Code:     12001,
		Message:  "This task template does not exist.",
	}
}
