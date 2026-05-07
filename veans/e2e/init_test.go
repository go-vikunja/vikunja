package e2e

import (
	"context"
	"fmt"
	"os"
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
	project := h.CreateProject(t, "veans-e2e-"+suffix, "VE"+strings.ToUpper(suffix[:4]))
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

// uniqueSuffix returns a short random-ish slug for naming test artifacts.
// Time-based is fine here — tests don't need cryptographic uniqueness.
func uniqueSuffix() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "host"
	}
	return strings.ToLower(fmt.Sprintf("%s-%d", trunc(hostname, 4), time.Now().UnixNano()))[:18]
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
