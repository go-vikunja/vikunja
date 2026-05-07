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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/output"
	"code.vikunja.io/veans/internal/status"
)

func newClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim <id>",
		Short: "Claim a task: assign the bot, move to In Progress, tag with branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := loadRuntime()
			if err != nil {
				return err
			}
			id, err := rt.resolveTaskID(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			// Move to In Progress.
			bid, err := status.BucketID(status.InProgress, rt.cfg.Buckets)
			if err != nil {
				return err
			}
			task, err := rt.client.UpdateTask(cmd.Context(), id, &client.Task{
				ID:       id,
				BucketID: bid,
				Done:     false,
			})
			if err != nil {
				return err
			}

			// Assign the bot. Idempotent on repeat — Vikunja returns 409 if
			// already assigned, which we map to a soft-skip.
			if err := rt.client.AddAssignee(cmd.Context(), id, rt.cfg.Bot.UserID); err != nil {
				var oe *output.Error
				if !errors.As(err, &oe) || oe.Code != output.CodeConflict {
					return err
				}
			}

			// Tag with the current branch label, if there is one.
			if branch := currentGitBranch(cmd.Context()); branch != "" {
				labelTitle := branchLabel(branch)
				l, err := getOrCreateLabelByTitle(cmd.Context(), rt.client, labelTitle)
				if err != nil {
					return err
				}
				if err := rt.client.AddLabelToTask(cmd.Context(), id, l.ID); err != nil {
					var oe *output.Error
					if !errors.As(err, &oe) || oe.Code != output.CodeConflict {
						return err
					}
				}
			}

			fresh, err := rt.client.GetTask(cmd.Context(), id)
			if err == nil {
				task = fresh
			}

			if globals.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(task)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Claimed %s  %s\n",
				rt.cfg.FormatTaskID(task.Index), task.Title)
			return nil
		},
	}
	return cmd
}
