package websocket

const (
	// Client actions
	ActionAuth        = "auth"
	ActionSubscribe   = "subscribe"
	ActionUnsubscribe = "unsubscribe"

	// Server actions
	ActionAuthSuccess  = "auth.success"
	ActionUnsubscribed = "unsubscribed"
)

// IncomingMessage represents a message from the client.
type IncomingMessage struct {
	Action string `json:"action"`
	// Token is set for auth action.
	Token string `json:"token,omitempty"`
	// Topic is set for subscribe/unsubscribe actions.
	Topic string `json:"topic,omitempty"`
}

// OutgoingMessage represents a message from the server to the client.
// Exactly one of Event, Error, or Action will be set.
type OutgoingMessage struct {
	// Event is set for push events (e.g. "notification.created").
	Event string `json:"event,omitempty"`
	// Error is set for error responses (e.g. "forbidden").
	Error string `json:"error,omitempty"`
	// Action is set for server-initiated actions (e.g. "auth.success", "unsubscribed").
	Action string `json:"action,omitempty"`
	// Success is set for auth.success action.
	Success bool `json:"success,omitempty"`
	// Reason provides context for server-initiated actions.
	Reason string `json:"reason,omitempty"`
	// Topic identifies the subscription topic (omitted for auth responses).
	Topic string `json:"topic,omitempty"`
	// Data carries the event payload.
	Data interface{} `json:"data,omitempty"`
}
