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

package auth

// osOpen launches the OS's default browser at the given URL. The file is
// kept platform-neutral by delegating to "xdg-open" — the same approach
// most Go CLIs take. macOS, Linux and BSD all ship it (or a compat alias);
// on Windows the ./browser_windows.go shim takes precedence via build tag.

import (
	"context"
	"os/exec"
	"runtime"
	"time"
)

func osOpen(ctx context.Context, url string) error {
	// Cap the launch attempt so a misbehaving xdg-open shim can't block
	// the OAuth flow indefinitely. The browser process itself runs
	// independently of this child and survives the timeout.
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(cctx, "open", url)
	case "windows":
		cmd = exec.CommandContext(cctx, "rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.CommandContext(cctx, "xdg-open", url)
	}
	return cmd.Start()
}
