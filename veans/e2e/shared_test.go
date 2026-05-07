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
	cmd := exec.CommandContext(t.Context(), "git", args...)
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
