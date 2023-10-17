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
	"bytes"
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/version"
	"code.vikunja.io/web"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"
	"xorm.io/xorm"
)

type Webhook struct {
	ID        int64    `xorm:"bigint autoincr not null unique pk" json:"id" param:"webhook"`
	TargetURL string   `xorm:"not null" valid:"minstringlength(1)" minLength:"1" json:"target_url"`
	Events    []string `xorm:"JSON not null" valid:"minstringlength(1)" minLength:"1" json:"events"`
	ProjectID int64    `xorm:"bigint not null index" json:"project_id" param:"project"`
	Secret    string   `xorm:"null" json:"secret"`

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

var availableWebhookEvents map[string]bool
var availableWebhookEventsLock *sync.Mutex

func init() {
	availableWebhookEvents = make(map[string]bool)
	availableWebhookEventsLock = &sync.Mutex{}
}

func RegisterEventForWebhook(event events.Event) {
	if !config.WebhooksEnabled.GetBool() {
		return
	}

	availableWebhookEventsLock.Lock()
	defer availableWebhookEventsLock.Unlock()

	availableWebhookEvents[event.Name()] = true
	events.RegisterListener(event.Name(), &WebhookListener{
		EventName: event.Name(),
	})
}

func GetAvailableWebhookEvents() []string {
	evts := []string{}
	for e := range availableWebhookEvents {
		evts = append(evts, e)
	}

	sort.Strings(evts)

	return evts
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

func getWebhookHTTPClient() (client *http.Client) {
	client = http.DefaultClient
	client.Timeout = time.Duration(config.WebhooksTimeoutSeconds.GetInt()) * time.Second

	if config.WebhooksProxyURL.GetString() == "" || config.WebhooksProxyPassword.GetString() == "" {
		return
	}

	proxyURL, _ := url.Parse(config.WebhooksProxyURL.GetString())

	client.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		ProxyConnectHeader: http.Header{
			"Proxy-Authorization": []string{"Basic " + base64.StdEncoding.EncodeToString([]byte("vikunja:"+config.WebhooksProxyPassword.GetString()))},
			"User-Agent":          []string{"Vikunja/" + version.Version},
		},
	}

	return
}

func (w *Webhook) sendWebhookPayload(p *WebhookPayload) (err error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, w.TargetURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	if len(w.Secret) > 0 {
		sig256 := hmac.New(sha256.New, []byte(w.Secret))
		_, err = sig256.Write(payload)
		if err != nil {
			log.Errorf("Could not generate webhook signature for Webhook %d: %s", w.ID, err)
		}
		signature := hex.EncodeToString(sig256.Sum(nil))
		req.Header.Add("X-Vikunja-Signature", signature)
	}

	req.Header.Add("User-Agent", "Vikunja/"+version.Version)

	if config.WebhooksProxyURL.GetString() != "" && config.WebhooksProxyPassword.GetString() != "" {
		req.Header.Add("Proxy-Authorization", base64.StdEncoding.EncodeToString([]byte(config.WebhooksProxyPassword.GetString())))
	}

	client := getWebhookHTTPClient()
	_, err = client.Do(req)
	if err == nil {
		log.Debugf("Sent webhook payload for webhook %d for event %s", w.ID, p.EventName)
	}
	return
}
