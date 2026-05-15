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

package models

import (
	"os"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetThreadID(t *testing.T) {
	// Save original config value
	originalPublicURL := config.ServicePublicURL.GetString()
	defer func() {
		config.ServicePublicURL.Set(originalPublicURL)
	}()

	t.Run("default domain when no public URL", func(t *testing.T) {
		config.ServicePublicURL.Set("")
		threadID := getThreadID(123)
		expectedDomain := "vikunja"
		if hostname, err := os.Hostname(); err == nil && hostname != "" {
			expectedDomain = hostname
		}
		assert.Equal(t, "<task-123@"+expectedDomain+">", threadID)
	})

	t.Run("simple domain without port", func(t *testing.T) {
		config.ServicePublicURL.Set("https://vikunja.example.com")
		threadID := getThreadID(456)
		assert.Equal(t, "<task-456@vikunja.example.com>", threadID)
	})

	t.Run("domain with standard HTTPS port", func(t *testing.T) {
		config.ServicePublicURL.Set("https://vikunja.example.com:443")
		threadID := getThreadID(789)
		// Should strip port to create valid RFC 5322 domain
		assert.Equal(t, "<task-789@vikunja.example.com>", threadID)
	})

	t.Run("domain with non-standard port", func(t *testing.T) {
		config.ServicePublicURL.Set("http://localhost:8080")
		threadID := getThreadID(999)
		// Should strip port to create valid RFC 5322 domain
		assert.Equal(t, "<task-999@localhost>", threadID)
	})

	t.Run("domain with port 3456", func(t *testing.T) {
		config.ServicePublicURL.Set("http://vikunja.local:3456")
		threadID := getThreadID(111)
		// Should strip port to create valid RFC 5322 domain
		assert.Equal(t, "<task-111@vikunja.local>", threadID)
	})

	t.Run("IP address with port", func(t *testing.T) {
		config.ServicePublicURL.Set("http://192.168.1.100:8080")
		threadID := getThreadID(222)
		// Should strip port to create valid RFC 5322 domain
		assert.Equal(t, "<task-222@192.168.1.100>", threadID)
	})

	t.Run("invalid URL falls back to default", func(t *testing.T) {
		config.ServicePublicURL.Set("not a valid url")
		threadID := getThreadID(333)
		expectedDomain := "vikunja"
		if hostname, err := os.Hostname(); err == nil && hostname != "" {
			expectedDomain = hostname
		}
		assert.Equal(t, "<task-333@"+expectedDomain+">", threadID)
	})

	t.Run("URL with path", func(t *testing.T) {
		config.ServicePublicURL.Set("https://example.com:9000/vikunja")
		threadID := getThreadID(444)
		// Should use hostname without port
		assert.Equal(t, "<task-444@example.com>", threadID)
	})
}

func TestUndoneTasksOverdueNotification_TitleIsMarkdownEscaped(t *testing.T) {
	originalPublicURL := config.ServicePublicURL.GetString()
	t.Cleanup(func() { config.ServicePublicURL.Set(originalPublicURL) })
	config.ServicePublicURL.Set("https://vikunja.example.com/")

	maliciousTitle := "bad](https://evil.com) [click here"
	n := &UndoneTasksOverdueNotification{
		User: &user.User{ID: 1, Name: "alice", Username: "alice"},
		Tasks: map[int64]*Task{
			42: {ID: 42, Title: maliciousTitle, ProjectID: 7, DueDate: time.Now().Add(-1 * time.Hour)},
		},
		Projects: map[int64]*Project{
			7: {ID: 7, Title: "My Project"},
		},
	}

	mail := n.ToMail("en")
	require.NotNil(t, mail)

	opts, err := notifications.RenderMail(mail, "en")
	require.NoError(t, err)

	// The rendered HTML must NOT contain an anchor pointing at evil.com —
	// that would be a successful injection. The literal string "evil.com"
	// may still appear inside the link text because the malicious title is
	// rendered verbatim; that is harmless as long as it is not an active
	// link/image.
	assert.NotContains(t, opts.HTMLMessage, `href="https://evil.com`,
		"malicious URL must not be rendered as an anchor")
	assert.NotContains(t, opts.HTMLMessage, `src="https://evil.com`,
		"malicious URL must not be rendered as an image")
	assert.Contains(t, opts.HTMLMessage, "https://vikunja.example.com/tasks/42",
		"legitimate task link must still render")
	// Exactly one anchor to the task — the injection must not create a
	// second one. Vikunja templates also render the Action link, but this
	// notification has no Action for the individual task.
	assert.Equal(t, 1, strings.Count(opts.HTMLMessage, `<a href="https://vikunja.example.com/tasks/42`),
		"expected exactly one anchor to the task")
	// The malicious title text must still be displayed as literal text
	// (goldmark will render the backslash escapes as the original characters).
	assert.Contains(t, opts.HTMLMessage, "bad](https://evil.com) [click here",
		"malicious title must render as literal text")
}

func TestTaskCommentNotification_ToTitle(t *testing.T) {
	doer := &user.User{ID: 1, Name: "alice", Username: "alice"}
	task := &Task{ID: 42, Title: "Take out trash", Index: 7}

	t.Run("regular comment", func(t *testing.T) {
		n := &TaskCommentNotification{Doer: doer, Task: task, Mentioned: false}
		title := n.ToTitle("en")
		assert.Contains(t, title, "Take out trash")
		assert.NotContains(t, title, "alice", "regular comment title should not mention the doer")
	})

	t.Run("mention switches title", func(t *testing.T) {
		n := &TaskCommentNotification{Doer: doer, Task: task, Mentioned: true}
		title := n.ToTitle("en")
		assert.Contains(t, title, "alice", "mentioned title should mention the doer")
		assert.Contains(t, title, "Take out trash")
	})

	t.Run("regular and mentioned produce different titles", func(t *testing.T) {
		regular := (&TaskCommentNotification{Doer: doer, Task: task, Mentioned: false}).ToTitle("en")
		mentioned := (&TaskCommentNotification{Doer: doer, Task: task, Mentioned: true}).ToTitle("en")
		assert.NotEqual(t, regular, mentioned)
	})
}

func TestTaskAssignedNotification_ToTitle(t *testing.T) {
	doer := &user.User{ID: 1, Name: "alice", Username: "alice"}
	assignee := &user.User{ID: 2, Name: "bob", Username: "bob"}
	third := &user.User{ID: 3, Name: "carol", Username: "carol"}
	task := &Task{ID: 42, Title: "Take out trash", Index: 7}

	t.Run("to assignee themself", func(t *testing.T) {
		n := &TaskAssignedNotification{Doer: doer, Task: task, Assignee: assignee, Target: assignee}
		title := n.ToTitle("en")
		assert.Contains(t, title, "Take out trash")
	})

	t.Run("doer assigned to themself", func(t *testing.T) {
		n := &TaskAssignedNotification{Doer: doer, Task: task, Assignee: doer, Target: third}
		title := n.ToTitle("en")
		assert.Contains(t, title, "alice")
	})

	t.Run("doer assigned someone else, target is third party", func(t *testing.T) {
		n := &TaskAssignedNotification{Doer: doer, Task: task, Assignee: assignee, Target: third}
		title := n.ToTitle("en")
		assert.Contains(t, title, "bob")
	})

	t.Run("three branches produce three distinct titles", func(t *testing.T) {
		a := (&TaskAssignedNotification{Doer: doer, Task: task, Assignee: assignee, Target: assignee}).ToTitle("en")
		b := (&TaskAssignedNotification{Doer: doer, Task: task, Assignee: doer, Target: third}).ToTitle("en")
		c := (&TaskAssignedNotification{Doer: doer, Task: task, Assignee: assignee, Target: third}).ToTitle("en")
		assert.NotEqual(t, a, b)
		assert.NotEqual(t, a, c)
		assert.NotEqual(t, b, c)
	})
}

func TestUserMentionedInTaskNotification_ToTitle(t *testing.T) {
	doer := &user.User{ID: 1, Name: "alice", Username: "alice"}
	task := &Task{ID: 42, Title: "Take out trash", Index: 7}

	t.Run("existing task", func(t *testing.T) {
		n := &UserMentionedInTaskNotification{Doer: doer, Task: task, IsNew: false}
		title := n.ToTitle("en")
		assert.Contains(t, title, "alice")
		assert.Contains(t, title, "Take out trash")
	})

	t.Run("new task", func(t *testing.T) {
		n := &UserMentionedInTaskNotification{Doer: doer, Task: task, IsNew: true}
		title := n.ToTitle("en")
		assert.Contains(t, title, "alice")
		assert.Contains(t, title, "Take out trash")
	})

	t.Run("new task and existing task produce different titles", func(t *testing.T) {
		existing := (&UserMentionedInTaskNotification{Doer: doer, Task: task, IsNew: false}).ToTitle("en")
		newOne := (&UserMentionedInTaskNotification{Doer: doer, Task: task, IsNew: true}).ToTitle("en")
		assert.NotEqual(t, existing, newOne)
	})
}

func TestReminderDueNotification_TitleIsMarkdownEscaped(t *testing.T) {
	originalPublicURL := config.ServicePublicURL.GetString()
	t.Cleanup(func() { config.ServicePublicURL.Set(originalPublicURL) })
	config.ServicePublicURL.Set("https://vikunja.example.com/")

	n := &ReminderDueNotification{
		User:    &user.User{ID: 1, Name: "alice"},
		Task:    &Task{ID: 99, Title: "![](https://evil.com/track.png)"},
		Project: &Project{ID: 1, Title: "proj"},
	}

	mail := n.ToMail("en")
	opts, err := notifications.RenderMail(mail, "en")
	require.NoError(t, err)

	// The injection must not create an <img> tag. The literal title
	// characters may still be present as escaped text.
	assert.NotContains(t, opts.HTMLMessage, "<img src=\"https://evil.com",
		"tracking pixel must not render as an <img> tag")
	assert.NotContains(t, opts.HTMLMessage, `href="https://evil.com`,
		"tracking URL must not render as an anchor")
}
