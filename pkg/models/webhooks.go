// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
	"sync"
	"time"
	"xorm.io/xorm"
)

type Webhook struct {
	ID        int64    `xorm:"bigint autoincr not null unique pk" json:"id" param:"webhook"`
	TargetURL string   `xorm:"not null" valid:"minstringlength(1)" minLength:"1" json:"target_url"`
	Events    []string `xorm:"JSON not null" valid:"minstringlength(1)" minLength:"1" json:"event"`
	ProjectID int64    `xorm:"not null" json:"project_id" param:"project"`

	// The user who initially created the webhook target.
	CreatedBy   *user.User `xorm:"-" json:"created_by" valid:"-"`
	CreatedByID int64      `xorm:"bigint not null" json:"-"`

	// A timestamp when this webhook target was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this webhook target was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

func (w *Webhook) TableName() string {
	return "webhooks"
}

type WebhookEvent interface {
	events.Event
	ProjectID() int64
}

var availableWebhookEvents map[string]bool
var availableWebhookEventsLock *sync.Mutex

func init() {
	availableWebhookEvents = make(map[string]bool)
	availableWebhookEventsLock = &sync.Mutex{}
}

func RegisterEventForWebhook(event WebhookEvent) {
	availableWebhookEventsLock.Lock()
	defer availableWebhookEventsLock.Unlock()

	availableWebhookEvents[event.Name()] = true
	events.RegisterListener(event.Name(), &WebhookListener{
		EventName: event.Name(),
	})
}

func (w *Webhook) Create(s *xorm.Session, a web.Auth) (err error) {
	// TODO: check valid webhook events
	w.CreatedByID = a.GetID()
	_, err = s.Insert(w)
	return
}

func (w *Webhook) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	p := &Project{ID: w.ProjectID}
	can, _, err := p.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	ws := []*Webhook{}
	err = s.Where("project_id = ?", w.ProjectID).
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&ws)
	if err != nil {
		return
	}

	total, err := s.Where("project_id = ?", w.ProjectID).
		Count(&Webhook{})
	if err != nil {
		return
	}

	return ws, len(ws), total, err
}

func (w *Webhook) Update(s *xorm.Session, a web.Auth) (err error) {
	// TODO validate webhook events
	_, err = s.Where("id = ?", w.ID).
		Cols("events").
		Update(w)
	return
}

func (w *Webhook) Delete(s *xorm.Session, a web.Auth) (err error) {
	_, err = s.Where("id = ?", w.ID).Delete(&Webhook{})
	return
}
