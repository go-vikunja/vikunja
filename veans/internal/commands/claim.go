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

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/status"
)

func newClaimCmd() *cobra.Command {
	var (
		fieldsRaw string
		taskIDArg int64
	)
	cmd := &cobra.Command{
		Use:   "claim <PROJ-NN|#NN|NN> | claim --id <task id>",
		Short: "Claim a task: assign the bot, move to In Progress, tag with branch",
		RunE: func(cmd *cobra.Command, args []string) error {
			fields, err := parseFields(fieldsRaw)
			if err != nil {
				return err
			}
			rt, err := loadRuntime(cmd.Context())
			if err != nil {
				return err
			}
			id, err := rt.taskID(cmd.Context(), args, taskIDArg)
			if err != nil {
				return err
			}
			task, err := runClaim(cmd.Context(), rt, id)
			if err != nil {
				return err
			}
			return rt.printTask(cmd.OutOrStdout(), fields, task)
		},
	}
	cmd.Args = addTaskIDFlag(cmd, &taskIDArg)
	addFieldsFlag(cmd, &fieldsRaw)
	return cmd
}

func runClaim(ctx context.Context, rt *runtime, id int64) (*client.Task, error) {
	current, err := rt.client.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}
	progress := &steps{task: rt.cfg.FormatTaskID(current.Index)}

	bid, err := status.BucketID(status.InProgress, rt.cfg.Buckets)
	if err != nil {
		return nil, err
	}
	if err := rt.client.MoveTaskToBucket(ctx, rt.cfg.ProjectID, rt.cfg.ViewID, bid, id); err != nil {
		return nil, progress.fail("set bucket", err)
	}
	progress.done("set bucket")

	// Checked up front: v2 answers a duplicate assignee with a plain 400.
	if !taskAssignedTo(current, rt.cfg.Bot.UserID) {
		if err := rt.client.AddAssignee(ctx, id, rt.cfg.Bot.UserID); err != nil {
			return nil, progress.fail("assign", err)
		}
		progress.done("assign")
	}

	if branch := currentGitBranch(ctx); branch != "" {
		title := branchLabel(branch)
		if !taskHasLabel(current, title) {
			step := "label " + title
			l, err := getOrCreateLabelByTitle(ctx, rt.client, title)
			if err != nil {
				return nil, progress.fail(step, err)
			}
			if err := rt.client.AddLabelToTask(ctx, id, l.ID); err != nil {
				return nil, progress.fail(step, err)
			}
			progress.done(step)
		}
	}

	return rt.client.GetTask(ctx, id)
}
