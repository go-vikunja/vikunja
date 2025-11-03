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
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/version"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

var webhookClient *http.Client

type Webhook struct {
	// The generated ID of this webhook target
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"webhook"`
	// The target URL where the POST request with the webhook payload will be made
	TargetURL string `xorm:"not null" valid:"required,url" json:"target_url"`
	// The webhook events which should fire this webhook target
	Events []string `xorm:"JSON not null" valid:"required" json:"events"`
	// The project ID of the project this webhook target belongs to
	ProjectID int64 `xorm:"bigint not null index" json:"project_id" param:"project"`
	// If provided, webhook requests will be signed using HMAC. Check out the docs about how to use this: https://vikunja.io/docs/webhooks/#signing
	Secret string `xorm:"null" json:"secret"`

	// The user who initially created the webhook target.
	CreatedBy   *user.User `xorm:"-" json:"created_by" valid:"-"`
	CreatedByID int64      `xorm:"bigint not null" json:"-"`

	// A timestamp when this webhook target was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this webhook target was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
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

// Create creates a webhook target
// @Summary Create a webhook target
// @Description Create a webhook target which receives POST requests about specified events from a project.
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Project ID"
// @Param webhook body models.Webhook true "The webhook target object with required fields"
// @Success 200 {object} models.Webhook "The created webhook target."
// @Failure 400 {object} web.HTTPError "Invalid webhook object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id}/webhooks [put]
func (w *Webhook) Create(s *xorm.Session, a web.Auth) (err error) {

	if !strings.HasPrefix(w.TargetURL, "http") {
		return InvalidFieldError([]string{"target_url"})
	}

	for _, event := range w.Events {
		if _, has := availableWebhookEvents[event]; !has {
			return InvalidFieldError([]string{"events"})
		}
	}

	w.CreatedByID = a.GetID()
	w.ID = 0
	_, err = s.Insert(w)
	if err != nil {
		return err
	}

	w.CreatedBy, err = user.GetUserByID(s, a.GetID())
	return
}

// ReadAll returns all webhook targets for a project
// @Summary Get all api webhook targets for the specified project
// @Description Get all api webhook targets for the specified project.
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per bucket per page. This parameter is limited by the configured maximum of items per page."
// @Param id path int true "Project ID"
// @Success 200 {array} models.Webhook "The list of all webhook targets"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /projects/{id}/webhooks [get]
func (w *Webhook) ReadAll(s *xorm.Session, a web.Auth, _ string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
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

	userIDs := []int64{}
	for _, webhook := range ws {
		userIDs = append(userIDs, webhook.CreatedByID)
	}

	users, err := user.GetUsersByIDs(s, userIDs)
	if err != nil {
		return nil, 0, 0, err
	}

	for _, webhook := range ws {
		webhook.Secret = ""
		if createdBy, has := users[webhook.CreatedByID]; has {
			webhook.CreatedBy = createdBy
		}
	}

	return ws, len(ws), total, err
}

// Update updates a webhook target
// @Summary Change a webhook target's events.
// @Description Change a webhook target's events. You cannot change other values of a webhook.
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Project ID"
// @Param webhookID path int true "Webhook ID"
// @Success 200 {object} models.Webhook "Updated webhook target"
// @Failure 404 {object} web.HTTPError "The webhok target does not exist"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id}/webhooks/{webhookID} [post]
func (w *Webhook) Update(s *xorm.Session, _ web.Auth) (err error) {
	for _, event := range w.Events {
		if _, has := availableWebhookEvents[event]; !has {
			return InvalidFieldError([]string{"events"})
		}
	}

	_, err = s.Where("id = ?", w.ID).
		Cols("events").
		Update(w)
	return
}

// Delete deletes a webhook target
// @Summary Deletes an existing webhook target
// @Description Delete any of the project's webhook targets.
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Project ID"
// @Param webhookID path int true "Webhook ID"
// @Success 200 {object} models.Message "Successfully deleted."
// @Failure 404 {object} web.HTTPError "The webhok target does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id}/webhooks/{webhookID} [delete]
func (w *Webhook) Delete(s *xorm.Session, _ web.Auth) (err error) {
	_, err = s.Where("id = ?", w.ID).Delete(&Webhook{})
	return
}

func getWebhookHTTPClient() (client *http.Client) {

	if webhookClient != nil {
		return webhookClient
	}

	client = http.DefaultClient
	client.Timeout = time.Duration(config.WebhooksTimeoutSeconds.GetInt()) * time.Second

	if config.WebhooksProxyURL.GetString() == "" || config.WebhooksProxyPassword.GetString() == "" {
		webhookClient = client
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

	webhookClient = client

	return
}

func (w *Webhook) sendWebhookPayload(p *WebhookPayload) (err error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, w.TargetURL, bytes.NewReader(payload))
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
	req.Header.Add("Content-Type", "application/json")

	client := getWebhookHTTPClient()
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode > 399 {
		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		log.Errorf("Got response with status %d from webhook %d: %s", res.StatusCode, w.ID, responseBody)
	}

	log.Debugf("Sent webhook payload for webhook %d for event %s", w.ID, p.EventName)
	return
}
