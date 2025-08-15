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

package v2

import (
	"fmt"
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	v2 "code.vikunja.io/api/pkg/models/v2"
	"github.com/labstack/echo/v4"
)

func getProjectFromContext(c echo.Context) (*models.Project, error) {
	s := db.NewSession()
	defer s.Close()

	p, err := models.GetProjectByID(s, c.Get("project_id").(int64))
	if err != nil {
		return nil, err
	}
	return p, nil
}

// GetWebhooks handles getting all webhooks for a project.
func GetWebhooks(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	p, err := getProjectFromContext(c)
	if err != nil {
		return err
	}

	webhooks, err := models.GetWebhooksByProject(s, p.ID)
	if err != nil {
		return err
	}

	v2Webhooks := make([]*v2.Webhook, len(webhooks))
	for i, w := range webhooks {
		v2Webhooks[i] = &v2.Webhook{
			Webhook: *w,
			Links: &v2.WebhookLinks{
				Self: &v2.Link{
					Href: fmt.Sprintf("/api/v2/projects/%d/webhooks/%d", p.ID, w.ID),
				},
			},
		}
	}

	return c.JSON(http.StatusOK, v2Webhooks)
}

// CreateWebhook handles creating a new webhook.
func CreateWebhook(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	p, err := getProjectFromContext(c)
	if err != nil {
		return err
	}

	var w models.Webhook
	if err := c.Bind(&w); err != nil {
		return err
	}
	w.ProjectID = p.ID

	if err := w.Create(s); err != nil {
		return err
	}

	v2Webhook := &v2.Webhook{
		Webhook: w,
		Links: &v2.WebhookLinks{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/projects/%d/webhooks/%d", p.ID, w.ID),
			},
		},
	}

	return c.JSON(http.StatusCreated, v2Webhook)
}

// GetWebhook handles getting a webhook by its ID.
func GetWebhook(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	webhookID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid webhook ID.")
	}

	w := &models.Webhook{ID: webhookID}
	if err := w.ReadOne(s); err != nil {
		return err
	}

	v2Webhook := &v2.Webhook{
		Webhook: *w,
		Links: &v2.WebhookLinks{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/projects/%d/webhooks/%d", w.ProjectID, w.ID),
			},
		},
	}

	return c.JSON(http.StatusOK, v2Webhook)
}

// UpdateWebhook handles updating a webhook.
func UpdateWebhook(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	webhookID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid webhook ID.")
	}

	var w models.Webhook
	if err := c.Bind(&w); err != nil {
		return err
	}
	w.ID = webhookID

	if err := w.Update(s); err != nil {
		return err
	}

	v2Webhook := &v2.Webhook{
		Webhook: w,
		Links: &v2.WebhookLinks{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/projects/%d/webhooks/%d", w.ProjectID, w.ID),
			},
		},
	}

	return c.JSON(http.StatusOK, v2Webhook)
}

// DeleteWebhook handles deleting a webhook.
func DeleteWebhook(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	webhookID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid webhook ID.")
	}

	w := &models.Webhook{ID: webhookID}
	if err := w.Delete(s); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
