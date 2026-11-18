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
	"fmt"

	"code.vikunja.io/veans/internal/output"
)

// steps tracks a multi-step mutation so a failure partway through reports
// which task was already changed and what got applied.
type steps struct {
	task      string
	completed []string
}

func (s *steps) done(step string) { s.completed = append(s.completed, step) }

func (s *steps) fail(step string, err error) error {
	inner := output.AsError(err)
	msg := fmt.Sprintf("%s: %s failed: %s", s.task, step, inner.Message)
	if len(s.completed) > 0 {
		msg = fmt.Sprintf("%s: %s failed after %v succeeded: %s", s.task, step, s.completed, inner.Message)
	}
	if inner.Code == output.CodeAuth {
		// Earlier steps used the same token, so this is a missing scope, not a bad token.
		msg += " (the bot token lacks the permission for this call; run `veans login` to mint one with the current scopes)"
	}
	return &output.Error{
		Code:      inner.Code,
		Message:   msg,
		Task:      s.task,
		Completed: append([]string(nil), s.completed...),
		Failed:    step,
		Cause:     err,
	}
}
