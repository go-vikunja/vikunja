package notifications

import "code.vikunja.io/api/pkg/events"

// NotificationCreatedEvent is dispatched when a notification is saved to the database.
type NotificationCreatedEvent struct {
	Notification *DatabaseNotification `json:"notification"`
	UserID       int64                 `json:"user_id"`
}

// Name returns the event name.
func (n *NotificationCreatedEvent) Name() string {
	return "notification.created"
}

var _ events.Event = (*NotificationCreatedEvent)(nil)
