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

package websocket

import (
	"encoding/json"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"

	"github.com/ThreeDotsLabs/watermill/message"
)

// NotificationListener pushes new notifications to WebSocket clients.
type NotificationListener struct{}

// Name returns the listener name.
func (n *NotificationListener) Name() string {
	return "websocket.notification.push"
}

// Handle processes a notification created event and pushes it to the relevant WebSocket connections.
func (n *NotificationListener) Handle(msg *message.Message) error {
	var event notifications.NotificationCreatedEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}

	hub := GetHub()
	if hub == nil {
		log.Warningf("WebSocket: hub not initialized, skipping notification push")
		return nil
	}

	hub.PublishForUser(event.UserID, "notifications", "notification.created", event.Notification)
	return nil
}

// RegisterListeners registers WebSocket event listeners.
func RegisterListeners() {
	RegisterNotificationListener()
}

// RegisterNotificationListener registers the WebSocket notification listener.
func RegisterNotificationListener() {
	events.RegisterListener(
		(&notifications.NotificationCreatedEvent{}).Name(),
		&NotificationListener{},
	)
}
