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
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/config"
)

//go:embed prompt.tmpl
var promptTemplate string

// primeContext is the data passed into the agent prompt template.
type primeContext struct {
	Server            string
	ProjectID         int64
	ProjectTitle      string
	ProjectIdentifier string
	ViewID            int64
	Buckets           config.Buckets
	BotUsername       string
	TaskIDExample     string
}

func newPrimeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prime",
		Short: "Emit the agent system prompt for this project",
		Long: `Renders the embedded prompt template against this repo's .veans.yml and
prints it to stdout. Designed to be wired into Claude Code's SessionStart
and PreCompact hooks (or the OpenCode equivalent) so coding agents always
have an up-to-date Vikunja cheat sheet in context.

If no .veans.yml is found upward from the current directory, prime exits
silently with status 0 — that makes the hook safe to install globally.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := config.Find("")
			if err != nil {
				if errors.Is(err, config.ErrNotFound) {
					return nil // silent — globally-installed hook safety
				}
				return err
			}
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}

			// Fetch the project title for nicer prompt copy. Best-effort —
			// if the API call fails (network blip, expired token), we fall
			// back to "(unknown)" rather than aborting the prompt render.
			projectTitle := "(unknown)"
			if rt, err := loadRuntime(); err == nil {
				if p, err := rt.client.GetProject(cmd.Context(), cfg.ProjectID); err == nil {
					projectTitle = p.Title
				}
			}

			data := primeContext{
				Server:            cfg.Server,
				ProjectID:         cfg.ProjectID,
				ProjectTitle:      projectTitle,
				ProjectIdentifier: cfg.ProjectIdentifier,
				ViewID:            cfg.ViewID,
				Buckets:           cfg.Buckets,
				BotUsername:       cfg.Bot.Username,
				TaskIDExample:     cfg.FormatTaskID(1),
			}

			tpl, err := template.New("prime").Parse(promptTemplate)
			if err != nil {
				return fmt.Errorf("parse prompt template: %w", err)
			}
			return tpl.Execute(cmd.OutOrStdout(), data)
		},
	}
}

// silence linter noise on unused symbols when wiring hooks.
var _ = client.New
var _ = strings.TrimSpace
