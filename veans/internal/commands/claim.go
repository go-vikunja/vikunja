package commands

import (
	"encoding/json"
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
				if oe, ok := err.(*output.Error); !ok || oe.Code != output.CodeConflict {
					return err
				}
			}

			// Tag with the current branch label, if there is one.
			if branch := currentGitBranch(); branch != "" {
				labelTitle := branchLabel(branch)
				l, err := getOrCreateLabelByTitle(cmd.Context(), rt.client, labelTitle)
				if err != nil {
					return err
				}
				if err := rt.client.AddLabelToTask(cmd.Context(), id, l.ID); err != nil {
					if oe, ok := err.(*output.Error); !ok || oe.Code != output.CodeConflict {
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
