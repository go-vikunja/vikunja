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

package apiv2

import (
	"context"
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

// Element type is the foreign notifications.DatabaseNotification because that's
// what models.DatabaseNotifications.ReadAll returns directly, not a models wrapper.
type notificationListBody struct {
	Body Paginated[*notifications.DatabaseNotification]
}

// markAllReadBody is the mark-all-as-read confirmation. The action has no
// resource to return, so it carries a short status message rather than reusing
// emptyBody (204) — mirrors v1's {"message":"success"} response.
type markAllReadBody struct {
	Body struct {
		Message string `json:"message" readOnly:"true" doc:"A confirmation message."`
	}
}

// RegisterNotificationRoutes wires notification list / mark-read / mark-all onto the Huma API.
func RegisterNotificationRoutes(api huma.API) {
	tags := []string{"notifications"}

	Register(api, huma.Operation{
		OperationID: "notifications-list",
		Summary:     "List notifications",
		Description: "Returns the authenticated user's own notifications, newest first. Link shares have no notifications and are refused.",
		Method:      http.MethodGet,
		Path:        "/notifications",
		Tags:        tags,
	}, notificationsList)

	Register(api, huma.Operation{
		OperationID: "notifications-mark-read",
		Summary:     "Mark a notification as (un-)read",
		Description: "Marks one of the authenticated user's notifications as read or unread. A user can only mark their own notifications.",
		Method:      http.MethodPut,
		Path:        "/notifications/{notificationid}",
		Tags:        tags,
	}, notificationsMarkRead)

	Register(api, huma.Operation{
		OperationID: "notifications-mark-all-read",
		Summary:     "Mark all notifications as read",
		Description: "Marks every notification of the authenticated user as read. Link shares have no notifications and are refused.",
		Method:      http.MethodPost,
		Path:        "/notifications",
		Tags:        tags,
	}, notificationsMarkAllRead)
}

func init() { AddRouteRegistrar(RegisterNotificationRoutes) }

func notificationsList(ctx context.Context, in *ListParams) (*notificationListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.DatabaseNotifications{}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*notifications.DatabaseNotification)
	if !ok {
		return nil, fmt.Errorf("notifications.ReadAll returned unexpected type %T (expected []*notifications.DatabaseNotification)", result)
	}
	return &notificationListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func notificationsMarkRead(ctx context.Context, in *struct {
	ID   int64 `path:"notificationid"`
	Body models.DatabaseNotifications
}) (*singleBody[models.DatabaseNotifications], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	n := &in.Body
	n.ID = in.ID // URL wins over body
	if err := handler.DoUpdate(ctx, n, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.DatabaseNotifications]{Body: n}, nil
}

// notificationsMarkAllRead is a custom action: there is no CRUDable Do* for a
// bulk mark, so the handler owns the link-share guard, the session and the
// commit. Mirrors apiv1.MarkAllNotificationsAsRead.
func notificationsMarkAllRead(ctx context.Context, _ *struct{}) (*markAllReadBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if _, is := a.(*models.LinkSharing); is {
		return nil, huma.Error403Forbidden("link shares cannot have notifications")
	}

	s := db.NewSession()
	defer s.Close()

	if err := notifications.MarkAllNotificationsAsRead(s, a.GetID()); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	out := &markAllReadBody{}
	out.Body.Message = "success"
	return out, nil
}
