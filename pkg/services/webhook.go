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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// WebhookService handles webhook-related operations
type WebhookService struct {
	DB       *xorm.Engine
	Registry *ServiceRegistry
}

// NewWebhookService creates a new WebhookService
// Deprecated: Use ServiceRegistry.Webhook() instead.
func NewWebhookService(db *xorm.Engine) *WebhookService {
	registry := NewServiceRegistry(db)
	return registry.Webhook()
}

// InitWebhookService initializes the webhook service dependency injection.
func InitWebhookService() {
	models.RegisterWebhookService(&webhookServiceDelegator{})
}

// webhookServiceDelegator implements WebhookServiceProvider for dependency injection
type webhookServiceDelegator struct{}

func (d *webhookServiceDelegator) CanRead(s *xorm.Session, projectID int64, a web.Auth) (bool, int, error) {
	ws := NewWebhookService(s.Engine())
	return ws.CanRead(s, projectID, a)
}

func (d *webhookServiceDelegator) CanCreate(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	ws := NewWebhookService(s.Engine())
	return ws.CanCreate(s, projectID, a)
}

func (d *webhookServiceDelegator) CanUpdate(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	ws := NewWebhookService(s.Engine())
	return ws.CanUpdate(s, projectID, a)
}

func (d *webhookServiceDelegator) CanDelete(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	ws := NewWebhookService(s.Engine())
	return ws.CanDelete(s, projectID, a)
}

// CanRead checks if a user can read webhooks for a project
func (ws *WebhookService) CanRead(s *xorm.Session, projectID int64, a web.Auth) (bool, int, error) {
	p := &models.Project{ID: projectID}
	return p.CanRead(s, a)
}

// CanCreate checks if a user can create a webhook for a project
func (ws *WebhookService) CanCreate(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	return ws.canDoWebhook(s, projectID, a)
}

// CanUpdate checks if a user can update a webhook
func (ws *WebhookService) CanUpdate(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	return ws.canDoWebhook(s, projectID, a)
}

// CanDelete checks if a user can delete a webhook
func (ws *WebhookService) CanDelete(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	return ws.canDoWebhook(s, projectID, a)
}

// canDoWebhook checks if a user has admin permissions for the webhook's project
func (ws *WebhookService) canDoWebhook(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	// Link shares can't manage webhooks
	if _, isShareAuth := a.(*models.LinkSharing); isShareAuth {
		return false, nil
	}

	// User must have update permission on the project
	p := &models.Project{ID: projectID}
	return p.CanUpdate(s, a)
}
