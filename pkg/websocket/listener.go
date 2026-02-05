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
