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

package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	petname "github.com/dustinkirkland/golang-petname"

	"code.vikunja.io/veans/internal/auth"
	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/output"
)

// resolveBotUser settles the bot identity for `veans init`:
//
//  1. If a bot we already own with this username exists, ask whether to
//     reuse it. Reuse skips creation; the rest of init continues with
//     the existing bot's ID.
//  2. If the username is taken by someone else, propose a petname-based
//     alternative (e.g. "bot-clever-otter") and loop on rejection.
//  3. Otherwise, create the bot fresh.
//
// The flow is best-effort transparent: in non-interactive contexts
// (--bot-username collides with someone else's bot and no TTY), we
// surface a clear CONFLICT error pointing at --bot-username.
func resolveBotUser(ctx context.Context, c *client.Client, username, projectTitle string, p auth.Prompter, w io.Writer) (*client.BotUser, error) {
	for {
		// Step 1 + 2: see if anyone is using this name.
		mine, err := c.FindMyBotByUsername(ctx, username)
		if err != nil {
			return nil, output.Wrap(output.CodeUnknown, err, "look up existing bots: %v", err)
		}
		if mine != nil {
			ok, err := confirmReuse(p, w, mine.Username)
			if err != nil {
				return nil, err
			}
			if ok {
				progress(w, "Reusing existing bot user %q (id=%d)", mine.Username, mine.ID)
				return mine, nil
			}
			// User declined; fall through to prompt for a new name.
			username, err = promptForReplacementName(p, w, username, false)
			if err != nil {
				return nil, err
			}
			continue
		}

		// Step 3: try creating.
		bot, err := c.CreateBotUser(ctx, username, "veans bot for "+projectTitle)
		if err == nil {
			progress(w, "Created bot user %q (id=%d)", bot.Username, bot.ID)
			return bot, nil
		}

		// On "username already exists" we know it's owned by someone
		// other than us (we just checked FindMyBotByUsername). Anything
		// else is a real failure — surface it.
		var oe *output.Error
		if !errors.As(err, &oe) || !isUsernameTakenErr(oe) {
			return nil, err
		}
		username, err = promptForReplacementName(p, w, username, true)
		if err != nil {
			return nil, err
		}
	}
}

// confirmReuse asks whether to reuse a bot user this caller already owns.
// Default is yes — re-running init in a worktree that's already onboarded
// is the common path.
func confirmReuse(p auth.Prompter, w io.Writer, username string) (bool, error) {
	fmt.Fprintf(w, "Bot user %q already exists and is owned by you.\n", username)
	ans, err := p.ReadLine("Reuse it? [Y/n]: ")
	if err != nil {
		return false, output.Wrap(output.CodeUnknown, err, "read confirmation: %v", err)
	}
	switch strings.ToLower(strings.TrimSpace(ans)) {
	case "", "y", "yes":
		return true, nil
	}
	return false, nil
}

// promptForReplacementName asks for an alternate bot username, suggesting
// a petname-based default. ownedByOther=true means the previous name
// collided with someone else's bot; we phrase the prompt accordingly.
func promptForReplacementName(p auth.Prompter, w io.Writer, previous string, ownedByOther bool) (string, error) {
	suggested := suggestPetname()
	if ownedByOther {
		fmt.Fprintf(w, "Bot username %q is taken by another user.\n", previous)
	} else {
		fmt.Fprintln(w, "Pick a different bot username.")
	}
	fmt.Fprintf(w, "Suggestion: %s\n", suggested)
	ans, err := p.ReadLine(fmt.Sprintf("New bot username [%s]: ", suggested))
	if err != nil {
		return "", output.Wrap(output.CodeUnknown, err, "read username: %v", err)
	}
	name := strings.TrimSpace(ans)
	if name == "" {
		name = suggested
	}
	if !strings.HasPrefix(name, "bot-") {
		name = "bot-" + name
	}
	if name == previous {
		return "", output.New(output.CodeValidation, "new bot username must differ from %q", previous)
	}
	return name, nil
}

// suggestPetname proposes a memorable bot- name like "bot-clever-otter".
// Two words keeps the username short enough for Vikunja's 250-char limit
// while still giving plenty of namespace.
func suggestPetname() string {
	return "bot-" + petname.Generate(2, "-")
}

// isUsernameTakenErr returns true when the wrapped HTTP error from
// CreateBotUser indicates a username collision. Vikunja replies 400 with
// the canonical "user with this username already exists" message.
func isUsernameTakenErr(e *output.Error) bool {
	if e == nil {
		return false
	}
	if e.Code != output.CodeValidation {
		return false
	}
	msg := strings.ToLower(e.Message)
	return strings.Contains(msg, "username already exists") ||
		strings.Contains(msg, "user with this username")
}
