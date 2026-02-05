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
