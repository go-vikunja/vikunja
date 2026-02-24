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
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// TaskChain represents a chain template that defines a sequence of tasks
// with relative timing offsets. When instantiated, all tasks in the chain
// are created with calculated dates and linked via precedes/follows relations.
type TaskChain struct {
	// The unique, numeric id of this chain.
	ID int64 `xorm:"autoincr not null unique pk" json:"id" param:"chain"`
	// The chain name.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// A description of what this chain workflow does.
	Description string `xorm:"longtext null" json:"description"`
	// The steps in this chain, ordered by sequence.
	Steps []*TaskChainStep `xorm:"-" json:"steps"`
	// The user who owns this chain.
	OwnerID int64      `xorm:"bigint not null INDEX" json:"-"`
	Owner   *user.User `xorm:"-" json:"owner" valid:"-"`
	// Timestamps
	Created time.Time `xorm:"created not null" json:"created"`
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TaskChainStep represents a single step in a chain template.
type TaskChainStep struct {
	// The unique id of this step.
	ID int64 `xorm:"autoincr not null unique pk" json:"id"`
	// The chain this step belongs to.
	ChainID int64 `xorm:"bigint not null INDEX" json:"chain_id"`
	// The position/order of this step in the chain (0-based).
	Sequence int `xorm:"int not null default 0" json:"sequence"`
	// The task title for this step.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)"`
	// Optional description for this step's task.
	Description string `xorm:"longtext null" json:"description"`
	// Offset in days from the anchor task's start date.
	OffsetDays int `xorm:"int not null default 0" json:"offset_days"`
	// Duration of this step's task in days (used to set end_date = start_date + duration).
	DurationDays int `xorm:"int not null default 1" json:"duration_days"`
	// Task priority.
	Priority int64 `xorm:"bigint null" json:"priority"`
	// Task color in hex.
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|7)" maxLength:"7"`
	// Label IDs to apply to the created task.
	LabelIDs []int64 `xorm:"json null" json:"label_ids"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*TaskChain) TableName() string {
	return "task_chains"
}

func (*TaskChainStep) TableName() string {
	return "task_chain_steps"
}

// --- Permissions ---

func (tc *TaskChain) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	can, err := tc.canDoChain(s, a)
	return can, int(PermissionAdmin), err
}

func (tc *TaskChain) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	fresh := &TaskChain{ID: tc.ID}
	return fresh.canDoChain(s, a)
}

func (tc *TaskChain) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return tc.canDoChain(s, a)
}

func (tc *TaskChain) CanCreate(_ *xorm.Session, a web.Auth) (bool, error) {
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}
	return a.GetID() > 0, nil
}

func (tc *TaskChain) canDoChain(s *xorm.Session, a web.Auth) (bool, error) {
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}
	existing := &TaskChain{}
	exists, err := s.Where("id = ?", tc.ID).Get(existing)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, &ErrTaskChainDoesNotExist{ID: tc.ID}
	}
	if existing.OwnerID != a.GetID() {
		return false, nil
	}
	*tc = *existing
	return true, nil
}

// --- CRUD ---

func (tc *TaskChain) Create(s *xorm.Session, a web.Auth) (err error) {
	tc.OwnerID = a.GetID()
	tc.ID = 0

	_, err = s.Insert(tc)
	if err != nil {
		return err
	}

	// Insert steps
	for i, step := range tc.Steps {
		step.ChainID = tc.ID
		step.Sequence = i
		step.ID = 0
		if step.LabelIDs == nil {
			step.LabelIDs = []int64{}
		}
		if _, err := s.Insert(step); err != nil {
			return err
		}
	}

	return nil
}

func (tc *TaskChain) ReadOne(s *xorm.Session, _ web.Auth) error {
	existing := &TaskChain{}
	exists, err := s.Where("id = ?", tc.ID).Get(existing)
	if err != nil {
		return err
	}
	if !exists {
		return &ErrTaskChainDoesNotExist{ID: tc.ID}
	}
	*tc = *existing
	tc.Owner, _ = user.GetUserByID(s, tc.OwnerID)

	// Load steps
	tc.Steps = []*TaskChainStep{}
	err = s.Where("chain_id = ?", tc.ID).OrderBy("sequence asc").Find(&tc.Steps)
	return err
}

func (tc *TaskChain) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	if _, is := a.(*LinkSharing); is {
		return nil, 0, 0, nil
	}

	chains := []*TaskChain{}
	query := s.Where("owner_id = ?", a.GetID())
	if search != "" {
		query = query.And("title LIKE ?", "%"+search+"%")
	}
	if perPage > 0 {
		query = query.Limit(perPage, (page-1)*perPage)
	}
	err = query.OrderBy("created desc").Find(&chains)
	if err != nil {
		return nil, 0, 0, err
	}

	totalCount, err := s.Where("owner_id = ?", a.GetID()).Count(&TaskChain{})
	if err != nil {
		return nil, 0, 0, err
	}

	// Load steps and owner for each chain
	for _, c := range chains {
		c.Owner, _ = user.GetUserByID(s, c.OwnerID)
		c.Steps = []*TaskChainStep{}
		_ = s.Where("chain_id = ?", c.ID).OrderBy("sequence asc").Find(&c.Steps)
	}

	return chains, len(chains), totalCount, nil
}

func (tc *TaskChain) Update(s *xorm.Session, _ web.Auth) error {
	// Update chain metadata
	_, err := s.ID(tc.ID).Cols("title", "description").Update(tc)
	if err != nil {
		return err
	}

	// Replace all steps: delete old ones and insert new ones
	_, err = s.Where("chain_id = ?", tc.ID).Delete(&TaskChainStep{})
	if err != nil {
		return err
	}

	for i, step := range tc.Steps {
		step.ChainID = tc.ID
		step.Sequence = i
		step.ID = 0
		if step.LabelIDs == nil {
			step.LabelIDs = []int64{}
		}
		if _, err := s.Insert(step); err != nil {
			return err
		}
	}

	return nil
}

func (tc *TaskChain) Delete(s *xorm.Session, _ web.Auth) error {
	// Delete steps first
	_, err := s.Where("chain_id = ?", tc.ID).Delete(&TaskChainStep{})
	if err != nil {
		return err
	}
	_, err = s.ID(tc.ID).Delete(&TaskChain{})
	return err
}

// --- Errors ---

type ErrTaskChainDoesNotExist struct {
	ID int64
}

func IsErrTaskChainDoesNotExist(err error) bool {
	_, ok := err.(*ErrTaskChainDoesNotExist)
	return ok
}

func (err *ErrTaskChainDoesNotExist) Error() string {
	return fmt.Sprintf("Task chain does not exist [ID: %d]", err.ID)
}

func (err *ErrTaskChainDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: 404,
		Code:     12010,
		Message:  "This task chain does not exist.",
	}
}
