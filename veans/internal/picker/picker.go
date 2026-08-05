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

package picker

import (
	"errors"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"code.vikunja.io/veans/internal/client"
	"golang.org/x/term"
)

// Result is what the user chose: an existing project or the create-new action.
type Result struct {
	Project   *client.Project
	CreateNew bool
}

var (
	// ErrCanceled is returned when the user dismisses the picker (Esc / Ctrl-C).
	ErrCanceled = errors.New("selection canceled")
	// ErrNotATerminal is returned when stdin is not a TTY, so the interactive
	// picker can't run — callers should fall back to `--project <id>`.
	ErrNotATerminal = errors.New("not a terminal")
)

// Pick runs the interactive project picker over projects and returns the
// user's choice. Output is written to stderr (prompts go to stderr by
// convention) and the terminal is left in canonical mode on exit.
func Pick(projects []*client.Project) (Result, error) {
	// The picker reads stdin and draws to stderr; both must be a TTY, else it
	// would run invisibly (e.g. stderr redirected to a file) and look hung.
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stderr.Fd())) {
		return Result{}, ErrNotATerminal
	}

	m := newModel(buildForest(projects))
	prog := tea.NewProgram(m, tea.WithInput(os.Stdin), tea.WithOutput(os.Stderr))
	final, err := prog.Run()
	if err != nil {
		return Result{}, fmt.Errorf("run project picker: %w", err)
	}

	fm, ok := final.(*model)
	if !ok {
		return Result{}, fmt.Errorf("project picker returned unexpected model type %T", final)
	}
	if fm.canceled {
		return Result{}, ErrCanceled
	}
	if fm.createNew {
		return Result{CreateNew: true}, nil
	}
	return Result{Project: fm.result}, nil
}
