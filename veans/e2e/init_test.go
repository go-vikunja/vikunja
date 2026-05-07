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

package e2e

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/veans/internal/config"
	"code.vikunja.io/veans/internal/credentials"
)

// TestInit_HappyPath exercises the full bootstrap: pick project + view,
// create canonical buckets, create the bot user, share the project, mint
// the bot's token, and write .veans.yml. Verifies side effects via the
// admin client.
func TestInit_HappyPath(t *testing.T) {
	h := New(t)

	// Use a unique title and identifier per run so parallel jobs don't
	// collide on the bot username.
	suffix := uniqueSuffix()
	project := h.CreateProject(t, "veans-e2e-"+suffix, identifier(suffix))
	view := h.FindKanbanView(t, project.ID)

	ws := h.NewWorkspace(t)
	// Rename the workspace dir so the auto-generated bot username is unique.
	ws.BotUsername = "bot-veans-e2e-" + suffix

	stdout, stderr, code := h.Run(t, ws,
		"init",
		"--server", h.APIURL(),
		"--token", h.AdminToken(),
		"--project", fmt.Sprintf("%d", project.ID),
		"--view", fmt.Sprintf("%d", view.ID),
		"--bot-username", ws.BotUsername,
		"--yes-buckets",
	)
	if code != 0 {
		t.Fatalf("init exit %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	// Config written?
	cfg, err := config.Load(ws.ConfigPath)
	if err != nil {
		t.Fatalf("load .veans.yml: %v", err)
	}
	if cfg.ProjectID != project.ID || cfg.ViewID != view.ID {
		t.Fatalf("unexpected ids in config: %+v", cfg)
	}
	if cfg.Bot.Username != ws.BotUsername {
		t.Fatalf("bot username = %q, want %q", cfg.Bot.Username, ws.BotUsername)
	}
	if cfg.Buckets.Todo == 0 || cfg.Buckets.InProgress == 0 || cfg.Buckets.InReview == 0 || cfg.Buckets.Done == 0 || cfg.Buckets.Scrapped == 0 {
		t.Fatalf("buckets not fully populated: %+v", cfg.Buckets)
	}

	// Bot token persisted in the file backend (since HOME points at a
	// fresh tmpdir, the file backend takes over from the missing keyring).
	store := credentials.NewFileBackend(ws.XDGConfig + "/veans/credentials.yml")
	tok, err := store.Get(h.APIURL(), ws.BotUsername)
	if err != nil {
		t.Fatalf("token not persisted: %v", err)
	}
	if !strings.HasPrefix(tok, "tk_") {
		t.Fatalf("bot token doesn't look like a Vikunja API token: %q", tok)
	}

	// Bot exists on the server with the right username.
	bots, err := h.AdminClient().ListBotUsers(context.Background())
	if err != nil {
		t.Fatalf("list bots: %v", err)
	}
	found := false
	for _, b := range bots {
		if b.Username == ws.BotUsername {
			found = true
			if b.ID != cfg.Bot.UserID {
				t.Fatalf("bot user_id mismatch: server=%d cfg=%d", b.ID, cfg.Bot.UserID)
			}
			break
		}
	}
	if !found {
		t.Fatalf("bot %q not found on server", ws.BotUsername)
	}

	// Project shared with the bot at write permission.
	var shares []map[string]any
	_ = h.AdminClient().Do(context.Background(), "GET", fmt.Sprintf("/projects/%d/users", project.ID), nil, nil, &shares)
	shareFound := false
	for _, s := range shares {
		if u, _ := s["username"].(string); u == ws.BotUsername {
			if p, _ := s["permission"].(float64); int(p) >= 1 {
				shareFound = true
				break
			}
		}
	}
	if !shareFound {
		t.Fatalf("project not shared with bot at write permission: %v", shares)
	}
}

func TestInit_NoIdentifierFallsBackToHashNN(t *testing.T) {
	h := New(t)

	suffix := uniqueSuffix()
	project := h.CreateProject(t, "veans-e2e-noid-"+suffix, "")
	view := h.FindKanbanView(t, project.ID)

	ws := h.NewWorkspace(t)
	ws.BotUsername = "bot-veans-noid-" + suffix

	_, stderr, code := h.Run(t, ws,
		"init",
		"--server", h.APIURL(),
		"--token", h.AdminToken(),
		"--project", fmt.Sprintf("%d", project.ID),
		"--view", fmt.Sprintf("%d", view.ID),
		"--bot-username", ws.BotUsername,
		"--yes-buckets",
	)
	if code != 0 {
		t.Fatalf("init exit %d\n%s", code, stderr)
	}

	cfg, err := config.Load(ws.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ProjectIdentifier != "" {
		t.Fatalf("expected empty identifier, got %q", cfg.ProjectIdentifier)
	}
	if got := cfg.FormatTaskID(7); got != "#7" {
		t.Fatalf("expected #7, got %q", got)
	}
}

// uniqueSuffix returns a short slug derived from the current nanosecond
// timestamp, base-36-encoded so every character is alphanumeric. Tests
// also use this slug as a project identifier, which Vikunja caps at 10
// chars, so the encoding has to be compact and free of separators.
func uniqueSuffix() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// identifier returns a stable 10-char-or-fewer slug for use as a Vikunja
// project identifier. The base-36 timestamp's most-significant chars
// barely change across consecutive runs, so we use the trailing chars
// (which carry the nanosecond entropy) and uppercase them.
func identifier(suffix string) string {
	if len(suffix) > 8 {
		suffix = suffix[len(suffix)-8:]
	}
	return strings.ToUpper(suffix)
}
