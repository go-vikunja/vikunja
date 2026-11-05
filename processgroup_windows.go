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

//go:build mage && windows

package main

import (
	"os/exec"
	"strconv"
	"syscall"
)

const createNewProcessGroup = 0x00000200

// setProcessGroup configures a command to run in its own process group,
// so that Ctrl+C / termination signals can be handled as a group.
// Windows has no POSIX process groups; CREATE_NEW_PROCESS_GROUP is the
// closest equivalent and lets a Ctrl+C be delivered to the new console group.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNewProcessGroup}
}

// killProcessGroup terminates the whole process tree on Windows via
// taskkill /T /F and then waits for the parent process.
func killProcessGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	// Best-effort tree kill: taskkill is available on all supported Windows
	// versions and fails silently when the tree is already gone.
	tk := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	tk.Stdout = cmd.Stdout
	tk.Stderr = cmd.Stderr
	_ = tk.Run()
	return cmd.Wait()
}
