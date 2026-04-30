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
