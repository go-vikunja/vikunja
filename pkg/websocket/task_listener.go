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
	"cmp"
	"encoding/json"
	"slices"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"

	"github.com/ThreeDotsLabs/watermill/message"
	"xorm.io/xorm"
)

// TaskEventListener pushes task lifecycle events to subscribed WebSocket
// connections. wsEvent is "task.created", "task.updated" or "task.deleted".
//
// Unlike the notification and timer listeners, task events are interesting to
// every subscribed user with access to the task's project, not just one user,
// so the hub fans the event out with a per-connection project access check.
type TaskEventListener struct {
	wsEvent string
}

func (l *TaskEventListener) Name() string { return "websocket.push." + l.wsEvent }

// All task events share the {task, doer} shape; only the task itself is needed
// to scope and deliver the event.
func (l *TaskEventListener) Handle(msg *message.Message) error {
	var event struct {
		Task *models.Task `json:"task"`
	}
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}
	if event.Task == nil {
		return nil
	}

	hub := GetHub()
	if hub == nil {
		log.Warningf("WebSocket: hub not initialized, skipping %s push", l.wsEvent)
		return nil
	}

	s := db.NewSession()
	defer s.Close()

	hub.PublishTaskEvent(s, l.wsEvent, event.Task)
	return nil
}

// MentionedUser is the mention metadata carried on comment events. It is a
// trimmed copy of the user so clients can render mentions without an extra
// API round-trip.
type MentionedUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// CommentEventPayload is the data of task.comment.* events.
type CommentEventPayload struct {
	Comment  *models.TaskComment `json:"comment"`
	Mentions []MentionedUser     `json:"mentions"`
}

// TaskCommentEventListener pushes task comment events to subscribed WebSocket
// connections. wsEvent is "task.comment.created", "task.comment.edited" or
// "task.comment.deleted"; comment events carry the users mentioned in the
// comment text as metadata.
type TaskCommentEventListener struct {
	wsEvent string
}

func (l *TaskCommentEventListener) Name() string { return "websocket.push." + l.wsEvent }

// The comment event's Task is required: TaskComment.TaskID is not part of the
// serialized event (json:"-"), so the task's project is only known through it.
func (l *TaskCommentEventListener) Handle(msg *message.Message) error {
	var event struct {
		Task    *models.Task        `json:"task"`
		Comment *models.TaskComment `json:"comment"`
	}
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}
	if event.Comment == nil || event.Task == nil {
		return nil
	}

	hub := GetHub()
	if hub == nil {
		log.Warningf("WebSocket: hub not initialized, skipping %s push", l.wsEvent)
		return nil
	}

	s := db.NewSession()
	defer s.Close()

	mentions, err := mentionedUsers(s, event.Comment.Comment)
	if err != nil {
		// A failure to resolve mentions must not drop the event itself.
		log.Errorf("WebSocket: resolving mentions for comment %d failed: %v", event.Comment.ID, err)
	}

	hub.PublishCommentEvent(s, l.wsEvent, event.Task, &CommentEventPayload{
		Comment:  event.Comment,
		Mentions: mentions,
	})
	return nil
}

// mentionedUsers resolves the users mentioned in a comment's HTML with the
// same mechanism the notification pipeline uses, returning them in a stable
// (id-sorted) order.
func mentionedUsers(s *xorm.Session, commentHTML string) ([]MentionedUser, error) {
	users, err := models.FindMentionedUsersInText(s, commentHTML)
	if err != nil {
		return nil, err
	}

	mentions := make([]MentionedUser, 0, len(users))
	for _, u := range users {
		mentions = append(mentions, MentionedUser{ID: u.ID, Username: u.Username, Name: u.Name})
	}
	slices.SortFunc(mentions, func(a, b MentionedUser) int { return cmp.Compare(a.ID, b.ID) })
	return mentions, nil
}
