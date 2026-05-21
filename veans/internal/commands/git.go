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

package commands

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

// runGit runs `git <args...>` in the current working directory and returns
// trimmed stdout. Errors are returned to the caller so they can decide
// whether silence or escalation is appropriate.
//
// The inherited environment is scrubbed of all GIT_* variables before
// invocation. Defense-in-depth: a stray GIT_DIR / GIT_WORK_TREE /
// GIT_INDEX_FILE in the caller's environment could redirect git to a
// different repository and cause downstream commands (e.g. `claim`
// attaching `veans:branch:<name>`) to act on the wrong branch.
// GIT_OPTIONAL_LOCKS=0 is set so a concurrent git process holding the
// index lock can't block veans.
func runGit(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(scrubGitEnv(os.Environ()), "GIT_OPTIONAL_LOCKS=0")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

// scrubGitEnv returns env entries whose keys do not start with "GIT_".
// PATH and other essentials are preserved so git can still be located
// and configured normally (e.g. SSH_AUTH_SOCK, HOME, USER).
func scrubGitEnv(env []string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		if strings.HasPrefix(kv, "GIT_") {
			continue
		}
		out = append(out, kv)
	}
	return out
}
