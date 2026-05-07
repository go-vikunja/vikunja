package e2e

import (
	"os/exec"
	"strings"
	"testing"

	"code.vikunja.io/veans/internal/config"
)

// provisionWorkspace runs `veans init` against a fresh project and returns
// the workspace + harness primed for command-level e2e tests. Each test that
// needs a working .veans.yml calls this at the top.
func provisionWorkspace(t *testing.T) (*Workspace, *Harness) {
	t.Helper()
	h := New(t)
	suffix := uniqueSuffix()
	project := h.CreateProject(t, "veans-e2e-"+suffix, "VE"+strings.ToUpper(suffix[:4]))
	view := h.FindKanbanView(t, project.ID)

	ws := h.NewWorkspace(t)
	ws.BotUsername = "bot-veans-e2e-" + suffix

	_, stderr, code := h.Run(t, ws,
		"init",
		"--server", h.APIURL(),
		"--token", h.AdminToken(),
		"--project", iToS(project.ID),
		"--view", iToS(view.ID),
		"--bot-username", ws.BotUsername,
		"--yes-buckets",
	)
	if code != 0 {
		t.Fatalf("provision init failed: %s", stderr)
	}
	return ws, h
}

// loadConfig reads .veans.yml out of a workspace.
func loadConfig(t *testing.T, ws *Workspace) *config.Config {
	t.Helper()
	c, err := config.Load(ws.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// gitInWorkspace runs git inside the workspace and fails the test on error.
func gitInWorkspace(t *testing.T, ws *Workspace, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = ws.Dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func iToS(n int64) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	negative := false
	if n < 0 {
		negative = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = digits[n%10]
		n /= 10
	}
	if negative {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
