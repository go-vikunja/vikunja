package websocket

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncomingAuthMessageDeserialization(t *testing.T) {
	raw := `{"action":"auth","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}`
	var msg IncomingMessage
	err := json.Unmarshal([]byte(raw), &msg)
	require.NoError(t, err)
	assert.Equal(t, ActionAuth, msg.Action)
	assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", msg.Token)
}

func TestIncomingSubscribeMessageDeserialization(t *testing.T) {
	raw := `{"action":"subscribe","topic":"notifications"}`
	var msg IncomingMessage
	err := json.Unmarshal([]byte(raw), &msg)
	require.NoError(t, err)
	assert.Equal(t, ActionSubscribe, msg.Action)
	assert.Equal(t, "notifications", msg.Topic)
}

func TestOutgoingEventSerialization(t *testing.T) {
	msg := OutgoingMessage{
		Event: "notification.created",
		Topic: "notifications",
		Data:  map[string]string{"hello": "world"},
	}
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"event":"notification.created"`)
	assert.Contains(t, string(data), `"topic":"notifications"`)
	assert.Contains(t, string(data), `"hello":"world"`)
}

func TestOutgoingErrorSerialization(t *testing.T) {
	msg := OutgoingMessage{
		Error: "forbidden",
		Topic: "project.tasks",
	}
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"error":"forbidden"`)
}

func TestOutgoingAuthSuccessSerialization(t *testing.T) {
	msg := OutgoingMessage{
		Action:  ActionAuthSuccess,
		Success: true,
	}
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"action":"auth.success"`)
	assert.Contains(t, string(data), `"success":true`)
}
